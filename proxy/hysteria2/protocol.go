package hysteria2

import (
	"encoding/binary"
	"io"
	gonet "net"

	hyProtocol "github.com/apernet/hysteria/core/international/protocol"
	"github.com/apernet/quic-go/quicvarint"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
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
	commandTCP byte = 1
	commandUDP byte = 3
)

// ConnWriter is TCP Connection Writer Wrapper for trojan protocol
type ConnWriter struct {
	io.Writer
	Target     net.Destination
	Account    *MemoryAccount
	headerSent bool
}

// Write implements io.Writer
func (c *ConnWriter) Write(p []byte) (n int, err error) {
	if !c.headerSent {
		if err := c.writeHeader(); err != nil {
			return 0, newError("failed to write request header").Base(err)
		}
	}

	return c.Writer.Write(p)
}

// WriteMultiBuffer implements buf.Writer
func (c *ConnWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	defer buf.ReleaseMulti(mb)

	for _, b := range mb {
		if !b.IsEmpty() {
			if _, err := c.Write(b.Bytes()); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *ConnWriter) WriteHeader() error {
	if !c.headerSent && c.Target.Network != net.Network_TCP {
		if err := c.writeHeader(); err != nil {
			return err
		}
	}
	return nil
}

func QuicLen(s int) int {
	return int(quicvarint.Len(uint64(s)))
}

func (c *ConnWriter) writeHeader() error {
	padding := "Jimmy Was Here"
	paddingLen := len(padding)
	addressAndPort := c.Target.Address.String() + ":" + c.Target.Port.String()
	addressLen := len(addressAndPort)
	size := QuicLen(addressLen) + addressLen + QuicLen(paddingLen) + paddingLen

	buf := make([]byte, size)
	i := hyProtocol.VarintPut(buf, uint64(addressLen))
	i += copy(buf[i:], addressAndPort)
	i += hyProtocol.VarintPut(buf[i:], uint64(paddingLen))
	copy(buf[i:], padding)

	_, err := c.Writer.Write(buf)
	if err == nil {
		c.headerSent = true
	}
	return err
}

// PacketWriter UDP Connection Writer Wrapper for trojan protocol
type PacketWriter struct {
	io.Writer
	Target net.Destination
}

// WriteMultiBuffer implements buf.Writer
func (w *PacketWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	for _, b := range mb {
		if b.IsEmpty() {
			continue
		}
		if _, err := w.writePacket(b.Bytes(), w.Target); err != nil {
			buf.ReleaseMulti(mb)
			return err
		}
	}

	return nil
}

// WriteMultiBufferWithMetadata writes udp packet with destination specified
func (w *PacketWriter) WriteMultiBufferWithMetadata(mb buf.MultiBuffer, dest net.Destination) error {
	for _, b := range mb {
		if b.IsEmpty() {
			continue
		}
		if _, err := w.writePacket(b.Bytes(), dest); err != nil {
			buf.ReleaseMulti(mb)
			return err
		}
	}

	return nil
}

func (w *PacketWriter) WriteTo(payload []byte, addr gonet.Addr) (int, error) {
	dest := net.DestinationFromAddr(addr)

	return w.writePacket(payload, dest)
}

func (w *PacketWriter) writePacket(payload []byte, dest net.Destination) (int, error) { // nolint: unparam
	var addrPortLen int32
	switch dest.Address.Family() {
	case net.AddressFamilyDomain:
		if protocol.IsDomainTooLong(dest.Address.Domain()) {
			return 0, newError("Super long domain is not supported: ", dest.Address.Domain())
		}
		addrPortLen = 1 + 1 + int32(len(dest.Address.Domain())) + 2
	case net.AddressFamilyIPv4:
		addrPortLen = 1 + 4 + 2
	case net.AddressFamilyIPv6:
		addrPortLen = 1 + 16 + 2
	default:
		panic("Unknown address type.")
	}

	length := len(payload)
	lengthBuf := [2]byte{}
	binary.BigEndian.PutUint16(lengthBuf[:], uint16(length))

	buffer := buf.NewWithSize(addrPortLen + 2 + 2 + int32(length))
	defer buffer.Release()

	if err := addrParser.WriteAddressPort(buffer, dest.Address, dest.Port); err != nil {
		return 0, err
	}
	if _, err := buffer.Write(lengthBuf[:]); err != nil {
		return 0, err
	}
	if _, err := buffer.Write(crlf); err != nil {
		return 0, err
	}
	if _, err := buffer.Write(payload); err != nil {
		return 0, err
	}
	_, err := w.Write(buffer.Bytes())
	if err != nil {
		return 0, err
	}

	return length, nil
}

// ConnReader is TCP Connection Reader Wrapper for trojan protocol
type ConnReader struct {
	io.Reader
	Target net.Destination
}

// Read implements io.Reader
func (c *ConnReader) Read(p []byte) (int, error) {
	return c.Reader.Read(p)
}

// ReadMultiBuffer implements buf.Reader
func (c *ConnReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	b := buf.New()
	_, err := b.ReadFrom(c)
	return buf.MultiBuffer{b}, err
}

// PacketPayload combines udp payload and destination
type PacketPayload struct {
	Target net.Destination
	Buffer buf.MultiBuffer
}

// PacketReader is UDP Connection Reader Wrapper for trojan protocol
type PacketReader struct {
	io.Reader
}

// ReadMultiBuffer implements buf.Reader
func (r *PacketReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	p, err := r.ReadMultiBufferWithMetadata()
	if p != nil {
		return p.Buffer, err
	}
	return nil, err
}

// ReadMultiBufferWithMetadata reads udp packet with destination
func (r *PacketReader) ReadMultiBufferWithMetadata() (*PacketPayload, error) {
	addr, port, err := addrParser.ReadAddressPort(nil, r)
	if err != nil {
		return nil, newError("failed to read address and port").Base(err)
	}

	var lengthBuf [2]byte
	if _, err := io.ReadFull(r, lengthBuf[:]); err != nil {
		return nil, newError("failed to read payload length").Base(err)
	}

	length := binary.BigEndian.Uint16(lengthBuf[:])

	var crlf [2]byte
	if _, err := io.ReadFull(r, crlf[:]); err != nil {
		return nil, newError("failed to read crlf").Base(err)
	}

	dest := net.UDPDestination(addr, port)

	b := buf.NewWithSize(int32(length))
	_, err = b.ReadFullFrom(r, int32(length))
	if err != nil {
		return nil, newError("failed to read payload").Base(err)
	}

	return &PacketPayload{Target: dest, Buffer: buf.MultiBuffer{b}}, nil
}

type PacketConnectionReader struct {
	reader  *PacketReader
	payload *PacketPayload
}

func (r *PacketConnectionReader) ReadFrom(p []byte) (n int, addr gonet.Addr, err error) {
	if r.payload == nil || r.payload.Buffer.IsEmpty() {
		r.payload, err = r.reader.ReadMultiBufferWithMetadata()
		if err != nil {
			return
		}
	}

	addr = &gonet.UDPAddr{
		IP:   r.payload.Target.Address.IP(),
		Port: int(r.payload.Target.Port),
	}

	r.payload.Buffer, n = buf.SplitFirstBytes(r.payload.Buffer, p)

	return
}
