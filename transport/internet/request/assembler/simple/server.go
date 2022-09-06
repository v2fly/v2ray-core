package simple

import (
	"bytes"
	"context"
	"sync"

	"github.com/v2fly/v2ray-core/v5/common"

	"github.com/v2fly/v2ray-core/v5/transport/internet/request"
)

func newServer(config *ServerConfig) request.SessionAssemblerServer {
	return &simpleAssemblerServer{}
}

type simpleAssemblerServer struct {
	sessions sync.Map
	assembly request.TransportServerAssembly
}

func (s *simpleAssemblerServer) OnTransportServerAssemblyReady(assembly request.TransportServerAssembly) {
	s.assembly = assembly
}

func (s *simpleAssemblerServer) OnRoundTrip(ctx context.Context, req request.Request, opts ...request.RoundTripperOption,
) (resp request.Response, err error) {
	connectionID := req.ConnectionTag
	session := newSimpleAssemblerServerSession(ctx)
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

func newSimpleAssemblerServerSession(ctx context.Context) *simpleAssemblerServerSession {
	sessionCtx, finish := context.WithCancel(ctx)
	return &simpleAssemblerServerSession{
		readBuffer:       bytes.NewBuffer(nil),
		readChan:         make(chan []byte, 16),
		requestProcessed: make(chan struct{}),
		writeLock:        new(sync.Mutex),
		writeBuffer:      bytes.NewBuffer(nil),
		maxWriteSize:     4096,
		ctx:              sessionCtx,
		finish:           finish,
	}
}

type simpleAssemblerServerSession struct {
	maxWriteSize int

	readBuffer       *bytes.Buffer
	readChan         chan []byte
	requestProcessed chan struct{}

	writeLock   *sync.Mutex
	writeBuffer *bytes.Buffer

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

	n, err = s.writeBuffer.Write(p)
	length := s.writeBuffer.Len()
	s.writeLock.Unlock()
	if err != nil {
		return 0, err
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
	if req.Data != nil && len(req.Data) > 0 {
		select {
		case <-s.ctx.Done():
			return request.Response{}, s.ctx.Err()
		case s.readChan <- req.Data:
		}
	}

	s.writeLock.Lock()
	nextWrite := s.writeBuffer.Next(s.maxWriteSize)
	data := make([]byte, len(nextWrite))
	copy(data, nextWrite)
	s.writeLock.Unlock()
	select {
	case s.requestProcessed <- struct{}{}:
	case <-s.ctx.Done():
		return request.Response{}, s.ctx.Err()
	default:
	}
	return request.Response{Data: data}, nil
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
