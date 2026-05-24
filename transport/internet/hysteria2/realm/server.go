package realm

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/pion/stun/v3"
)

const defaultEventBuffer = 16
const defaultStunCacheTTL = time.Second * 10
const defaultHeartbeatInterval = time.Second * 15

type PunchPacketEvent struct {
	Addr   netip.AddrPort
	Packet PunchPacket
}

type PunchPacketEventWithMeta struct {
	Meta PunchMetadata
	Ch   chan PunchPacketEvent
}

type STUNPacketEvent struct {
	Message *stun.Message
	Addr    netip.AddrPort
}

type PunchPacketConn struct {
	cleaned     chan struct{}
	ctx         context.Context
	cancel      context.CancelFunc
	rClient     *Client
	id          string
	stunServers []string
	net.PacketConn

	events map[string]*PunchPacketEventWithMeta
	stun   chan STUNPacketEvent
	mu     sync.Mutex

	locals     []netip.AddrPort
	localsMu   sync.Mutex
	localsLast time.Time
}

func NewPunchPacketConn(scheme, host, port, token, id string, stunServers []string, raw net.PacketConn) (*PunchPacketConn, error) {
	start := time.Now()
	servers := resolveSTUNServers(raw.LocalAddr().(*net.UDPAddr).IP, stunServers)
	newError("[realm] get stun servers ", servers, " with ", time.Since(start)).AtDebug().WriteToLog()
	if len(servers) == 0 {
		return nil, errors.New("empty stun servers")
	}

	start = time.Now()
	locals := Discover(raw, servers)
	newError("[realm] get stun locals ", locals, " with ", time.Since(start)).AtDebug().WriteToLog()
	if len(locals) == 0 {
		return nil, errors.New("empty stun locals")
	}

	rClient, err := NewClient(scheme, host, port, token)
	if err != nil {
		return nil, newError("http create").Base(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	conn := &PunchPacketConn{
		cleaned:     make(chan struct{}),
		ctx:         ctx,
		cancel:      cancel,
		rClient:     rClient,
		id:          id,
		stunServers: stunServers,
		PacketConn:  raw,

		events: make(map[string]*PunchPacketEventWithMeta),
		stun:   make(chan STUNPacketEvent, defaultEventBuffer),

		locals:     locals,
		localsLast: time.Now(),
	}

	go conn.run()

	return conn, nil
}

func (c *PunchPacketConn) ReadFrom(p []byte) (int, net.Addr, error) {
	for {
		n, addr, err := c.PacketConn.ReadFrom(p)
		if err != nil {
			return n, addr, err
		}
		if c.addSTUN(p[:n]) {
			continue
		}
		if c.addPunch(p[:n], addr) {
			continue
		}
		return n, addr, nil
	}
}

func (c *PunchPacketConn) Close() error {
	c.cancel()
	<-c.cleaned
	return c.PacketConn.Close()
}

func (c *PunchPacketConn) SyscallConn() (syscall.RawConn, error) {
	sc, ok := c.PacketConn.(syscall.Conn)
	if !ok {
		return nil, errors.ErrUnsupported
	}
	return sc.SyscallConn()
}

func (c *PunchPacketConn) addSTUN(packet []byte) bool {
	if !stun.IsMessage(packet) {
		return false
	}
	msg, addr, err := parseSTUNBindingResponse(packet)
	if err != nil {
		return false
	}
	select {
	case c.stun <- STUNPacketEvent{Message: msg, Addr: addr}:
	default:
	}
	return true
}

func (c *PunchPacketConn) addPunch(packet []byte, addr net.Addr) bool {
	var added bool
	c.mu.Lock()
	for _, ev := range c.events {
		punchPacket, err := DecodePunchPacket(packet, ev.Meta)
		if err != nil {
			continue
		}
		select {
		case ev.Ch <- PunchPacketEvent{
			Addr:   addr.(*net.UDPAddr).AddrPort(),
			Packet: punchPacket,
		}:
		default:
		}
		added = true
		break
	}
	c.mu.Unlock()
	return added
}

func (c *PunchPacketConn) run() {
	backoff := time.Second
retry:
	resp, err := c.rClient.Register(c.ctx, c.id, addrPortStrings(c.getlocals(false)))
	if err != nil {
		newError("[realm] failed to register session for ", c.id, " retry in ", backoff).Base(err).AtDebug().WriteToLog()
		if waitctx(c.ctx, backoff) {
			close(c.cleaned)
			return
		}
		backoff *= 2
		if backoff > 30*time.Second {
			backoff = 30 * time.Second
		}
		goto retry
	}
	backoff = time.Second
	newError("[realm] ", c.id, " sesssion ", resp.SessionID, " ", resp.TTL, " registered").AtDebug().WriteToLog()

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 2)
	go c.heartbeatLoop(ctx, resp.SessionID, resp.TTL, errCh)
	go c.eventsLoop(ctx, resp.SessionID, resp.TTL, errCh)
	select {
	case <-c.ctx.Done():
	case err = <-errCh:
	}
	cancel()
	newError("[realm] session ", resp.SessionID, " end with err ", err).AtDebug().WriteToLog()

	select {
	case <-c.ctx.Done():
		_ = c.rClient.Deregister(context.Background(), c.id, resp.SessionID)
		newError("[realm] ", c.id, " ", resp.SessionID, " deregistered").AtDebug().WriteToLog()
		close(c.cleaned)
		return
	default:
		goto retry
	}
}

func (c *PunchPacketConn) heartbeatLoop(ctx context.Context, sid string, ttl int, errCh chan<- error) {
	interval := defaultHeartbeatInterval
	if ttl > 0 {
		interval = time.Second * time.Duration(ttl) / 2
	}

	last := time.Now()
	cur := c.getlocals(false)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			errCh <- nil
			return
		case <-ticker.C:
			req := HeartbeatRequest{}
			if new := c.getlocals(false); !slices.Equal(cur, new) {
				cur = new
				req.Addresses = addrPortStrings(cur)
			}
			start := time.Now()
			resp, err := c.rClient.Heartbeat(ctx, c.id, sid, req)
			if err != nil {
				var statusErr *StatusError
				if errors.As(err, &statusErr) && (statusErr.StatusCode == http.StatusUnauthorized || statusErr.StatusCode == http.StatusNotFound) {
					errCh <- errors.New("session invalid")
					return
				}
				if time.Since(last) > time.Second*time.Duration(ttl) {
					errCh <- errors.New("session lost")
					return
				}
				continue
			}
			last = start
			newError("[realm] heartbeat ", resp.TTL, " with ", time.Since(start)).AtDebug().WriteToLog()
			if resp.TTL > 0 && resp.TTL != ttl {
				ttl = resp.TTL
				ticker.Reset(time.Second * time.Duration(ttl) / 2)
			}
		}
	}
}

func (c *PunchPacketConn) eventsLoop(ctx context.Context, sid string, ttl int, errCh chan<- error) {
	backoff := time.Second
	last := time.Now()
	for {
		start := time.Now()
		stream, err := c.rClient.Events(ctx, c.id, sid)
		if err != nil {
			var statusErr *StatusError
			if errors.As(err, &statusErr) && (statusErr.StatusCode == http.StatusUnauthorized || statusErr.StatusCode == http.StatusNotFound) {
				errCh <- errors.New("session invalid")
				return
			}
			if time.Since(last) > time.Second*time.Duration(ttl) {
				errCh <- errors.New("session lost")
				return
			}
			newError("[realm] ", sid, " open stream err ", err, " retry in ", backoff).AtDebug().WriteToLog()
			if waitctx(ctx, backoff) {
				errCh <- nil
				return
			}
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			continue
		}
		backoff = time.Second
		last = start
		newError("[realm] open stream with ", time.Since(start)).AtDebug().WriteToLog()
		for {
			ev, err := stream.Next()
			if err != nil {
				_ = stream.Close()
				break
			}
			last = time.Now()
			go c.punch(ctx, sid, ev, defaultPunchTimeout, defaultPunchInterval)
		}
	}
}

func (c *PunchPacketConn) punch(ctx context.Context, sid string, ev *PunchEvent, timeout, interval time.Duration) {
	newError("[realm] start punch event ", ev.Nonce, " ", ev.Addresses).AtDebug().WriteToLog()

	locals := c.getlocals(false)

	peers, _ := parseAddrPorts(ev.Addresses)
	newError("[realm] ", ev.Nonce, " get peers ", peers).AtDebug().WriteToLog()
	filteredPeers, seen := candidatePunchAddrs(locals, peers)
	newError("[realm] ", ev.Nonce, " filtered peers ", filteredPeers).AtDebug().WriteToLog()
	expandedPeers := expandSymmetricNATCandidates(filteredPeers, seen)
	newError("[realm] ", ev.Nonce, " expanded peers ", expandedPeers).AtDebug().WriteToLog()

	if len(expandedPeers) == 0 {
		newError("[realm] punch ", ev.Nonce, " FAIL > empty peers")
		return
	}

	start := time.Now()
	_ = c.rClient.ConnectResponse(ctx, c.id, sid, ev.Nonce, addrPortStrings(locals))
	newError("[realm] ", ev.Nonce, " connect response ", locals, " with ", time.Since(start)).AtDebug().WriteToLog()

	c.mu.Lock()
	if _, ok := c.events[ev.Nonce]; ok {
		c.mu.Unlock()
		return
	}
	ch := make(chan PunchPacketEvent, defaultEventBuffer)
	c.events[ev.Nonce] = &PunchPacketEventWithMeta{Meta: ev.PunchMetadata, Ch: ch}
	c.mu.Unlock()

	start = time.Now()
	sendPunchPackets(c, expandedPeers, ev.PunchMetadata, PunchPacketHello)
	deadline := time.NewTimer(timeout)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			newError("[realm] punch ", ev.Nonce, " FAIL > session end").AtDebug().WriteToLog()
			goto end
		case <-deadline.C:
			newError("[realm] punch ", ev.Nonce, " FAIL > timeout").AtDebug().WriteToLog()
			goto end
		case <-ticker.C:
			sendPunchPackets(c, expandedPeers, ev.PunchMetadata, PunchPacketHello)
		case event := <-ch:
			if event.Packet.Type == PunchPacketHello {
				sendPunchPacket(c, event.Addr, ev.PunchMetadata, PunchPacketAck)
			}
			newError("[realm] punch ", ev.Nonce, " SUCCESS ", event.Addr, " with ", time.Since(start)).AtDebug().WriteToLog()
			goto end
		}
	}
end:
	deadline.Stop()
	ticker.Stop()

	c.mu.Lock()
	delete(c.events, ev.Nonce)
	close(ch)
	c.mu.Unlock()
}

func (c *PunchPacketConn) getlocals(force bool) []netip.AddrPort {
	c.localsMu.Lock()
	if force || time.Since(c.localsLast) > defaultStunCacheTTL {
		start := time.Now()
		servers := resolveSTUNServers(c.LocalAddr().(*net.UDPAddr).IP, c.stunServers)
		newError("[realm] get stun servers ", servers, " with ", time.Since(start)).AtDebug().WriteToLog()
		if len(servers) > 0 {
			start = time.Now()
			locals := DiscoverWithDemux(c.WriteTo, c.stun, servers)
			newError("[realm] get stun locals ", locals, " with ", time.Since(start)).AtDebug().WriteToLog()
			if len(locals) > 0 {
				c.locals = locals
				c.localsLast = time.Now()
			}
		}
	}
	locals := append([]netip.AddrPort(nil), c.locals...)
	c.localsMu.Unlock()
	return locals
}

func waitctx(ctx context.Context, t time.Duration) bool {
	timer := time.NewTimer(t)
	defer timer.Stop()
	select {
	case <-timer.C:
		return false
	case <-ctx.Done():
		return true
	}
}
