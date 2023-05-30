package simple

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"time"

	"github.com/v2fly/v2ray-core/v5/common"

	"github.com/v2fly/v2ray-core/v5/transport/internet/request"
)

func newClient(config *ClientConfig) request.SessionAssemblerClient {
	return &simpleAssemblerClient{config: config}
}

type simpleAssemblerClient struct {
	assembly request.TransportClientAssembly
	config   *ClientConfig
}

func (s *simpleAssemblerClient) OnTransportClientAssemblyReady(assembly request.TransportClientAssembly) {
	s.assembly = assembly
}

func (s *simpleAssemblerClient) NewSession(ctx context.Context, opts ...request.SessionOption) (request.Session, error) {
	sessionID := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, sessionID)
	if err != nil {
		return nil, err
	}
	sessionContext, finish := context.WithCancel(ctx)
	session := &simpleAssemblerClientSession{
		sessionID: sessionID, tripper: s.assembly.Tripper(), readBuffer: bytes.NewBuffer(nil),
		ctx: sessionContext, finish: finish, writerChan: make(chan []byte), readerChan: make(chan []byte, 16), assembler: s,
	}
	go session.keepRunning()
	return session, nil
}

type simpleAssemblerClientSession struct {
	sessionID        []byte
	currentWriteWait int

	assembler  *simpleAssemblerClient
	tripper    request.Tripper
	readBuffer *bytes.Buffer
	writerChan chan []byte
	readerChan chan []byte
	ctx        context.Context
	finish     func()
}

func (s *simpleAssemblerClientSession) keepRunning() {
	s.currentWriteWait = int(s.assembler.config.InitialPollingIntervalMs)
	for s.ctx.Err() == nil {
		s.runOnce()
	}
}

func (s *simpleAssemblerClientSession) runOnce() {
	sendBuffer := bytes.NewBuffer(nil)
	if s.currentWriteWait != 0 {
		waitTimer := time.NewTimer(time.Millisecond * time.Duration(s.currentWriteWait))
		waitForFirstWrite := true
	copyFromWriterLoop:
		for {
			select {
			case <-s.ctx.Done():
				return
			case data := <-s.writerChan:
				sendBuffer.Write(data)
				if sendBuffer.Len() >= int(s.assembler.config.MaxWriteSize) {
					break copyFromWriterLoop
				}
				if waitForFirstWrite {
					waitForFirstWrite = false
					waitTimer.Reset(time.Millisecond * time.Duration(s.assembler.config.WaitSubsequentWriteMs))
				}
			case <-waitTimer.C:
				break copyFromWriterLoop
			}
		}
		waitTimer.Stop()
	}

	firstRound := true
	pollConnection := true
	for sendBuffer.Len() != 0 || firstRound {
		firstRound = false
		sendAmount := sendBuffer.Len()
		if sendAmount > int(s.assembler.config.MaxWriteSize) {
			sendAmount = int(s.assembler.config.MaxWriteSize)
		}
		data := sendBuffer.Next(sendAmount)
		if len(data) != 0 {
			pollConnection = false
		}
		for {
			resp, err := s.tripper.RoundTrip(s.ctx, request.Request{Data: data, ConnectionTag: s.sessionID})
			if err != nil {
				newError("failed to send data").Base(err).WriteToLog()
				if s.ctx.Err() != nil {
					return
				}
				time.Sleep(time.Millisecond * time.Duration(s.assembler.config.FailedRetryIntervalMs))
				continue
			}
			if len(resp.Data) != 0 {
				s.readerChan <- resp.Data
			}
			if len(resp.Data) != 0 {
				pollConnection = false
			}
			break
		}
	}
	if pollConnection {
		s.currentWriteWait = int(s.assembler.config.BackoffFactor * float32(s.currentWriteWait))
		if s.currentWriteWait > int(s.assembler.config.MaxPollingIntervalMs) {
			s.currentWriteWait = int(s.assembler.config.MaxPollingIntervalMs)
		}
		if s.currentWriteWait < int(s.assembler.config.MinPollingIntervalMs) {
			s.currentWriteWait = int(s.assembler.config.MinPollingIntervalMs)
		}
	} else {
		s.currentWriteWait = int(0)
	}
}

func (s *simpleAssemblerClientSession) Read(p []byte) (n int, err error) {
	if s.readBuffer.Len() == 0 {
		select {
		case <-s.ctx.Done():
			return 0, s.ctx.Err()
		case data := <-s.readerChan:
			s.readBuffer.Write(data)
		}
	}
	n, err = s.readBuffer.Read(p)
	if err == io.EOF {
		s.readBuffer.Reset()
		return 0, nil
	}
	return
}

func (s *simpleAssemblerClientSession) Write(p []byte) (n int, err error) {
	buf := make([]byte, len(p))
	copy(buf, p)
	select {
	case <-s.ctx.Done():
		return 0, s.ctx.Err()
	case s.writerChan <- buf:
		return len(p), nil
	}
}

func (s *simpleAssemblerClientSession) Close() error {
	s.finish()
	return nil
}

func init() {
	common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		clientConfig, ok := config.(*ClientConfig)
		if !ok {
			return nil, newError("not a ClientConfig")
		}
		return newClient(clientConfig), nil
	}))
}
