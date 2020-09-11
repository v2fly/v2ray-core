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
	buffer := buf.StackNew()
	defer buffer.Release()

	_, _, err := addrParser.ReadAddressPort(&buffer, pr.Reader)
	if err != nil {
		return nil, newError("failed to read address and port").Base(err)
	}

	buffer.Clear()
	if _, err := buffer.ReadFullFrom(pr.Reader, 2); err != nil {
		return nil, newError("failed to read payload length").Base(err)
	}

	remain := int(binary.BigEndian.Uint16(buffer.BytesTo(2)))
	if remain > maxLength {
		return nil, newError("oversize payload")
	}

	if _, err := buffer.ReadFullFrom(pr.Reader, 2); err != nil {
		return nil, newError("failed to read crlf").Base(err)
	}

	var mb buf.MultiBuffer
	for remain > 0 {
		length := buf.Size
		if remain < length {
			length = remain
		}

		b := buf.New()
		n, err := b.ReadFullFrom(pr.Reader, int32(length))
		if err != nil {
			b.Release()
			buf.ReleaseMulti(mb)
			return nil, newError("failed to read payload").Base(err)
		}
		remain -= int(n)
		mb = append(mb, b)
	}

	return mb, nil
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

func ReadHeader(r io.Reader) (*net.Destination, buf.Reader, error) {
	var crlf [2]byte
	var command [1]byte
	var hash [56]byte
	if _, err := io.ReadFull(r, hash[:]); err != nil {
		return nil, nil, newError("failed to read user hash").Base(err)
	}

	if _, err := io.ReadFull(r, crlf[:]); err != nil {
		return nil, nil, newError("failed to read crlf").Base(err)
	}

	if _, err := io.ReadFull(r, command[:]); err != nil {
		return nil, nil, newError("failed to read command").Base(err)
	}

	network := net.Network_TCP
	if command[0] == CommandUDP {
		network = net.Network_UDP
	}

	addr, port, err := addrParser.ReadAddressPort(nil, r)
	if err != nil {
		return nil, nil, newError("failed to read address and port").Base(err)
	}

	if _, err := io.ReadFull(r, crlf[:]); err != nil {
		return nil, nil, newError("failed to read crlf").Base(err)
	}

	var reader buf.Reader
	if network == net.Network_UDP {
		reader = &PacketReader{r}
	} else {
		reader = buf.NewReader(r)
	}

	return &net.Destination{Address: addr, Port: port, Network: network}, reader, nil
}

func WriteHeader(w io.Writer, target net.Destination, account *MemoryAccount) error {
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
	return err
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
