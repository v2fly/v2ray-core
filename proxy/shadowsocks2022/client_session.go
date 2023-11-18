package shadowsocks2022

import (
	"context"
	"crypto/rand"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"io"
	gonet "net"
	"sync"
	"time"
)

func NewClientUDPSession(ctx context.Context, conn io.ReadWriteCloser, packetProcessor UDPClientPacketProcessor) *ClientUDPSession {
	session := &ClientUDPSession{
		locker:          &sync.Mutex{},
		conn:            conn,
		packetProcessor: packetProcessor,
		sessionMap:      make(map[string]*ClientUDPSessionConn),
	}
	session.ctx, session.finish = context.WithCancel(ctx)

	go session.KeepReading()
	return session
}

type ClientUDPSession struct {
	locker *sync.Mutex

	conn            io.ReadWriteCloser
	packetProcessor UDPClientPacketProcessor
	sessionMap      map[string]*ClientUDPSessionConn

	ctx    context.Context
	finish func()
}

func (c *ClientUDPSession) Close() error {
	c.finish()
	return c.conn.Close()
}

func (c *ClientUDPSession) WriteUDPRequest(request *UDPRequest) error {
	buffer := buf.New()
	defer buffer.Release()
	err := c.packetProcessor.EncodeUDPRequest(request, buffer)
	if request.Payload != nil {
		request.Payload.Release()
	}
	if err != nil {
		return newError("unable to encode udp request").Base(err)
	}
	_, err = c.conn.Write(buffer.Bytes())
	if err != nil {
		return newError("unable to write to conn").Base(err)
	}
	return nil
}

func (c *ClientUDPSession) KeepReading() {
	for c.ctx.Err() == nil {
		udpResp := &UDPResponse{}
		buffer := make([]byte, 1600)
		n, err := c.conn.Read(buffer)
		if err != nil {
			newError("unable to read from conn").Base(err).WriteToLog()
			return
		}
		if n != 0 {
			err := c.packetProcessor.DecodeUDPResp(buffer[:n], udpResp)
			if err != nil {
				newError("unable to decode udp response").Base(err).WriteToLog()
				continue
			}
			c.locker.Lock()
			session, ok := c.sessionMap[string(udpResp.ClientSessionID[:])]
			if ok {
				select {
				case session.readChan <- udpResp:
				default:
				}
			} else {
				newError("misbehaving server: unknown client session ID").Base(err).WriteToLog()
			}
			c.locker.Unlock()
		}
	}
}

func (c *ClientUDPSession) NewSessionConn() (internet.AbstractPacketConn, error) {
	sessionID := make([]byte, 8)
	_, err := rand.Read(sessionID)
	if err != nil {
		return nil, newError("unable to generate session id").Base(err)
	}

	connctx, connfinish := context.WithCancel(c.ctx)

	sessionConn := &ClientUDPSessionConn{
		sessionID:         string(sessionID),
		readChan:          make(chan *UDPResponse, 16),
		parent:            c,
		ctx:               connctx,
		finish:            connfinish,
		nextWritePacketID: 0,
	}
	c.locker.Lock()
	c.sessionMap[sessionConn.sessionID] = sessionConn
	c.locker.Unlock()
	return sessionConn, nil
}

type ClientUDPSessionConn struct {
	sessionID string
	readChan  chan *UDPResponse
	parent    *ClientUDPSession

	nextWritePacketID uint64

	ctx    context.Context
	finish func()
}

func (c *ClientUDPSessionConn) Close() error {
	delete(c.parent.sessionMap, c.sessionID)
	c.finish()
	return nil
}

func (c *ClientUDPSessionConn) WriteTo(p []byte, addr gonet.Addr) (n int, err error) {
	thisPacketID := c.nextWritePacketID
	c.nextWritePacketID += 1
	req := &UDPRequest{
		SessionID: [8]byte{},
		PacketID:  thisPacketID,
		TimeStamp: uint64(time.Now().Unix()),
		Address:   net.IPAddress(addr.(*gonet.UDPAddr).IP),
		Port:      addr.(*net.UDPAddr).Port,
		Payload:   nil,
	}
	copy(req.SessionID[:], c.sessionID)
	req.Payload = buf.New()
	req.Payload.Write(p)
	err = c.parent.WriteUDPRequest(req)
	if err != nil {
		return 0, newError("unable to write to parent session").Base(err)
	}
	return len(p), nil
}

func (c *ClientUDPSessionConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	select {
	case <-c.ctx.Done():
		return 0, nil, io.EOF
	case resp := <-c.readChan:
		n = copy(p, resp.Payload.Bytes())
		resp.Payload.Release()
		addr = &net.UDPAddr{IP: resp.Address.IP(), Port: int(resp.Port)}
	}
	return
}
