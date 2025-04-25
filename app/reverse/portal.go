package reverse

import (
	"context"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/buf"
	"github.com/ghxhy/v2ray-core/v5/common/mux"
	"github.com/ghxhy/v2ray-core/v5/common/net"
	"github.com/ghxhy/v2ray-core/v5/common/session"
	"github.com/ghxhy/v2ray-core/v5/common/task"
	"github.com/ghxhy/v2ray-core/v5/features/outbound"
	"github.com/ghxhy/v2ray-core/v5/transport"
	"github.com/ghxhy/v2ray-core/v5/transport/pipe"
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
		if w.draining {
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
			if w.IsFull() {
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

type PortalWorker struct {
	client   *mux.ClientWorker
	control  *task.Periodic
	writer   buf.Writer
	reader   buf.Reader
	draining bool
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
		Execute:  w.heartbeat,
		Interval: time.Second * 2,
	}
	w.control.Start()
	return w, nil
}

func (w *PortalWorker) heartbeat() error {
	if w.client.Closed() {
		return newError("client worker stopped")
	}

	if w.draining || w.writer == nil {
		return newError("already disposed")
	}

	msg := &Control{}
	msg.FillInRandom()

	if w.client.TotalConnections() > 256 {
		w.draining = true
		msg.State = Control_DRAIN

		defer func() {
			common.Close(w.writer)
			common.Interrupt(w.reader)
			w.writer = nil
		}()
	}

	b, err := proto.Marshal(msg)
	common.Must(err)
	mb := buf.MergeBytes(nil, b)
	return w.writer.WriteMultiBuffer(mb)
}

func (w *PortalWorker) IsFull() bool {
	return w.client.IsFull()
}

func (w *PortalWorker) Closed() bool {
	return w.client.Closed()
}
