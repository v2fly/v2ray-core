package trojan

import (
	"encoding/binary"
	"io"

	"v2ray.com/core/common/buf"
	"v2ray.com/core/common/net"
	"v2ray.com/core/common/protocol"
)

var (
	crlf = []byte{'\r', '\n'}

	addrParser = protocol.NewAddressParser(
		protocol.AddressFamilyByte(0x01, net.AddressFamilyIPv4),
		protocol.AddressFamilyByte(0x04, net.AddressFamilyIPv6),
		protocol.AddressFamilyByte(0x03, net.AddressFamilyDomain),
	)
)

const (
	maxLength = 8192

	CommandTCP byte = 1
	CommandUDP byte = 3
)

type PacketReader struct {
	io.Reader
}

func (pr *PacketReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	_, mb, err := ReadPacket(pr.Reader)
	return mb, err
}

func ReadPacket(r io.Reader) (*net.Destination, buf.MultiBuffer, error) {
	addr, port, err := addrParser.ReadAddressPort(nil, r)
	if err != nil {
		return nil, nil, newError("failed to read address and port").Base(err)
	}

	var lengthBuf [2]byte
	if _, err := io.ReadFull(r, lengthBuf[:]); err != nil {
		return nil, nil, newError("failed to read payload length").Base(err)
	}

	remain := int(binary.BigEndian.Uint16(lengthBuf[:]))
	if remain > maxLength {
		return nil, nil, newError("oversize payload")
	}

	var crlf [2]byte
	if _, err := io.ReadFull(r, crlf[:]); err != nil {
		return nil, nil, newError("failed to read crlf").Base(err)
	}

	dest := net.UDPDestination(addr, port)
	var mb buf.MultiBuffer
	for remain > 0 {
		length := buf.Size
		if remain < length {
			length = remain
		}

		b := buf.New()
		mb = append(mb, b)
		n, err := b.ReadFullFrom(r, int32(length))
		if err != nil {
			return &dest, mb, newError("failed to read payload").Base(err)
		}

		remain -= int(n)
	}

	return &dest, mb, nil
}

type PacketWriter struct {
	io.Writer
	Target net.Destination
}

func (pw *PacketWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	b := make([]byte, maxLength)
	for !mb.IsEmpty() {
		var length int
		mb, length = buf.SplitBytes(mb, b)
		if _, err := WritePacket(pw.Writer, pw.Target, b[:length]); err != nil {
			buf.ReleaseMulti(mb)
			return newError("failed to write packet").Base(err)
		}
	}

	return nil
}

func ReadHeader(r io.Reader) (*net.Destination, error) {
	var crlf [2]byte
	var command [1]byte
	var hash [56]byte
	if _, err := io.ReadFull(r, hash[:]); err != nil {
		return nil, newError("failed to read user hash").Base(err)
	}

	if _, err := io.ReadFull(r, crlf[:]); err != nil {
		return nil, newError("failed to read crlf").Base(err)
	}

	if _, err := io.ReadFull(r, command[:]); err != nil {
		return nil, newError("failed to read command").Base(err)
	}

	network := net.Network_TCP
	if command[0] == CommandUDP {
		network = net.Network_UDP
	}

	addr, port, err := addrParser.ReadAddressPort(nil, r)
	if err != nil {
		return nil, newError("failed to read address and port").Base(err)
	}

	if _, err := io.ReadFull(r, crlf[:]); err != nil {
		return nil, newError("failed to read crlf").Base(err)
	}

	return &net.Destination{Address: addr, Port: port, Network: network}, nil
}

func WriteHeader(w io.Writer, target net.Destination, account *MemoryAccount) (buf.Writer, error) {
	buffer := buf.StackNew()
	defer buffer.Release()

	command := CommandTCP
	if target.Network == net.Network_UDP {
		command = CommandUDP
	}

	buffer.Write(account.Key)
	buffer.Write(crlf)
	buffer.WriteByte(command)
	addrParser.WriteAddressPort(&buffer, target.Address, target.Port)
	buffer.Write(crlf)

	_, err := w.Write(buffer.Bytes())
	if err != nil {
		return nil, err
	}

	var writer buf.Writer
	if target.Network == net.Network_UDP {
		writer = &PacketWriter{
			Writer: w,
			Target: target,
		}
	} else {
		writer = buf.NewWriter(w)
	}

	return writer, err
}

func WritePacket(w io.Writer, target net.Destination, payload []byte) (int, error) {
	buffer := buf.StackNew()
	defer buffer.Release()

	length := len(payload)
	lengthBuf := [2]byte{}
	binary.BigEndian.PutUint16(lengthBuf[:], uint16(length))
	addrParser.WriteAddressPort(&buffer, target.Address, target.Port)
	buffer.Write(lengthBuf[:])
	buffer.Write(crlf)
	buffer.Write(payload)
	_, err := w.Write(buffer.Bytes())
	if err != nil {
		return 0, err
	}

	return length, nil
}
