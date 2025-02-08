package packetconn

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"sync"
	"time"

	"github.com/golang-collections/go-datastructures/queue"

	"github.com/v2fly/v2ray-core/v5/transport/internet/request"
)

func newRequestToPacketConnClient(ctx context.Context, config *ClientConfig) (*requestToPacketConnClient, error) { //nolint: unparam
	return &requestToPacketConnClient{ctx: ctx, config: config}, nil
}

type requestToPacketConnClient struct {
	assembly request.TransportClientAssembly
	ctx      context.Context
	config   *ClientConfig
}

func (r *requestToPacketConnClient) OnTransportClientAssemblyReady(assembly request.TransportClientAssembly) {
	r.assembly = assembly
}

func (r *requestToPacketConnClient) Dial() (io.ReadWriteCloser, error) {
	sessionID := make([]byte, 16)
	_, err := rand.Read(sessionID)
	if err != nil {
		return nil, err
	}
	ctxWithCancel, cancel := context.WithCancel(r.ctx)

	clientSess := &requestToPacketConnClientSession{
		sessionID:              sessionID,
		currentPollingInterval: int(r.config.PollingIntervalInitial),
		maxRequestSize:         int(r.config.MaxRequestSize),
		maxWriteDelay:          int(r.config.MaxWriteDelay),
		assembly:               r.assembly,
		writerChan:             make(chan []byte, 256),
		readerChan:             make(chan []byte, 256),
		ctx:                    ctxWithCancel,
		finish:                 cancel,
	}
	go clientSess.keepRunning()
	return clientSess, nil
}

type requestToPacketConnClientSession struct {
	sessionID              []byte
	currentPollingInterval int

	maxRequestSize int
	maxWriteDelay  int

	assembly   request.TransportClientAssembly
	writerChan chan []byte
	readerChan chan []byte
	ctx        context.Context
	finish     func()
	nextWrite  []byte
}

func (r *requestToPacketConnClientSession) keepRunning() {
	for r.ctx.Err() == nil {
		r.runOnce()
	}
}

func (r *requestToPacketConnClientSession) runOnce() {
	requestBody := bytes.NewBuffer(nil)
	waitTimer := time.NewTimer(time.Duration(r.currentPollingInterval) * time.Millisecond)
	var seenPacket bool
	packetBundler := NewPacketBundle()
copyFromChan:
	for {
		select {
		case <-r.ctx.Done():
			return
		case <-waitTimer.C:
			break copyFromChan
		case packet := <-r.writerChan:
			if !seenPacket {
				seenPacket = true
				waitTimer.Stop()
				waitTimer.Reset(time.Duration(r.maxWriteDelay) * time.Millisecond)
			}
			sizeOffset := packetBundler.Overhead() + len(packet)
			if requestBody.Len()+sizeOffset > r.maxRequestSize {
				r.nextWrite = packet
				break copyFromChan
			}
			err := packetBundler.WriteToBundle(packet, requestBody)
			if err != nil {
				newError("failed to write to bundle").Base(err).WriteToLog()
			}
		}
	}
	waitTimer.Stop()
	go func() {
		reader, writer := io.Pipe()
		defer writer.Close()
		streamingRespOpt := &pipedStreamingRespOption{writer}
		go func() {
			for {
				if packet, err := packetBundler.ReadFromBundle(reader); err == nil {
					r.readerChan <- packet
				} else {
					return
				}
			}
		}()
		resp, err := r.assembly.Tripper().RoundTrip(r.ctx, request.Request{Data: requestBody.Bytes(), ConnectionTag: r.sessionID},
			streamingRespOpt)
		if err != nil {
			newError("failed to roundtrip").Base(err).WriteToLog()
			if r.ctx.Err() != nil {
				return
			}
		}
		if len(resp.Data) != 0 {
			respReader := bytes.NewReader(resp.Data)
			for respReader.Len() != 0 {
				packet, err := packetBundler.ReadFromBundle(respReader)
				if err != nil {
					newError("failed to read from bundle").Base(err).WriteToLog()
					if r.ctx.Err() != nil {
						return
					}
				}
				r.readerChan <- packet
			}
		}
	}()
}

type pipedStreamingRespOption struct {
	writer *io.PipeWriter
}

func (p *pipedStreamingRespOption) RoundTripperOption() {
}

func (p *pipedStreamingRespOption) GetResponseWriter() io.Writer {
	return p.writer
}

func (r *requestToPacketConnClientSession) Write(p []byte) (n int, err error) {
	buf := make([]byte, len(p))
	copy(buf, p)
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	case r.writerChan <- buf:
		return len(p), nil
	}
}

func (r *requestToPacketConnClientSession) Read(p []byte) (n int, err error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	case buf := <-r.readerChan:
		copy(p, buf)
		return len(buf), nil
	}
}

func (r *requestToPacketConnClientSession) Close() error {
	r.finish()
	return nil
}

func newRequestToPacketConnServer(ctx context.Context, config *ServerConfig) *requestToPacketConnServer {
	return &requestToPacketConnServer{
		sessionMap: sync.Map{},
		ctx:        ctx,
		config:     config,
	}
}

type requestToPacketConnServer struct {
	packetSessionReceiver request.SessionReceiver

	sessionMap sync.Map

	ctx    context.Context
	config *ServerConfig
}

func (r *requestToPacketConnServer) onSessionReceiverReady(sessrecv request.SessionReceiver) {
	r.packetSessionReceiver = sessrecv
}

func (r *requestToPacketConnServer) OnRoundTrip(ctx context.Context, req request.Request,
	opts ...request.RoundTripperOption,
) (resp request.Response, err error) {
	SessionID := req.ConnectionTag
	if SessionID == nil {
		return request.Response{}, newError("nil session id")
	}
	sessionID := string(SessionID)
	var session *requestToPacketConnServerSession
	sessionAny, found := r.sessionMap.Load(sessionID)
	if found {
		var ok bool
		session, ok = sessionAny.(*requestToPacketConnServerSession)
		if !ok {
			return request.Response{}, newError("failed to cast session")
		}
	}
	if !found {
		ctxWithFinish, finish := context.WithCancel(ctx)
		session = &requestToPacketConnServerSession{
			SessionID:                      SessionID,
			writingConnectionQueue:         queue.New(64),
			writerChan:                     make(chan []byte, int(r.config.PacketWritingBuffer)),
			readerChan:                     make(chan []byte, 256),
			ctx:                            ctxWithFinish,
			finish:                         finish,
			server:                         r,
			maxWriteSize:                   int(r.config.MaxWriteSize),
			maxWriteDuration:               int(r.config.MaxWriteDurationMs),
			maxSimultaneousWriteConnection: int(r.config.MaxSimultaneousWriteConnection),
		}
		_, loaded := r.sessionMap.LoadOrStore(sessionID, session)
		if !loaded {
			err = r.packetSessionReceiver.OnNewSession(ctx, session)
		}
	}
	if err != nil {
		return request.Response{}, err
	}
	return session.OnRoundTrip(ctx, req, opts...)
}

func (r *requestToPacketConnServer) removeSessionID(sessionID []byte) {
	r.sessionMap.Delete(string(sessionID))
}

type requestToPacketConnServerSession struct {
	SessionID []byte

	writingConnectionQueue *queue.Queue

	writerChan chan []byte
	readerChan chan []byte
	ctx        context.Context
	finish     func()
	server     *requestToPacketConnServer

	maxWriteSize                   int
	maxWriteDuration               int
	maxSimultaneousWriteConnection int
}

func (r *requestToPacketConnServerSession) Read(p []byte) (n int, err error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	case buf := <-r.readerChan:
		copy(p, buf)
		return len(buf), nil
	}
}

var debugStats struct {
	packetWritten int
	packetDropped int
}

/*
var _ = func() bool {
	go func() {
		for {
			time.Sleep(time.Second)
			newError("packet written: ", debugStats.packetWritten, " packet dropped: ", debugStats.packetDropped).WriteToLog()
		}
	}()
	return true
}()*/

func (r *requestToPacketConnServerSession) Write(p []byte) (n int, err error) {
	buf := make([]byte, len(p))
	copy(buf, p)
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	case r.writerChan <- buf:
		debugStats.packetWritten++
		return len(p), nil
	default: // This write will be called from global listener's routine, it must not block
		debugStats.packetDropped++
		return len(p), nil
	}
}

func (r *requestToPacketConnServerSession) Close() error {
	r.server.removeSessionID(r.SessionID)
	r.finish()
	return nil
}

type writingConnection struct {
	focus     func()
	finish    func()
	finishCtx context.Context
}

func (r *requestToPacketConnServerSession) OnRoundTrip(ctx context.Context, req request.Request,
	opts ...request.RoundTripperOption,
) (resp request.Response, err error) {
	// TODO: fix connection graceful close
	var streamingRespWriter io.Writer
	var streamingRespWriterFlusher request.OptionSupportsStreamingResponseExtensionFlusher
	for _, opt := range opts {
		if streamingRespOpt, ok := opt.(request.OptionSupportsStreamingResponse); ok {
			streamingRespWriter = streamingRespOpt.GetResponseWriter()
			if streamingRespWriterFlusherOpt, ok := opt.(request.OptionSupportsStreamingResponseExtensionFlusher); ok {
				streamingRespWriterFlusher = streamingRespWriterFlusherOpt
			}
		}
	}
	packetBundler := NewPacketBundle()
	reqReader := bytes.NewReader(req.Data)
	for reqReader.Len() != 0 {
		packet, err := packetBundler.ReadFromBundle(reqReader)
		if err != nil {
			err = newError("failed to read from bundle").Base(err)
			return request.Response{}, err
		}
		r.readerChan <- packet
	}
	onFocusCtx, focus := context.WithCancel(ctx)
	onFinishCtx, finish := context.WithCancel(ctx)
	r.writingConnectionQueue.Put(&writingConnection{
		focus:     focus,
		finish:    finish,
		finishCtx: onFinishCtx,
	})

	amountToEnd := r.writingConnectionQueue.Len() - int64(r.maxSimultaneousWriteConnection)
	for amountToEnd > 0 {
		{
			_, _ = r.writingConnectionQueue.TakeUntil(func(i interface{}) bool {
				i.(*writingConnection).finish()
				amountToEnd--
				return amountToEnd > 0
			})
		}
	}

	{
		_, _ = r.writingConnectionQueue.TakeUntil(func(i interface{}) bool {
			i.(*writingConnection).focus()
			return false
		})
	}

	bufferedRespWriter := bytes.NewBuffer(nil)
	finishWrite := func() {
		resp.Data = bufferedRespWriter.Bytes()
		{
			_, _ = r.writingConnectionQueue.TakeUntil(func(i interface{}) bool {
				i.(*writingConnection).focus()
				if i.(*writingConnection).finishCtx.Err() != nil { //nolint: gosimple
					return true
				}
				return false
			})
		}
	}

	progressiveSend := streamingRespWriter != nil
	var respWriter io.Writer
	if progressiveSend {
		respWriter = streamingRespWriter
	} else {
		respWriter = bufferedRespWriter
	}

	var bytesSent int
	onReceivePacket := func(packet []byte) bool {
		bytesSent += len(packet) + packetBundler.Overhead()
		err := packetBundler.WriteToBundle(packet, respWriter)
		if err != nil {
			newError("failed to write to bundle").Base(err).WriteToLog()
		}
		if streamingRespWriterFlusher != nil {
			streamingRespWriterFlusher.Flush()
		}
		if bytesSent >= r.maxWriteSize {
			return false
		}
		return true
	}

	finishWriteTimer := time.NewTimer(time.Millisecond * time.Duration(r.maxWriteDuration))

	if !progressiveSend {
		select {
		case <-onFocusCtx.Done():
		case <-onFinishCtx.Done():
			finishWrite()
			return resp, nil
		}
	} else {
		select {
		case <-onFinishCtx.Done():
			finishWrite()
			return resp, nil
		default:
		}
	}
	firstRead := true
	for {
		select {
		case <-onFinishCtx.Done():
			finishWrite()
			finishWriteTimer.Stop()
			return resp, nil
		case packet := <-r.writerChan:
			keepSending := onReceivePacket(packet)
			if firstRead {
				firstRead = false
			}
			if !keepSending {
				finishWrite()
				finishWriteTimer.Stop()
				return resp, nil
			}
		case <-finishWriteTimer.C:
			finishWrite()
			return resp, nil
		}
	}
}
