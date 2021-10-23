//go:build !confonly
// +build !confonly

package ntp

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"strings"
	"time"

	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/protocol/ntp"
	"github.com/v2fly/v2ray-core/v4/features/routing"
)

var _ Server = (*ClassicNTPClient)(nil)

type ClassicNTPClient struct {
	ctx        context.Context
	name       string
	address    *net.Destination
	dispatcher routing.Dispatcher
}

func (c *ClassicNTPClient) QueryClockOffset() (time.Duration, error) {
	packet := new(ntp.Message)
	packet.SetVersion(4)
	packet.SetMode(ntp.Client)
	packet.SetLeap(ntp.LeapNotInSync)

	keyArray := make([]byte, 8)
	_, err := rand.Read(keyArray)
	var transmitTime time.Time
	if err == nil {
		packet.TransmitTime = ntp.Time(binary.BigEndian.Uint64(keyArray))
		transmitTime = time.Now()
	} else {
		transmitTime = time.Now()
		packet.TransmitTime = ntp.ToNtpTime(transmitTime)
	}

	link, err := c.dispatcher.Dispatch(c.ctx, *c.address)
	if err != nil {
		return 0, newError("failed to open udp connection").Base(err)
	}
	conn := net.NewConnection(
		net.ConnectionInputMulti(link.Writer),
		net.ConnectionOutputMulti(link.Reader),
	)
	err = binary.Write(conn, binary.BigEndian, packet)
	if err != nil {
		return 0, newError("failed to write ntp request").Base(err)
	}

	recvMsg := new(ntp.Message)
	err = binary.Read(conn, binary.BigEndian, recvMsg)
	if err != nil {
		return 0, newError("failed to read ntp response").Base(err)
	}

	delta := time.Since(transmitTime)
	if delta < 0 {
		// The local system may have had its clock adjusted since it
		// sent the query. In go 1.9 and later, time.Since ensures
		// that a monotonic clock is used, so delta can never be less
		// than zero. In versions before 1.9, a monotonic clock is
		// not used, so we have to check.
		return 0, newError("client clock ticked backwards")
	}
	recvTime := ntp.ToNtpTime(transmitTime.Add(delta))

	if recvMsg.GetMode() != ntp.Server {
		return 0, newError("invalid mode in recvMsg")
	}
	if recvMsg.TransmitTime == ntp.Time(0) {
		return 0, newError("invalid transmit time in recvMsg")
	}
	if recvMsg.OriginTime != packet.TransmitTime {
		return 0, newError("server recvMsg mismatch")
	}
	if recvMsg.ReceiveTime > recvMsg.TransmitTime {
		return 0, newError("server clock ticked backwards")
	}

	recvMsg.OriginTime = ntp.ToNtpTime(transmitTime)
	response := ntp.ParseTime(recvMsg, recvTime)

	return response.ClockOffset, nil
}

func NewClassicNTPClient(ctx context.Context, address net.Destination, dispatcher routing.Dispatcher) *ClassicNTPClient {
	// default to 123 if unspecific
	if address.Port == 0 {
		address.Port = net.Port(123)
	}
	s := &ClassicNTPClient{
		ctx:        ctx,
		address:    &address,
		name:       strings.ToUpper(address.String()),
		dispatcher: dispatcher,
	}
	newError("NTP: created UDP client initialized for ", address.NetAddr()).AtInfo().WriteToLog()
	return s
}
