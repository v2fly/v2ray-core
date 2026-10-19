package reverse

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/mux"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/features/outbound"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/pipe"
)

type Portal struct {
	ctx    context.Context
	ohm    outbound.Manager
	tag    string
	domain string
	picker *StaticMuxPicker
	client *mux.ClientManager
}

func NewPortal(ctx context.Context, config *PortalConfig, ohm outbound.Manager) (*Portal, error) {
	if config.Tag == "" {
		return nil, newError("portal tag is empty")
	}

	if config.Domain == "" {
		return nil, newError("portal domain is empty")
	}

	picker, err := NewStaticMuxPicker()
	if err != nil {
		return nil, err
	}

	return &Portal{
		ctx:    ctx,
		ohm:    ohm,
		tag:    config.Tag,
		domain: config.Domain,
		picker: picker,
		client: &mux.ClientManager{
			Picker: picker,
		},
	}, nil
}

func (p *Portal) Start() error {
	return p.ohm.AddHandler(p.ctx, &Outbound{
		portal: p,
		tag:    p.tag,
	})
}

func (p *Portal) Close() error {
	return p.ohm.RemoveHandler(p.ctx, p.tag)
}

func (p *Portal) HandleConnection(ctx context.Context, link *transport.Link) error {
	outboundMeta := session.OutboundFromContext(ctx)
	if outboundMeta == nil {
		return newError("outbound metadata not found").AtError()
	}

	if isDomain(outboundMeta.Target, p.domain) {
		muxClient, err := mux.NewClientWorker(*link, mux.ClientStrategy{})
		if err != nil {
			return newError("failed to create mux client worker").Base(err).AtWarning()
		}

		worker, err := NewPortalWorker(ctx, muxClient)
		if err != nil {
			return newError("failed to create portal worker").Base(err)
		}

		p.picker.AddWorker(worker)
		return nil
	}

	return p.client.Dispatch(ctx, link)
}

type Outbound struct {
	portal *Portal
	tag    string
}

func (o *Outbound) Tag() string {
	return o.tag
}

func (o *Outbound) Dispatch(ctx context.Context, link *transport.Link) {
	if err := o.portal.HandleConnection(ctx, link); err != nil {
		newError("failed to process reverse connection").Base(err).WriteToLog(session.ExportIDToError(ctx))
		common.Interrupt(link.Writer)
	}
}

func (o *Outbound) Start() error {
	return nil
}

func (o *Outbound) Close() error {
	return nil
}

type StaticMuxPicker struct {
	access  sync.Mutex
	workers []*PortalWorker
	cTask   *task.Periodic
}

func NewStaticMuxPicker() (*StaticMuxPicker, error) {
	p := &StaticMuxPicker{}
	p.cTask = &task.Periodic{
		Execute:  p.cleanup,
		Interval: time.Second * 30,
	}
	p.cTask.Start()
	return p, nil
}

func (p *StaticMuxPicker) cleanup() error {
	p.access.Lock()
	defer p.access.Unlock()

	var activeWorkers []*PortalWorker
	for _, w := range p.workers {
		if !w.Closed() {
			activeWorkers = append(activeWorkers, w)
		}
	}

	if len(activeWorkers) != len(p.workers) {
		p.workers = activeWorkers
	}

	return nil
}

func (p *StaticMuxPicker) PickAvailable() (*mux.ClientWorker, error) {
	p.access.Lock()
	defer p.access.Unlock()

	if len(p.workers) == 0 {
		return nil, newError("empty worker list")
	}

	minIdx := -1
	var minConn uint32 = 9999
	for i, w := range p.workers {
		if w.state.Load() != portalWorkerActive {
			continue
		}
		if w.client.Closed() {
			continue
		}
		if w.client.ActiveConnections() < minConn {
			minConn = w.client.ActiveConnections()
			minIdx = i
		}
	}

	if minIdx == -1 {
		for i, w := range p.workers {
			if w.state.Load() == portalWorkerFailed || w.IsFull() {
				continue
			}
			if w.client.ActiveConnections() < minConn {
				minConn = w.client.ActiveConnections()
				minIdx = i
			}
		}
	}

	if minIdx != -1 {
		return p.workers[minIdx].client, nil
	}

	return nil, newError("no mux client worker available")
}

func (p *StaticMuxPicker) AddWorker(worker *PortalWorker) {
	p.access.Lock()
	defer p.access.Unlock()

	p.workers = append(p.workers, worker)
}

const (
	controlHeartbeatInterval = 2 * time.Second
	controlSilenceTimeout    = 10 * time.Second
)

const (
	portalWorkerActive   int32 = iota
	portalWorkerDraining       // Graceful rollover: still usable until a replacement arrives.
	portalWorkerFailed         // Health timeout: never select again.
)

type heartbeatState struct {
	probeID     atomic.Uint64
	mu          sync.Mutex // Guards ACK data and the expiration decision.
	lastAckID   uint64
	lastAckTime time.Time
}

type PortalWorker struct {
	client    *mux.ClientWorker
	control   *task.Periodic
	writer    buf.Writer
	reader    buf.Reader
	state     atomic.Int32
	heartbeat heartbeatState
}

func NewPortalWorker(ctx context.Context, client *mux.ClientWorker) (*PortalWorker, error) {
	opt := []pipe.Option{pipe.WithSizeLimit(16 * 1024)}
	uplinkReader, uplinkWriter := pipe.New(opt...)
	downlinkReader, downlinkWriter := pipe.New(opt...)

	ctx = session.ContextWithOutbound(ctx, &session.Outbound{
		Target: net.UDPDestination(net.DomainAddress(internalDomain), 0),
	})
	f := client.Dispatch(ctx, &transport.Link{
		Reader: uplinkReader,
		Writer: downlinkWriter,
	})
	if !f {
		return nil, newError("unable to dispatch control connection")
	}
	w := &PortalWorker{
		client: client,
		reader: downlinkReader,
		writer: uplinkWriter,
	}
	w.control = &task.Periodic{
		Execute:  w.sendHeartbeat,
		Interval: controlHeartbeatInterval,
	}
	go w.receiveAcks()
	w.control.Start()
	go w.watchAcks()
	return w, nil
}

func (w *PortalWorker) sendHeartbeat() error {
	if w.client.Closed() {
		return newError("client worker stopped")
	}

	if w.state.Load() != portalWorkerActive {
		return newError("already disposed")
	}

	msg := &Control{HeartbeatId: w.heartbeat.probeID.Add(1)}
	msg.FillInRandom()

	if w.client.TotalConnections() > 256 {
		if !w.state.CompareAndSwap(portalWorkerActive, portalWorkerDraining) {
			return newError("already disposed")
		}
		msg.State = Control_DRAIN

		defer func() {
			common.Close(w.writer)
			common.Interrupt(w.reader)
		}()
	}

	b, err := proto.Marshal(msg)
	common.Must(err)
	mb := buf.MergeBytes(nil, b)
	return w.writer.WriteMultiBuffer(mb)
}

func (w *PortalWorker) receiveAcks() {
	for {
		mb, err := w.reader.ReadMultiBuffer()
		for _, b := range mb {
			var ack Control
			if proto.Unmarshal(b.Bytes(), &ack) == nil {
				w.recordAck(ack.HeartbeatId, time.Now())
			}
		}
		buf.ReleaseMulti(mb)
		if err != nil {
			return
		}
	}
}

func (w *PortalWorker) recordAck(id uint64, now time.Time) {
	if id == 0 || id > w.heartbeat.probeID.Load() || w.state.Load() != portalWorkerActive {
		return
	}
	w.heartbeat.mu.Lock()
	if id > w.heartbeat.lastAckID {
		w.heartbeat.lastAckID = id
		w.heartbeat.lastAckTime = now
	}
	w.heartbeat.mu.Unlock()
}

func (w *PortalWorker) watchAcks() {
	ticker := time.NewTicker(controlHeartbeatInterval)
	defer ticker.Stop()
	for range ticker.C {
		if w.client.Closed() || w.state.Load() != portalWorkerActive || w.expireIfUnhealthy(time.Now()) {
			return
		}
	}
}

func (w *PortalWorker) expireIfUnhealthy(now time.Time) bool {
	if w.client.Closed() {
		return false
	}
	w.heartbeat.mu.Lock()
	lastAck := w.heartbeat.lastAckTime
	// Missing the first reply may mean an older bridge, so only confirmed
	// workers are subject to the acknowledgement deadline.
	if lastAck.IsZero() || now.Sub(lastAck) < controlSilenceTimeout || !w.state.CompareAndSwap(portalWorkerActive, portalWorkerFailed) {
		w.heartbeat.mu.Unlock()
		return false
	}
	w.heartbeat.mu.Unlock()

	common.Must(w.client.Close()) // Closed() must change before the next picker call.
	newError("portal worker control acknowledgement timed out").AtWarning().WriteToLog()
	common.Must(w.control.Close())
	common.Interrupt(w.writer)
	common.Interrupt(w.reader)
	return true
}

func (w *PortalWorker) IsFull() bool {
	return w.client.IsFull()
}

func (w *PortalWorker) Closed() bool {
	return w.client.Closed()
}
