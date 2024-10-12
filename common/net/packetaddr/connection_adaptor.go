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
	errUnsupported   = errors.New("unsupported action")
)

func ToPacketAddrConn(link *transport.Link, dest net.Destination) (net.PacketConn, error) {
	if !dest.Address.Family().IsDomain() {
		return nil, errNotPacketConn
	}
	switch dest.Address.Domain() {
	case seqPacketMagicAddress:
		return &packetConnectionAdaptor{
			readerAccess: &sync.Mutex{},
			readerBuffer: nil,
			link:         link,
		}, nil
	default:
		return nil, errNotPacketConn
	}
}

func CreatePacketAddrConn(ctx context.Context, dispatcher routing.Dispatcher, isStream bool) (net.PacketConn, error) {
	if isStream {
		return nil, errUnsupported
	}
	packetDest := net.Destination{
		Address: net.DomainAddress(seqPacketMagicAddress),
		Port:    0,
		Network: net.Network_UDP,
	}
	link, err := dispatcher.Dispatch(ctx, packetDest)
	if err != nil {
		return nil, err
	}
	return &packetConnectionAdaptor{
		readerAccess: &sync.Mutex{},
		readerBuffer: nil,
		link:         link,
	}, nil
}

type packetConnectionAdaptor struct {
	readerAccess *sync.Mutex
	readerBuffer buf.MultiBuffer
	link         *transport.Link
}

func (c *packetConnectionAdaptor) ReadFrom(p []byte) (n int, addr gonet.Addr, err error) {
	c.readerAccess.Lock()
	defer c.readerAccess.Unlock()
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
	n, addr, err := pc.PacketConn.ReadFrom(recbuf.Bytes())
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
	_, err = pc.PacketConn.WriteTo(data.Bytes(), addr)
	if err != nil {
		return 0, err
	}
	data.Release()
	return len(p), nil
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
	default:
		return false, errNotPacketConn
	}
}
