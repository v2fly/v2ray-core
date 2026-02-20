package wgcommon

import (
	gonet "net"
	"net/netip"
	"sync"
	"time"

	"golang.zx2c4.com/wireguard/conn"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

// netPacketConnToWg is machine generated
type netPacketConnToWg struct {
	mu         sync.Mutex
	conn       net.PacketConn
	actualPort uint16
	closed     bool // tracks whether the bind is logically closed (not the conn)
}

// NewNetPacketConnToWg constructs a wireguard conn.Bind adapter from a
// common/net.PacketConn. It returns a Bind implementation that delegates
// reads/writes to the provided PacketConn.
//
// Important: the Bind does NOT own the PacketConn lifecycle. WireGuard calls
// Close() + Open() internally during BindUpdate(); Close() here only marks
// the bind as logically closed without closing the underlying conn, so that
// Open() can re-use it.
func NewNetPacketConnToWg(c net.PacketConn) conn.Bind {
	if c == nil {
		return &netPacketConnToWg{}
	}
	n := &netPacketConnToWg{conn: c, closed: true}
	if la := c.LocalAddr(); la != nil {
		if ua, ok := la.(*gonet.UDPAddr); ok {
			n.actualPort = uint16(ua.Port)
		}
	}
	return n
}

// wgEndpoint is a minimal implementation of conn.Endpoint backed by netip.AddrPort.
type wgEndpoint struct {
	ap     netip.AddrPort
	hasSrc bool
	srcIP  netip.Addr
}

func (e *wgEndpoint) ClearSrc() {
	e.hasSrc = false
}

func (e *wgEndpoint) SrcToString() string {
	if !e.hasSrc {
		return ""
	}
	// return just IP (no port) if src port is unknown
	return e.srcIP.String()
}

func (e *wgEndpoint) DstToString() string {
	return e.ap.String()
}

func (e *wgEndpoint) DstToBytes() []byte {
	b, _ := e.ap.MarshalBinary()
	return b
}

func (e *wgEndpoint) DstIP() netip.Addr {
	return e.ap.Addr()
}

func (e *wgEndpoint) SrcIP() netip.Addr {
	if e.hasSrc {
		return e.srcIP
	}
	return netip.Addr{}
}

func (n *netPacketConnToWg) Open(port uint16) (fns []conn.ReceiveFunc, actualPort uint16, err error) {
	if n.conn == nil {
		return nil, 0, nil
	}
	n.mu.Lock()
	n.closed = false
	n.mu.Unlock()

	// Clear the read deadline that Close() may have set so reads can proceed.
	_ = n.conn.SetReadDeadline(time.Time{})

	// determine actualPort from LocalAddr if possible
	if la := n.conn.LocalAddr(); la != nil {
		if ua, ok := la.(*gonet.UDPAddr); ok {
			n.actualPort = uint16(ua.Port)
		}
	}
	actualPort = n.actualPort

	fn := func(packets [][]byte, sizes []int, eps []conn.Endpoint) (int, error) {
		var i int
		for i = 0; i < len(packets); i++ {
			nRead, addr, err := n.conn.ReadFrom(packets[i])
			if err != nil {
				if i == 0 {
					return 0, err
				}
				return i, nil
			}
			sizes[i] = nRead
			// build endpoint from addr
			if udpAddr, ok := addr.(*gonet.UDPAddr); ok {
				ip, _ := netip.AddrFromSlice(udpAddr.IP)
				ap := netip.AddrPortFrom(ip, uint16(udpAddr.Port))
				eps[i] = &wgEndpoint{ap: ap}
			} else {
				// fallback: parse string
				s := addr.String()
				if ap, perr := netip.ParseAddrPort(s); perr == nil {
					eps[i] = &wgEndpoint{ap: ap}
				} else {
					eps[i] = &wgEndpoint{}
				}
			}
		}
		return i, nil
	}
	return []conn.ReceiveFunc{fn}, actualPort, nil
}

func (n *netPacketConnToWg) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.closed = true
	// Do NOT close the underlying conn here. WireGuard calls Close()+Open()
	// internally during BindUpdate(). The actual PacketConn lifecycle is
	// managed externally by the session that created it.
	//
	// Set a past read deadline to unblock any pending ReadFrom calls in
	// receive goroutines so that WireGuard's stopping.Wait() can complete.
	if n.conn != nil {
		_ = n.conn.SetReadDeadline(time.Unix(1, 0))
	}
	return nil
}

func (n *netPacketConnToWg) SetMark(mark uint32) error {
	// best-effort: underlying PacketConn may not support setting fwmark; ignore.
	return nil
}

func (n *netPacketConnToWg) Send(bufs [][]byte, ep conn.Endpoint) error {
	if n.conn == nil {
		return nil
	}
	// Use DstToString to obtain "ip:port" and resolve to UDPAddr
	addrStr := ep.DstToString()
	udpAddr, err := gonet.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		return err
	}
	for _, b := range bufs {
		if _, werr := n.conn.WriteTo(b, udpAddr); werr != nil {
			return werr
		}
	}
	return nil
}

func (n *netPacketConnToWg) ParseEndpoint(s string) (conn.Endpoint, error) {
	ap, err := netip.ParseAddrPort(s)
	if err != nil {
		return nil, err
	}
	return &wgEndpoint{ap: ap}, nil
}

func (n *netPacketConnToWg) BatchSize() int {
	// underlying common/net.PacketConn may not support batch; report 1.
	return 1
}
