package shadowsocks2022

import (
	"context"
	"crypto/rand"
	"io"
	gonet "net"
	"sync"
	"time"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"

	"github.com/pion/transport/v2/replaydetector"
)

func NewClientUDPSession(ctx context.Context, conn io.ReadWriteCloser, packetProcessor UDPClientPacketProcessor) *ClientUDPSession {
	session := &ClientUDPSession{
		locker:          &sync.RWMutex{},
		conn:            conn,
		packetProcessor: packetProcessor,
		sessionMap:      make(map[string]*ClientUDPSessionConn),
		sessionMapAlias: make(map[string]string),
	}
	session.ctx, session.finish = context.WithCancel(ctx)

	go session.KeepReading()
	return session
}

type ClientUDPSession struct {
	locker *sync.RWMutex

	conn            io.ReadWriteCloser
	packetProcessor UDPClientPacketProcessor
	sessionMap      map[string]*ClientUDPSessionConn

	sessionMapAlias map[string]string

	ctx    context.Context
	finish func()
}

func (c *ClientUDPSession) GetCachedState(sessionID string) UDPClientPacketProcessorCachedState {
	c.locker.RLock()
	defer c.locker.RUnlock()

	state, ok := c.sessionMap[sessionID]
	if !ok {
		return nil
	}
	return state.cachedProcessorState
}

func (c *ClientUDPSession) GetCachedServerState(serverSessionID string) UDPClientPacketProcessorCachedState {
	c.locker.RLock()
	defer c.locker.RUnlock()

	clientSessionID := c.getCachedStateAlias(serverSessionID)
	if clientSessionID == "" {
		return nil
	}
	state, ok := c.sessionMap[clientSessionID]
	if !ok {
		return nil
	}

	if serverState, ok := state.trackedServerSessionID[serverSessionID]; !ok {
		return nil
	} else {
		return serverState.cachedRecvProcessorState
	}
}

func (c *ClientUDPSession) getCachedStateAlias(serverSessionID string) string {
	state, ok := c.sessionMapAlias[serverSessionID]
	if !ok {
		return ""
	}
	return state
}

func (c *ClientUDPSession) PutCachedState(sessionID string, cache UDPClientPacketProcessorCachedState) {
	c.locker.RLock()
	defer c.locker.RUnlock()

	state, ok := c.sessionMap[sessionID]
	if !ok {
		return
	}
	state.cachedProcessorState = cache
}

func (c *ClientUDPSession) PutCachedServerState(serverSessionID string, cache UDPClientPacketProcessorCachedState) {
	c.locker.RLock()
	defer c.locker.RUnlock()

	clientSessionID := c.getCachedStateAlias(serverSessionID)
	if clientSessionID == "" {
		return
	}
	state, ok := c.sessionMap[clientSessionID]
	if !ok {
		return
	}

	if serverState, ok := state.trackedServerSessionID[serverSessionID]; ok {
		serverState.cachedRecvProcessorState = cache
		return
	}
}

func (c *ClientUDPSession) Close() error {
	c.finish()
	return c.conn.Close()
}

func (c *ClientUDPSession) WriteUDPRequest(request *UDPRequest) error {
	buffer := buf.New()
	defer buffer.Release()
	err := c.packetProcessor.EncodeUDPRequest(request, buffer, c)
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
			err := c.packetProcessor.DecodeUDPResp(buffer[:n], udpResp, c)
			if err != nil {
				newError("unable to decode udp response").Base(err).WriteToLog()
				continue
			}

			{
				timeDifference := int64(udpResp.TimeStamp) - time.Now().Unix()
				if timeDifference < -30 || timeDifference > 30 {
					newError("udp packet timestamp difference too large, packet discarded, time diff = ", timeDifference).WriteToLog()
					continue
				}
			}

			c.locker.RLock()
			session, ok := c.sessionMap[string(udpResp.ClientSessionID[:])]
			c.locker.RUnlock()
			if ok {
				select {
				case session.readChan <- udpResp:
				default:
				}
			} else {
				newError("misbehaving server: unknown client session ID").Base(err).WriteToLog()
			}
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
		sessionID:              string(sessionID),
		readChan:               make(chan *UDPResponse, 128),
		parent:                 c,
		ctx:                    connctx,
		finish:                 connfinish,
		nextWritePacketID:      0,
		trackedServerSessionID: make(map[string]*ClientUDPSessionServerTracker),
	}
	c.locker.Lock()
	c.sessionMap[sessionConn.sessionID] = sessionConn
	c.locker.Unlock()
	return sessionConn, nil
}

type ClientUDPSessionServerTracker struct {
	cachedRecvProcessorState UDPClientPacketProcessorCachedState
	rxReplayDetector         replaydetector.ReplayDetector
	lastSeen                 time.Time
}

type ClientUDPSessionConn struct {
	sessionID string
	readChan  chan *UDPResponse
	parent    *ClientUDPSession

	nextWritePacketID      uint64
	trackedServerSessionID map[string]*ClientUDPSessionServerTracker

	cachedProcessorState UDPClientPacketProcessorCachedState

	ctx    context.Context
	finish func()
}

func (c *ClientUDPSessionConn) Close() error {
	c.parent.locker.Lock()
	delete(c.parent.sessionMap, c.sessionID)
	for k := range c.trackedServerSessionID {
		delete(c.parent.sessionMapAlias, k)
	}
	c.parent.locker.Unlock()
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
	for {
		select {
		case <-c.ctx.Done():
			return 0, nil, io.EOF
		case resp := <-c.readChan:
			n = copy(p, resp.Payload.Bytes())
			resp.Payload.Release()

			var trackedState *ClientUDPSessionServerTracker
			if trackedStateReceived, ok := c.trackedServerSessionID[string(resp.SessionID[:])]; !ok {
				for key, value := range c.trackedServerSessionID {
					if time.Since(value.lastSeen) > 65*time.Second {
						delete(c.trackedServerSessionID, key)
					}
				}

				state := &ClientUDPSessionServerTracker{
					rxReplayDetector: replaydetector.New(1024, ^uint64(0)),
				}
				c.trackedServerSessionID[string(resp.SessionID[:])] = state
				c.parent.locker.RLock()
				c.parent.sessionMapAlias[string(resp.SessionID[:])] = string(resp.ClientSessionID[:])
				c.parent.locker.RUnlock()
				trackedState = state
			} else {
				trackedState = trackedStateReceived
			}

			if accept, ok := trackedState.rxReplayDetector.Check(resp.PacketID); ok {
				accept()
			} else {
				newError("misbehaving server: replayed packet").Base(err).WriteToLog()
				continue
			}
			trackedState.lastSeen = time.Now()

			addr = &net.UDPAddr{IP: resp.Address.IP(), Port: resp.Port}
		}
		return n, addr, nil
	}
}
