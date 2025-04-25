package mux

import (
	"encoding/binary"
	"io"

	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/bitmask"
	"github.com/ghxhy/v2ray-core/v5/common/buf"
	"github.com/ghxhy/v2ray-core/v5/common/net"
	"github.com/ghxhy/v2ray-core/v5/common/protocol"
	"github.com/ghxhy/v2ray-core/v5/common/serial"
)

type SessionStatus byte

const (
	SessionStatusNew       SessionStatus = 0x01
	SessionStatusKeep      SessionStatus = 0x02
	SessionStatusEnd       SessionStatus = 0x03
	SessionStatusKeepAlive SessionStatus = 0x04
)

const (
	OptionData  bitmask.Byte = 0x01
	OptionError bitmask.Byte = 0x02
)

type TargetNetwork byte

const (
	TargetNetworkTCP TargetNetwork = 0x01
	TargetNetworkUDP TargetNetwork = 0x02
)

var addrParser = protocol.NewAddressParser(
	protocol.AddressFamilyByte(byte(protocol.AddressTypeIPv4), net.AddressFamilyIPv4),
	protocol.AddressFamilyByte(byte(protocol.AddressTypeDomain), net.AddressFamilyDomain),
	protocol.AddressFamilyByte(byte(protocol.AddressTypeIPv6), net.AddressFamilyIPv6),
	protocol.PortThenAddress(),
)

/*
Frame format
2 bytes - length
2 bytes - session id
1 bytes - status
1 bytes - option

1 byte - network
2 bytes - port
n bytes - address

*/

type FrameMetadata struct {
	Target        net.Destination
	SessionID     uint16
	Option        bitmask.Byte
	SessionStatus SessionStatus
}

func (f FrameMetadata) WriteTo(b *buf.Buffer) error {
	lenBytes := b.Extend(2)

	len0 := b.Len()
	sessionBytes := b.Extend(2)
	binary.BigEndian.PutUint16(sessionBytes, f.SessionID)

	common.Must(b.WriteByte(byte(f.SessionStatus)))
	common.Must(b.WriteByte(byte(f.Option)))

	if f.SessionStatus == SessionStatusNew {
		switch f.Target.Network {
		case net.Network_TCP:
			common.Must(b.WriteByte(byte(TargetNetworkTCP)))
		case net.Network_UDP:
			common.Must(b.WriteByte(byte(TargetNetworkUDP)))
		}

		if err := addrParser.WriteAddressPort(b, f.Target.Address, f.Target.Port); err != nil {
			return err
		}
	}

	len1 := b.Len()
	binary.BigEndian.PutUint16(lenBytes, uint16(len1-len0))
	return nil
}

// Unmarshal reads FrameMetadata from the given reader.
func (f *FrameMetadata) Unmarshal(reader io.Reader) error {
	metaLen, err := serial.ReadUint16(reader)
	if err != nil {
		return err
	}
	if metaLen > 512 {
		return newError("invalid metalen ", metaLen).AtError()
	}

	b := buf.New()
	defer b.Release()

	if _, err := b.ReadFullFrom(reader, int32(metaLen)); err != nil {
		return err
	}
	return f.UnmarshalFromBuffer(b)
}

// UnmarshalFromBuffer reads a FrameMetadata from the given buffer.
// Visible for testing only.
func (f *FrameMetadata) UnmarshalFromBuffer(b *buf.Buffer) error {
	if b.Len() < 4 {
		return newError("insufficient buffer: ", b.Len())
	}

	f.SessionID = binary.BigEndian.Uint16(b.BytesTo(2))
	f.SessionStatus = SessionStatus(b.Byte(2))
	f.Option = bitmask.Byte(b.Byte(3))
	f.Target.Network = net.Network_Unknown

	if f.SessionStatus == SessionStatusNew {
		if b.Len() < 8 {
			return newError("insufficient buffer: ", b.Len())
		}
		network := TargetNetwork(b.Byte(4))
		b.Advance(5)

		addr, port, err := addrParser.ReadAddressPort(nil, b)
		if err != nil {
			return newError("failed to parse address and port").Base(err)
		}

		switch network {
		case TargetNetworkTCP:
			f.Target = net.TCPDestination(addr, port)
		case TargetNetworkUDP:
			f.Target = net.UDPDestination(addr, port)
		default:
			return newError("unknown network type: ", network)
		}
	}

	return nil
}
