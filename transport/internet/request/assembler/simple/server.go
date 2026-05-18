package simple

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/v2fly/v2ray-core/v5/common"

	"github.com/v2fly/v2ray-core/v5/transport/internet/request"
)

func newServer(config *ServerConfig) request.SessionAssemblerServer {
	return &simpleAssemblerServer{config: config}
}

type simpleAssemblerServer struct {
	sessions sync.Map
	assembly request.TransportServerAssembly
	config   *ServerConfig
}

func (s *simpleAssemblerServer) OnTransportServerAssemblyReady(assembly request.TransportServerAssembly) {
	s.assembly = assembly
}

func (s *simpleAssemblerServer) OnRoundTrip(ctx context.Context, req request.Request, opts ...request.RoundTripperOption,
) (resp request.Response, err error) {
	connectionID := req.ConnectionTag
	session := newSimpleAssemblerServerSession(ctx, serverMaxWriteSize(s.config), serverPollingResponseWait(s.config))
	loadedSession, loaded := s.sessions.LoadOrStore(string(connectionID), session)
	if loaded {
		session = loadedSession.(*simpleAssemblerServerSession)
	} else {
		if err := s.assembly.SessionReceiver().OnNewSession(ctx, session); err != nil {
			return request.Response{}, newError("failed to create new session").Base(err)
		}
	}
	return session.OnRoundTrip(ctx, req, opts...)
}

func newSimpleAssemblerServerSession(ctx context.Context, maxWriteSize int, pollingResponseWait time.Duration) *simpleAssemblerServerSession {
	sessionCtx, finish := context.WithCancel(ctx)
	return &simpleAssemblerServerSession{
		readBuffer:       bytes.NewBuffer(nil),
		readChan:         make(chan []byte, 16),
		requestProcessed: make(chan struct{}),
		writeLock:        new(sync.Mutex),
		writeBuffer:      bytes.NewBuffer(nil),
		writeAvailable:   make(chan struct{}, 1),
		maxWriteSize:     maxWriteSize,
		pollingWait:      pollingResponseWait,
		ctx:              sessionCtx,
		finish:           finish,
	}
}

type simpleAssemblerServerSession struct {
	maxWriteSize int
	pollingWait  time.Duration

	readBuffer       *bytes.Buffer
	readChan         chan []byte
	requestProcessed chan struct{}

	writeLock      *sync.Mutex
	writeBuffer    *bytes.Buffer
	writeAvailable chan struct{}

	ctx    context.Context
	finish func()
}

func (s *simpleAssemblerServerSession) Read(p []byte) (n int, err error) {
	if s.readBuffer.Len() == 0 {
		select {
		case <-s.ctx.Done():
			return 0, s.ctx.Err()
		case data := <-s.readChan:
			s.readBuffer.Write(data)
		}
	}
	return s.readBuffer.Read(p)
}

func (s *simpleAssemblerServerSession) Write(p []byte) (n int, err error) {
	s.writeLock.Lock()

	wasEmpty := s.writeBuffer.Len() == 0
	n, err = s.writeBuffer.Write(p)
	length := s.writeBuffer.Len()
	s.writeLock.Unlock()
	if err != nil {
		return 0, err
	}
	if n > 0 && wasEmpty {
		s.signalWriteAvailable()
	}
	if length > s.maxWriteSize {
		select {
		case <-s.requestProcessed:
		case <-s.ctx.Done():
			return 0, s.ctx.Err()
		}
	}
	return
}

func (s *simpleAssemblerServerSession) Close() error {
	s.finish()
	return nil
}

func (s *simpleAssemblerServerSession) OnRoundTrip(ctx context.Context, req request.Request, opts ...request.RoundTripperOption,
) (resp request.Response, err error) {
	if len(req.Data) > 0 {
		select {
		case <-s.ctx.Done():
			return request.Response{}, s.ctx.Err()
		case s.readChan <- req.Data:
		}
	}

	data := s.nextWriteData()
	if len(data) == 0 && len(req.Data) == 0 && s.pollingWait > 0 {
		data, err = s.waitForWriteData(ctx)
		if err != nil {
			return request.Response{}, err
		}
	}
	if err := s.signalRequestProcessed(); err != nil {
		return request.Response{}, err
	}
	return request.Response{Data: data}, nil
}

func (s *simpleAssemblerServerSession) nextWriteData() []byte {
	s.writeLock.Lock()
	nextWrite := s.writeBuffer.Next(s.maxWriteSize)
	data := make([]byte, len(nextWrite))
	copy(data, nextWrite)
	s.writeLock.Unlock()
	return data
}

func (s *simpleAssemblerServerSession) waitForWriteData(ctx context.Context) ([]byte, error) {
	timer := time.NewTimer(s.pollingWait)
	defer timer.Stop()
	for {
		select {
		case <-s.writeAvailable:
			data := s.nextWriteData()
			if len(data) != 0 {
				return data, nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-s.ctx.Done():
			return nil, s.ctx.Err()
		case <-timer.C:
			return nil, nil
		}
	}
}

func (s *simpleAssemblerServerSession) signalWriteAvailable() {
	select {
	case s.writeAvailable <- struct{}{}:
	default:
	}
}

func (s *simpleAssemblerServerSession) signalRequestProcessed() error {
	select {
	case s.requestProcessed <- struct{}{}:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
		return nil
	}
}

func serverMaxWriteSize(config *ServerConfig) int {
	if config != nil && config.MaxWriteSize > 0 {
		return int(config.MaxWriteSize)
	}
	return 4096
}

func serverPollingResponseWait(config *ServerConfig) time.Duration {
	if config != nil && config.PollingResponseWaitMs > 0 {
		return time.Duration(config.PollingResponseWaitMs) * time.Millisecond
	}
	return 0
}

func init() {
	common.Must(common.RegisterConfig((*ServerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		serverConfig, ok := config.(*ServerConfig)
		if !ok {
			return nil, newError("not a SimpleServerConfig")
		}
		return newServer(serverConfig), nil
	}))
}
