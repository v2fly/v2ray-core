package ntp

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

// The LeapIndicator is used to warn if a leap second should be inserted
// or deleted in the last minute of the current month.
type LeapIndicator uint8

const (
	// LeapNoWarning indicates no impending leap second.
	LeapNoWarning LeapIndicator = 0

	// LeapAddSecond indicates the last minute of the day has 61 seconds.
	LeapAddSecond = 1

	// LeapDelSecond indicates the last minute of the day has 59 seconds.
	LeapDelSecond = 2

	// LeapNotInSync indicates an unsynchronized leap second.
	LeapNotInSync = 3
)

// Internal constants
const (
	nanoPerSec      = 1000000000
	maxStratum      = 16
	maxPollInterval = (1 << 17) * time.Second
	maxDispersion   = 16 * time.Second
)

// Internal variables
var (
	ntpEpoch = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
)

type Mode uint8

// NTP modes. This package uses only Client Mode.
const (
	Reserved Mode = 0 + iota
	SymmetricActive
	SymmetricPassive
	Client
	Server
	Broadcast
	ControlMessage
	ReservedPrivate
)

// Time is a 64-bit fixed-point (Q32.32) representation of the number of
// seconds elapsed.
type Time uint64

// Duration interprets the fixed-point NTPTime as a number of elapsed seconds
// and returns the corresponding time.Duration value.
func (t Time) Duration() time.Duration {
	sec := (t >> 32) * nanoPerSec
	frac := (t & 0xffffffff) * nanoPerSec
	nsec := frac >> 32
	if uint32(frac) >= 0x80000000 {
		nsec++
	}
	return time.Duration(sec + nsec)
}

// Time interprets the fixed-point NTPTime as an absolute time and returns
// the corresponding time.Time value.
func (t Time) Time() time.Time {
	return ntpEpoch.Add(t.Duration())
}

// ToNtpTime converts the time.Time value t into its 64-bit fixed-point
// NTPTime representation.
func ToNtpTime(t time.Time) Time {
	nsec := uint64(t.Sub(ntpEpoch))
	sec := nsec / nanoPerSec
	nsec = (nsec - sec*nanoPerSec) << 32
	frac := nsec / nanoPerSec
	if nsec%nanoPerSec >= nanoPerSec/2 {
		frac++
	}
	return Time(sec<<32 | frac)
}

// An TimeShort is a 32-bit fixed-point (Q16.16) representation of the
// number of seconds elapsed.
type TimeShort uint32

// Duration interprets the fixed-point NTPTimeShort as a number of elapsed
// seconds and returns the corresponding time.Duration value.
func (t TimeShort) Duration() time.Duration {
	sec := uint64(t>>16) * nanoPerSec
	frac := uint64(t&0xffff) * nanoPerSec
	nsec := frac >> 16
	if uint16(frac) >= 0x8000 {
		nsec++
	}
	return time.Duration(sec + nsec)
}

// Message is an internal representation of an NTP packet.
type Message struct {
	LiVnMode       uint8 // Leap Indicator (2) + Version (3) + Mode (3)
	Stratum        uint8
	Poll           int8
	Precision      int8
	RootDelay      TimeShort
	RootDispersion TimeShort
	ReferenceID    uint32
	ReferenceTime  Time
	OriginTime     Time
	ReceiveTime    Time
	TransmitTime   Time
}

// SetVersion sets the NTP protocol version on the message.
func (m *Message) SetVersion(v int) {
	m.LiVnMode = (m.LiVnMode & 0xc7) | uint8(v)<<3
}

// SetMode sets the NTP protocol Mode on the message.
func (m *Message) SetMode(md Mode) {
	m.LiVnMode = (m.LiVnMode & 0xf8) | uint8(md)
}

// SetLeap modifies the leap indicator on the message.
func (m *Message) SetLeap(li LeapIndicator) {
	m.LiVnMode = (m.LiVnMode & 0x3f) | uint8(li)<<6
}

// GetVersion returns the version value in the message.
func (m *Message) GetVersion() int {
	return int((m.LiVnMode >> 3) & 0x07)
}

// GetMode returns the Mode value in the message.
func (m *Message) GetMode() Mode {
	return Mode(m.LiVnMode & 0x07)
}

// GetLeap returns the leap indicator on the message.
func (m *Message) GetLeap() LeapIndicator {
	return LeapIndicator((m.LiVnMode >> 6) & 0x03)
}

// A Response contains time data, some of which is returned by the NTP Server
// and some of which is calculated by the Client.
type Response struct {
	// Time is the transmit time reported by the Server just before it
	// responded to the Client's NTP query.
	Time time.Time

	// ClockOffset is the estimated offset of the Client clock relative to
	// the Server. Add this to the Client's system clock time to obtain a
	// more accurate time.
	ClockOffset time.Duration

	// RTT is the measured round-trip-time delay estimate between the Client
	// and the Server.
	RTT time.Duration

	// Precision is the reported precision of the Server's clock.
	Precision time.Duration

	// Stratum is the "stratum level" of the Server. The smaller the number,
	// the closer the Server is to the reference clock. Stratum 1 servers are
	// attached directly to the reference clock. A stratum value of 0
	// indicates the "kiss of death," which typically occurs when the Client
	// issues too many requests to the Server in a short period of time.
	Stratum uint8

	// ReferenceID is a 32-bit identifier identifying the Server or
	// reference clock.
	ReferenceID uint32

	// ReferenceTime is the time when the Server's system clock was last
	// set or corrected.
	ReferenceTime time.Time

	// RootDelay is the Server's estimated aggregate round-trip-time delay to
	// the stratum 1 Server.
	RootDelay time.Duration

	// RootDispersion is the Server's estimated maximum measurement error
	// relative to the stratum 1 Server.
	RootDispersion time.Duration

	// RootDistance is an estimate of the total synchronization distance
	// between the Client and the stratum 1 Server.
	RootDistance time.Duration

	// Leap indicates whether a leap second should be added or removed from
	// the current month's last minute.
	Leap LeapIndicator

	// MinError is a lower bound on the error between the Client and Server
	// clocks. When the Client and Server are not synchronized to the same
	// clock, the reported timestamps may appear to violate the principle of
	// causality. In other words, the NTP Server's response may indicate
	// that a message was received before it was sent. In such cases, the
	// minimum error may be useful.
	MinError time.Duration

	// KissCode is a 4-character string describing the reason for a
	// "kiss of death" response (stratum = 0). For a list of standard kiss
	// codes, see https://tools.ietf.org/html/rfc5905#section-7.4.
	KissCode string

	// Poll is the maximum interval between successive NTP polling messages.
	// It is not relevant for simple NTP clients like this one.
	Poll time.Duration
}

// Validate checks if the response is valid for the purposes of time
// synchronization.
func (r *Response) Validate() error {
	// Handle invalid stratum values.
	if r.Stratum == 0 {
		return fmt.Errorf("kiss of death received: %s", r.KissCode)
	}
	if r.Stratum >= maxStratum {
		return errors.New("invalid stratum in response")
	}

	// Handle invalid leap second indicator.
	if r.Leap == LeapNotInSync {
		return errors.New("invalid leap second")
	}

	// Estimate the "freshness" of the time. If it exceeds the maximum
	// polling interval (~36 hours), then it cannot be considered "fresh".
	freshness := r.Time.Sub(r.ReferenceTime)
	if freshness > maxPollInterval {
		return errors.New("server clock not fresh")
	}

	// Calculate the peer synchronization distance, lambda:
	//  	lambda := RootDelay/2 + RootDispersion
	// If this value exceeds MAXDISP (16s), then the time is not suitable
	// for synchronization purposes.
	// https://tools.ietf.org/html/rfc5905#appendix-A.5.1.1.
	lambda := r.RootDelay/2 + r.RootDispersion
	if lambda > maxDispersion {
		return errors.New("invalid dispersion")
	}

	// If the Server's transmit time is before its reference time, the
	// response is invalid.
	if r.Time.Before(r.ReferenceTime) {
		return errors.New("invalid time reported")
	}

	// nil means the response is valid.
	return nil
}

// Query performs the NTP Server query and returns the response message
// along with the local system time it was received.
func Query(conn net.Conn) (*Message, Time, error) {
	// Allocate a message to hold the response.
	recvMsg := new(Message)

	// Allocate a message to hold the query.
	xmitMsg := new(Message)
	xmitMsg.SetMode(Client)
	xmitMsg.SetVersion(4)
	xmitMsg.SetLeap(LeapNotInSync)

	// To ensure privacy and prevent spoofing, try to use a random 64-bit
	// value for the TransmitTime. If crypto/rand couldn't generate a
	// random value, fall back to using the system clock. Keep track of
	// when the messsage was actually transmitted.
	bits := make([]byte, 8)
	_, err := rand.Read(bits)
	var xmitTime time.Time
	if err == nil {
		xmitMsg.TransmitTime = Time(binary.BigEndian.Uint64(bits))
		xmitTime = time.Now()
	} else {
		xmitTime = time.Now()
		xmitMsg.TransmitTime = ToNtpTime(xmitTime)
	}

	// Transmit the query.
	err = binary.Write(conn, binary.BigEndian, xmitMsg)
	if err != nil {
		return nil, 0, err
	}

	// Receive the response.
	err = binary.Read(conn, binary.BigEndian, recvMsg)
	if err != nil {
		return nil, 0, err
	}

	// Keep track of the time the response was received.
	delta := time.Since(xmitTime)
	if delta < 0 {
		// The local system may have had its clock adjusted since it
		// sent the query. In go 1.9 and later, time.Since ensures
		// that a monotonic clock is used, so delta can never be less
		// than zero. In versions before 1.9, a monotonic clock is
		// not used, so we have to check.
		return nil, 0, errors.New("client clock ticked backwards")
	}
	recvTime := ToNtpTime(xmitTime.Add(delta))

	// Check for invalid fields.
	if recvMsg.GetMode() != Server {
		return nil, 0, errors.New("invalid Mode in response")
	}
	if recvMsg.TransmitTime == Time(0) {
		return nil, 0, errors.New("invalid transmit time in response")
	}
	if recvMsg.OriginTime != xmitMsg.TransmitTime {
		return nil, 0, errors.New("server response mismatch")
	}
	if recvMsg.ReceiveTime > recvMsg.TransmitTime {
		return nil, 0, errors.New("server clock ticked backwards")
	}

	// Correct the received message's origin time using the actual
	// transmit time.
	recvMsg.OriginTime = ToNtpTime(xmitTime)

	return recvMsg, recvTime, nil
}

// ParseTime parses the NTP packet along with the packet receive time to
// generate a Response record.
func ParseTime(m *Message, recvTime Time) *Response {
	r := &Response{
		Time:           m.TransmitTime.Time(),
		ClockOffset:    offset(m.OriginTime, m.ReceiveTime, m.TransmitTime, recvTime),
		RTT:            rtt(m.OriginTime, m.ReceiveTime, m.TransmitTime, recvTime),
		Precision:      toInterval(m.Precision),
		Stratum:        m.Stratum,
		ReferenceID:    m.ReferenceID,
		ReferenceTime:  m.ReferenceTime.Time(),
		RootDelay:      m.RootDelay.Duration(),
		RootDispersion: m.RootDispersion.Duration(),
		Leap:           m.GetLeap(),
		MinError:       minError(m.OriginTime, m.ReceiveTime, m.TransmitTime, recvTime),
		Poll:           toInterval(m.Poll),
	}

	// Calculate values depending on other calculated values
	r.RootDistance = rootDistance(r.RTT, r.RootDelay, r.RootDispersion)

	// If a kiss of death was received, interpret the reference ID as
	// a kiss code.
	if r.Stratum == 0 {
		r.KissCode = kissCode(r.ReferenceID)
	}

	return r
}

// The following helper functions calculate additional metadata about the
// timestamps received from an NTP Server.  The timestamps returned by
// the Server are given the following variable names:
//
//   org = Origin Timestamp (Client send time)
//   rec = Receive Timestamp (Server receive time)
//   xmt = Transmit Timestamp (Server reply time)
//   dst = Destination Timestamp (Client receive time)

func rtt(org, rec, xmt, dst Time) time.Duration {
	// round trip delay time
	//   rtt = (dst-org) - (xmt-rec)
	a := dst.Time().Sub(org.Time())
	b := xmt.Time().Sub(rec.Time())
	rtt := a - b
	if rtt < 0 {
		rtt = 0
	}
	return rtt
}

func offset(org, rec, xmt, dst Time) time.Duration {
	// local clock offset
	//   offset = ((rec-org) + (xmt-dst)) / 2
	a := rec.Time().Sub(org.Time())
	b := xmt.Time().Sub(dst.Time())
	return (a + b) / time.Duration(2)
}

func minError(org, rec, xmt, dst Time) time.Duration {
	// Each NTP response contains two pairs of send/receive timestamps.
	// When either pair indicates a "causality violation", we calculate the
	// error as the difference in time between them. The minimum error is
	// the greater of the two causality violations.
	var error0, error1 Time
	if org >= rec {
		error0 = org - rec
	}
	if xmt >= dst {
		error1 = xmt - dst
	}
	if error0 > error1 {
		return error0.Duration()
	}
	return error1.Duration()
}

func rootDistance(rtt, rootDelay, rootDisp time.Duration) time.Duration {
	// The root distance is:
	// 	the maximum error due to all causes of the local clock
	//	relative to the primary Server. It is defined as half the
	//	total delay plus total dispersion plus peer jitter.
	//	(https://tools.ietf.org/html/rfc5905#appendix-A.5.5.2)
	//
	// In the reference implementation, it is calculated as follows:
	//	rootDist = max(MINDISP, rootDelay + rtt)/2 + rootDisp
	//			+ peerDisp + PHI * (uptime - peerUptime)
	//			+ peerJitter
	// For an SNTP Client which sends only a single packet, most of these
	// terms are irrelevant and become 0.
	totalDelay := rtt + rootDelay
	return totalDelay/2 + rootDisp
}

func toInterval(t int8) time.Duration {
	switch {
	case t > 0:
		return time.Duration(uint64(time.Second) << uint(t))
	case t < 0:
		return time.Duration(uint64(time.Second) >> uint(-t))
	default:
		return time.Second
	}
}

func kissCode(id uint32) string {
	isPrintable := func(ch byte) bool { return ch >= 32 && ch <= 126 }

	b := []byte{
		byte(id >> 24),
		byte(id >> 16),
		byte(id >> 8),
		byte(id),
	}
	for _, ch := range b {
		if !isPrintable(ch) {
			return ""
		}
	}
	return string(b)
}
