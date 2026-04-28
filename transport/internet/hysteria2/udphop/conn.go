package udphop

import (
	"errors"
	"math/rand"
	"net"
	"sync"
	"syscall"
	"time"
)

const (
	packetQueueSize = 1024
	udpBufferSize   = 2048

	defaultHopInterval = 30 * time.Second
)

type udpHopPacketConn struct {
	Addrs          []net.Addr
	HopIntervalMin time.Duration
	HopIntervalMax time.Duration
	ListenUDPFunc  func() (net.PacketConn, error)

	connMutex   sync.RWMutex
	prevConn    net.PacketConn
	currentConn net.PacketConn
	addrIndex   int

	deadline      time.Time
	readDeadline  time.Time
	writeDeadline time.Time

	recvQueue chan *udpPacket
	closeChan chan struct{}
	closed    bool

	bufPool sync.Pool
}

type udpPacket struct {
	Buf  []byte
	N    int
	Addr net.Addr
	Err  error
}

func NewUDPHopPacketConn(addrs []net.Addr, hopIntervalMin time.Duration, hopIntervalMax time.Duration, listenUDPFunc func() (net.PacketConn, error)) (net.PacketConn, error) {
	if len(addrs) == 0 {
		panic("len(addrs) == 0")
	}
	if hopIntervalMin == 0 {
		hopIntervalMin = defaultHopInterval
	}
	if hopIntervalMax == 0 {
		hopIntervalMax = defaultHopInterval
	}
	if hopIntervalMin < 5*time.Second {
		panic("hopIntervalMin < 5*time.Second")
	}
	if hopIntervalMax < 5*time.Second {
		panic("hopIntervalMax < 5*time.Second")
	}
	if hopIntervalMax < hopIntervalMin {
		panic("hopIntervalMax < hopIntervalMin")
	}
	if listenUDPFunc == nil {
		panic("listenUDPFunc is nil")
	}
	curConn, err := listenUDPFunc()
	if err != nil {
		return nil, err
	}
	hConn := &udpHopPacketConn{
		Addrs:          addrs,
		HopIntervalMin: hopIntervalMin,
		HopIntervalMax: hopIntervalMax,
		ListenUDPFunc:  listenUDPFunc,
		prevConn:       nil,
		currentConn:    curConn,
		addrIndex:      rand.Intn(len(addrs)),
		recvQueue:      make(chan *udpPacket, packetQueueSize),
		closeChan:      make(chan struct{}),
		bufPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, udpBufferSize)
			},
		},
	}
	go hConn.recvLoop(curConn)
	go hConn.hopLoop()
	return hConn, nil
}

func (u *udpHopPacketConn) recvLoop(conn net.PacketConn) {
	for {
		buf := u.bufPool.Get().([]byte)
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			u.bufPool.Put(buf)
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				u.recvQueue <- &udpPacket{nil, 0, nil, netErr}
				continue
			}
			return
		}
		select {
		case u.recvQueue <- &udpPacket{buf, n, addr, nil}:
		default:
			u.bufPool.Put(buf)
		}
	}
}

func (u *udpHopPacketConn) hopLoop() {
	timer := time.NewTimer(u.nextHopInterval())
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			u.hop()
			timer.Reset(u.nextHopInterval())
		case <-u.closeChan:
			return
		}
	}
}

func (u *udpHopPacketConn) nextHopInterval() time.Duration {
	if u.HopIntervalMin == u.HopIntervalMax {
		return u.HopIntervalMin
	}
	return u.HopIntervalMin + time.Duration(rand.Int63n(int64(u.HopIntervalMax-u.HopIntervalMin)+1))
}

func (u *udpHopPacketConn) hop() {
	u.connMutex.Lock()
	defer u.connMutex.Unlock()
	if u.closed {
		return
	}
	newConn, err := u.ListenUDPFunc()
	if err != nil {
		return
	}
	if u.prevConn != nil {
		_ = u.prevConn.Close()
	}
	u.prevConn = u.currentConn
	u.currentConn = newConn
	if !u.deadline.IsZero() {
		_ = u.currentConn.SetDeadline(u.deadline)
	}
	if !u.readDeadline.IsZero() {
		_ = u.currentConn.SetReadDeadline(u.readDeadline)
	}
	if !u.writeDeadline.IsZero() {
		_ = u.currentConn.SetWriteDeadline(u.writeDeadline)
	}
	go u.recvLoop(newConn)
	u.addrIndex = rand.Intn(len(u.Addrs))
}

func (u *udpHopPacketConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	for {
		select {
		case p := <-u.recvQueue:
			if p.Err != nil {
				return 0, nil, p.Err
			}
			n := copy(b, p.Buf[:p.N])
			u.bufPool.Put(p.Buf)
			return n, u.Addrs[0], nil
		case <-u.closeChan:
			return 0, nil, net.ErrClosed
		}
	}
}

func (u *udpHopPacketConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	u.connMutex.RLock()
	defer u.connMutex.RUnlock()
	if u.closed {
		return 0, net.ErrClosed
	}
	return u.currentConn.WriteTo(b, u.Addrs[u.addrIndex])
}

func (u *udpHopPacketConn) Close() error {
	u.connMutex.Lock()
	defer u.connMutex.Unlock()
	if u.closed {
		return nil
	}
	if u.prevConn != nil {
		_ = u.prevConn.Close()
	}
	err := u.currentConn.Close()
	close(u.closeChan)
	u.closed = true
	u.Addrs = nil
	return err
}

func (u *udpHopPacketConn) LocalAddr() net.Addr {
	u.connMutex.RLock()
	defer u.connMutex.RUnlock()
	return u.currentConn.LocalAddr()
}

func (u *udpHopPacketConn) SetDeadline(t time.Time) error {
	u.connMutex.Lock()
	defer u.connMutex.Unlock()
	u.deadline = t
	u.readDeadline = t
	u.writeDeadline = t
	if u.prevConn != nil {
		_ = u.prevConn.SetDeadline(t)
	}
	return u.currentConn.SetDeadline(t)
}

func (u *udpHopPacketConn) SetReadDeadline(t time.Time) error {
	u.connMutex.Lock()
	defer u.connMutex.Unlock()
	u.deadline = time.Time{}
	u.readDeadline = t
	if u.prevConn != nil {
		_ = u.prevConn.SetReadDeadline(t)
	}
	return u.currentConn.SetReadDeadline(t)
}

func (u *udpHopPacketConn) SetWriteDeadline(t time.Time) error {
	u.connMutex.Lock()
	defer u.connMutex.Unlock()
	u.deadline = time.Time{}
	u.writeDeadline = t
	if u.prevConn != nil {
		_ = u.prevConn.SetWriteDeadline(t)
	}
	return u.currentConn.SetWriteDeadline(t)
}

func (u *udpHopPacketConn) SyscallConn() (syscall.RawConn, error) {
	u.connMutex.RLock()
	defer u.connMutex.RUnlock()
	sc, ok := u.currentConn.(syscall.Conn)
	if !ok {
		return nil, errors.New("not supported")
	}
	return sc.SyscallConn()
}

func ToAddrs(ip net.IP, ports []uint32) []net.Addr {
	var addrs []net.Addr
	for _, port := range ports {
		addr := &net.UDPAddr{
			IP:   ip,
			Port: int(port),
		}
		addrs = append(addrs, addr)
	}
	return addrs
}
