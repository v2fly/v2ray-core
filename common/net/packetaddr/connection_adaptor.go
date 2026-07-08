package packetaddr

import (
	"context"
	gonet "net"
	"sync"
	"time"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/errors"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport"
)

var (
	errNotPacketConn = errors.New("not a packet connection")
)

func ToPacketAddrConn(link *transport.Link, dest net.Destination) (net.PacketConn, error) {
	if !dest.Address.Family().IsDomain() {
		return nil, errNotPacketConn
	}
	switch dest.Address.Domain() {
	case seqPacketMagicAddress:
		return &packetConnectionAdaptor{
			readerAccess: &sync.Mutex{},
			writerAccess: &sync.Mutex{},
			readerBuffer: nil,
			isStream:     false,
			link:         link,
		}, nil
	case streamPacketMagicAddress:
		return &packetConnectionAdaptor{
			readerAccess: &sync.Mutex{},
			writerAccess: &sync.Mutex{},
			streamReader: &buf.BufferedReader{Reader: link.Reader},
			isStream:     true,
			link:         link,
		}, nil
	default:
		return nil, errNotPacketConn
	}
}

func CreatePacketAddrConn(ctx context.Context, dispatcher routing.Dispatcher, isStream bool) (net.PacketConn, error) {
	packetDest := net.Destination{
		Address: net.DomainAddress(seqPacketMagicAddress),
		Port:    0,
		Network: net.Network_UDP,
	}
	if isStream {
		packetDest.Address = net.DomainAddress(streamPacketMagicAddress)
		packetDest.Network = net.Network_TCP
	}
	link, err := dispatcher.Dispatch(ctx, packetDest)
	if err != nil {
		return nil, err
	}
	conn := &packetConnectionAdaptor{
		readerAccess: &sync.Mutex{},
		writerAccess: &sync.Mutex{},
		readerBuffer: nil,
		isStream:     isStream,
		link:         link,
	}
	if isStream {
		conn.streamReader = &buf.BufferedReader{Reader: link.Reader}
	}
	return conn, nil
}

type packetConnectionAdaptor struct {
	readerAccess *sync.Mutex
	writerAccess *sync.Mutex
	readerBuffer buf.MultiBuffer
	streamReader *buf.BufferedReader
	isStream     bool
	link         *transport.Link
}

func (c *packetConnectionAdaptor) ReadFrom(p []byte) (n int, addr gonet.Addr, err error) {
	c.readerAccess.Lock()
	defer c.readerAccess.Unlock()
	if c.isStream {
		packet, err := ExtractPacketFromStream(c.streamReader)
		if err != nil {
			return 0, nil, err
		}
		data, addr, err := ExtractAddressFromPacket(packet)
		if err != nil {
			packet.Release()
			return 0, nil, err
		}
		n = copy(p, data.Bytes())
		data.Release()
		return n, addr, nil
	}
	if c.readerBuffer.IsEmpty() {
		c.readerBuffer, err = c.link.Reader.ReadMultiBuffer()
		if err != nil {
			return 0, nil, err
		}
	}
	c.readerBuffer, n = buf.SplitFirstBytes(c.readerBuffer, p)
	var w *buf.Buffer
	w, addr, err = ExtractAddressFromPacket(buf.FromBytes(p[:n]))
	n = copy(p, w.Bytes())
	w.Release()
	return
}

func (c *packetConnectionAdaptor) WriteTo(p []byte, addr gonet.Addr) (n int, err error) {
	_, ok := addr.(*gonet.UDPAddr)
	if !ok {
		// address other than UDPAddr is not supported, and will be dropped.
		return 0, nil
	}
	payloadLen := len(p)
	var buffer *buf.Buffer
	buffer, err = AttachAddressToPacket(buf.FromBytes(p), addr)
	if err != nil {
		return 0, err
	}
	if c.isStream {
		buffer, err = AttachLengthToPacket(buffer)
		if err != nil {
			return 0, err
		}
	}
	c.writerAccess.Lock()
	defer c.writerAccess.Unlock()
	mb := buf.MultiBuffer{buffer}
	err = c.link.Writer.WriteMultiBuffer(mb)
	if err != nil {
		return 0, err
	}
	return payloadLen, nil
}

func (c *packetConnectionAdaptor) Close() error {
	c.readerAccess.Lock()
	defer c.readerAccess.Unlock()
	c.readerBuffer = buf.ReleaseMulti(c.readerBuffer)
	return common.Interrupt(c.link)
}

func (c packetConnectionAdaptor) LocalAddr() gonet.Addr {
	return &gonet.UnixAddr{Name: "unsupported"}
}

func (c packetConnectionAdaptor) SetDeadline(t time.Time) error {
	return nil
}

func (c packetConnectionAdaptor) SetReadDeadline(t time.Time) error {
	return nil
}

func (c packetConnectionAdaptor) SetWriteDeadline(t time.Time) error {
	return nil
}

func ToPacketAddrConnWrapper(conn net.PacketConn, isStream bool) FusedConnection {
	if isStream {
		return &streamPacketConnWrapper{PacketConn: conn}
	}
	return &packetConnWrapper{conn}
}

type packetConnWrapper struct {
	net.PacketConn
}

func (pc *packetConnWrapper) RemoteAddr() gonet.Addr {
	return nil
}

type FusedConnection interface {
	net.PacketConn
	net.Conn
}

func (pc *packetConnWrapper) Read(p []byte) (n int, err error) {
	recbuf := buf.StackNew()
	recbuf.Extend(2048)
	n, addr, err := pc.ReadFrom(recbuf.Bytes())
	if err != nil {
		return 0, err
	}
	recbuf.Resize(0, int32(n))
	result, err := AttachAddressToPacket(&recbuf, addr)
	if err != nil {
		return 0, err
	}
	n = copy(p, result.Bytes())
	result.Release()
	return n, nil
}

func (pc *packetConnWrapper) Write(p []byte) (n int, err error) {
	data, addr, err := ExtractAddressFromPacket(buf.FromBytes(p))
	if err != nil {
		return 0, err
	}
	_, err = pc.WriteTo(data.Bytes(), addr)
	if err != nil {
		return 0, err
	}
	data.Release()
	return len(p), nil
}

type streamPacketConnWrapper struct {
	net.PacketConn

	readAccess  sync.Mutex
	readBuffer  []byte
	packetRead  []byte
	writeAccess sync.Mutex
	writeBuffer []byte
}

func (pc *streamPacketConnWrapper) RemoteAddr() gonet.Addr {
	return nil
}

func (pc *streamPacketConnWrapper) Read(p []byte) (n int, err error) {
	pc.readAccess.Lock()
	defer pc.readAccess.Unlock()
	for len(pc.readBuffer) == 0 {
		if pc.packetRead == nil {
			pc.packetRead = make([]byte, maxStreamPacketAddrSegmentLen)
		}
		var addr gonet.Addr
		n, addr, err = pc.PacketConn.ReadFrom(pc.packetRead)
		if err != nil {
			return 0, err
		}
		packet, err := AttachAddressToPacket(buf.FromBytes(pc.packetRead[:n]), addr)
		if err != nil {
			return 0, err
		}
		frame, err := AttachLengthToPacket(packet)
		if err != nil {
			return 0, err
		}
		pc.readBuffer = append(pc.readBuffer[:0], frame.Bytes()...)
		frame.Release()
	}
	n = copy(p, pc.readBuffer)
	pc.readBuffer = pc.readBuffer[n:]
	return n, nil
}

func (pc *streamPacketConnWrapper) Write(p []byte) (n int, err error) {
	pc.writeAccess.Lock()
	defer pc.writeAccess.Unlock()
	pc.writeBuffer = append(pc.writeBuffer, p...)
	for len(pc.writeBuffer) >= streamPacketLengthSize {
		length := int(pc.writeBuffer[0])<<8 | int(pc.writeBuffer[1])
		if length == 0 {
			return 0, errors.New("invalid empty packetaddr segment")
		}
		frameEnd := streamPacketLengthSize + length
		if len(pc.writeBuffer) < frameEnd {
			break
		}
		data, addr, err := ExtractAddressFromPacket(buf.FromBytes(pc.writeBuffer[streamPacketLengthSize:frameEnd]))
		if err != nil {
			return 0, err
		}
		if _, err = pc.PacketConn.WriteTo(data.Bytes(), addr); err != nil {
			data.Release()
			return 0, err
		}
		data.Release()
		pc.writeBuffer = pc.writeBuffer[frameEnd:]
	}
	if len(pc.writeBuffer) == 0 && cap(pc.writeBuffer) > buf.Size {
		pc.writeBuffer = nil
	}
	return len(p), nil
}

func (pc *streamPacketConnWrapper) Close() error {
	return pc.PacketConn.Close()
}

func (pc *packetConnWrapper) Close() error {
	return pc.PacketConn.Close()
}

func GetDestinationSubsetOf(dest net.Destination) (bool, error) {
	if !dest.Address.Family().IsDomain() {
		return false, errNotPacketConn
	}
	switch dest.Address.Domain() {
	case seqPacketMagicAddress:
		return false, nil
	case streamPacketMagicAddress:
		return true, nil
	default:
		return false, errNotPacketConn
	}
}
