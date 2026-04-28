package hysteria2

import (
	"context"
	"encoding/binary"
	"io"
	"sync"
	"time"

	"github.com/apernet/quic-go"
	"github.com/apernet/quic-go/quicvarint"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

type interConn struct {
	stream *quic.Stream
	local  net.Addr
	remote net.Addr

	client bool
}

func (c *interConn) Read(b []byte) (int, error) {
	return c.stream.Read(b)
}

func (c *interConn) Write(b []byte) (int, error) {
	if c.client {
		c.client = false
		_, err := c.stream.Write(append(quicvarint.Append(nil, FrameTypeTCPRequest), b...))
		return len(b), err
	}
	return c.stream.Write(b)
}

func (c *interConn) Close() error {
	c.stream.CancelRead(0)
	return c.stream.Close()
}

func (c *interConn) LocalAddr() net.Addr {
	return c.local
}

func (c *interConn) RemoteAddr() net.Addr {
	return c.remote
}

func (c *interConn) SetDeadline(t time.Time) error {
	return c.stream.SetDeadline(t)
}

func (c *interConn) SetReadDeadline(t time.Time) error {
	return c.stream.SetReadDeadline(t)
}

func (c *interConn) SetWriteDeadline(t time.Time) error {
	return c.stream.SetWriteDeadline(t)
}

type InterConn struct {
	local  net.Addr
	remote net.Addr

	write func(p []byte) error
	close func()

	id    uint32
	ch    chan []byte
	time  time.Time
	mutex sync.Mutex
}

func (c *InterConn) Time() time.Time {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.time
}

func (c *InterConn) Update() {
	c.mutex.Lock()
	c.time = time.Now()
	c.mutex.Unlock()
}

func (c *InterConn) Read(b []byte) (int, error) {
	p, ok := <-c.ch
	if !ok {
		return 0, io.EOF
	}
	n := copy(b, p)
	if n != len(p) {
		return 0, io.ErrShortBuffer
	}
	c.Update()
	return n, nil
}

func (c *InterConn) Write(b []byte) (int, error) {
	c.Update()
	binary.BigEndian.PutUint32(b, c.id)
	return len(b), c.write(b)
}

func (c *InterConn) Close() error {
	c.close()
	return nil
}

func (c *InterConn) LocalAddr() net.Addr {
	return c.local
}

func (c *InterConn) RemoteAddr() net.Addr {
	return c.remote
}

func (c *InterConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *InterConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *InterConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type udpSessionManager struct {
	conn   *quic.Conn
	m      map[uint32]*InterConn
	next   uint32
	closed bool
	mutex  sync.RWMutex

	addConn internet.ConnHandler
}

func (m *udpSessionManager) clean() {
	ticker := time.NewTicker(idleCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		if m.closed {
			return
		}

		m.mutex.RLock()
		now := time.Now()
		timeoutConn := make([]*InterConn, 0, len(m.m))
		for _, udpConn := range m.m {
			if now.Sub(udpConn.Time()) > UDPIdleTimeout {
				timeoutConn = append(timeoutConn, udpConn)
			}
		}
		m.mutex.RUnlock()

		for _, udpConn := range timeoutConn {
			m.mutex.Lock()
			if _, found := m.m[udpConn.id]; found {
				close(udpConn.ch)
				delete(m.m, udpConn.id)
			}
			m.mutex.Unlock()
		}
	}
}

func (m *udpSessionManager) run() {
	for {
		d, err := m.conn.ReceiveDatagram(context.Background())
		if err != nil {
			break
		}

		if len(d) < 4 {
			continue
		}
		id := binary.BigEndian.Uint32(d[:4])

		m.feed(id, d)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.closed = true

	for _, udpConn := range m.m {
		close(udpConn.ch)
		delete(m.m, udpConn.id)
	}
}

func (m *udpSessionManager) udp() (net.Conn, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.closed {
		return nil, newError("closed")
	}

	id := m.next
	ch := make(chan []byte, udpMessageChanSize)
	udpConn := &InterConn{
		local:  m.conn.LocalAddr(),
		remote: m.conn.RemoteAddr(),

		write: m.conn.SendDatagram,
		close: func() {
			m.mutex.Lock()
			if _, found := m.m[id]; found {
				close(ch)
				delete(m.m, id)
			}
			m.mutex.Unlock()
		},

		id: id,
		ch: ch,
	}
	m.m[id] = udpConn
	m.next++

	return udpConn, nil
}

func (m *udpSessionManager) feed(id uint32, d []byte) {
	m.mutex.RLock()
	udpConn, ok := m.m[id]
	if ok {
		select {
		case udpConn.ch <- d:
		default:
		}
		m.mutex.RUnlock()
		return
	}
	m.mutex.RUnlock()

	if m.addConn == nil {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	udpConn, ok = m.m[id]
	if !ok {
		ch := make(chan []byte, udpMessageChanSize)
		udpConn = &InterConn{
			local:  m.conn.LocalAddr(),
			remote: m.conn.RemoteAddr(),

			write: m.conn.SendDatagram,
			close: func() {
				m.mutex.Lock()
				if _, found := m.m[id]; found {
					close(ch)
					delete(m.m, id)
				}
				m.mutex.Unlock()
			},

			id:   id,
			ch:   ch,
			time: time.Now(),
		}
		m.m[id] = udpConn
		m.addConn(udpConn)
	}

	select {
	case udpConn.ch <- d:
	default:
	}
}
