package fusedPacketConn

import (
	"errors"
	"time"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

var errClosed = errors.New("fused packet conn is closed")

// FusedPacketConn combines two PacketConn socket to create a dual stack PacketConn
// When sending packet, the correct PacketConn for that destination address will be chosen
// When receiving packet, will receive packet from either socket
// Other operations will be done on both conn
type FusedPacketConn struct {
	ipv6 net.PacketConn
	ipv4 net.PacketConn

	readCh chan readResult
	done   chan struct{}

	localAddrPreferIPv6 bool
}

type readResult struct {
	data []byte
	addr net.Addr
	err  error
}

func NewFusedPacketConn(ipv4, ipv6 net.PacketConn, readBufSize int, localAddrPreferIPv6 bool) *FusedPacketConn {
	f := &FusedPacketConn{
		ipv4:                ipv4,
		ipv6:                ipv6,
		readCh:              make(chan readResult, 2),
		done:                make(chan struct{}),
		localAddrPreferIPv6: localAddrPreferIPv6,
	}
	go f.readLoop(ipv4, readBufSize)
	go f.readLoop(ipv6, readBufSize)
	return f
}

func (f *FusedPacketConn) readLoop(conn net.PacketConn, bufSize int) {
	for {
		buf := make([]byte, bufSize)
		n, addr, err := conn.ReadFrom(buf)
		select {
		case <-f.done:
			return
		case f.readCh <- readResult{data: buf[:n], addr: addr, err: err}:
		}
		if err != nil {
			return
		}
	}
}

func (f *FusedPacketConn) ReadFrom(p []byte) (int, net.Addr, error) {
	select {
	case <-f.done:
		return 0, nil, errClosed
	case r := <-f.readCh:
		if r.err != nil {
			return 0, r.addr, r.err
		}
		n := copy(p, r.data)
		return n, r.addr, nil
	}
}

func isIPv4Addr(addr net.Addr) bool {
	udpAddr, ok := addr.(*net.UDPAddr)
	if !ok {
		return false
	}
	return udpAddr.IP.To4() != nil
}

func (f *FusedPacketConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	if isIPv4Addr(addr) {
		return f.ipv4.WriteTo(p, addr)
	}
	return f.ipv6.WriteTo(p, addr)
}

func (f *FusedPacketConn) Close() error {
	close(f.done)
	err4 := f.ipv4.Close()
	err6 := f.ipv6.Close()
	if err4 != nil {
		return err4
	}
	return err6
}

func (f *FusedPacketConn) LocalAddr() net.Addr {
	if f.localAddrPreferIPv6 {
		return f.ipv6.LocalAddr()
	}
	return f.ipv4.LocalAddr()
}

func (f *FusedPacketConn) SetDeadline(t time.Time) error {
	err4 := f.ipv4.SetDeadline(t)
	err6 := f.ipv6.SetDeadline(t)
	if err4 != nil {
		return err4
	}
	return err6
}

func (f *FusedPacketConn) SetReadDeadline(t time.Time) error {
	err4 := f.ipv4.SetReadDeadline(t)
	err6 := f.ipv6.SetReadDeadline(t)
	if err4 != nil {
		return err4
	}
	return err6
}

func (f *FusedPacketConn) SetWriteDeadline(t time.Time) error {
	err4 := f.ipv4.SetWriteDeadline(t)
	err6 := f.ipv6.SetWriteDeadline(t)
	if err4 != nil {
		return err4
	}
	return err6
}
