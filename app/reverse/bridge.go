package reverse

import (
	"context"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/mux"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/pipe"
)

// Bridge is a component in reverse proxy, that relays connections from Portal to local address.
type Bridge struct {
	ctx         context.Context
	dispatcher  routing.Dispatcher
	tag         string
	domain      string
	workers     []*BridgeWorker
	monitorTask *task.Periodic
}

// NewBridge creates a new Bridge instance.
func NewBridge(ctx context.Context, config *BridgeConfig, dispatcher routing.Dispatcher) (*Bridge, error) {
	if config.Tag == "" {
		return nil, newError("bridge tag is empty")
	}
	if config.Domain == "" {
		return nil, newError("bridge domain is empty")
	}

	b := &Bridge{
		ctx:        ctx,
		dispatcher: dispatcher,
		tag:        config.Tag,
		domain:     config.Domain,
	}
	b.monitorTask = &task.Periodic{
		Execute:  b.monitor,
		Interval: controlHeartbeatInterval,
	}
	return b, nil
}

func (b *Bridge) cleanup() {
	var activeWorkers []*BridgeWorker
	now := time.Now()

	for _, w := range b.workers {
		w.expireIfUnhealthy(now)
		if w.IsActive() {
			activeWorkers = append(activeWorkers, w)
		}
	}

	if len(activeWorkers) != len(b.workers) {
		b.workers = activeWorkers
	}
}

func (b *Bridge) monitor() error {
	b.cleanup()

	var numConnections uint32
	var numWorker uint32

	for _, w := range b.workers {
		if w.IsActive() {
			numConnections += w.Connections()
			numWorker++
		}
	}

	if numWorker == 0 || numConnections/numWorker > 16 {
		worker, err := NewBridgeWorker(b.ctx, b.domain, b.tag, b.dispatcher)
		if err != nil {
			newError("failed to create bridge worker").Base(err).AtWarning().WriteToLog()
			return nil
		}
		b.workers = append(b.workers, worker)
	}

	return nil
}

func (b *Bridge) Start() error {
	return b.monitorTask.Start()
}

func (b *Bridge) Close() error {
	return b.monitorTask.Close()
}

type BridgeWorker struct {
	tag           string
	worker        *mux.ServerWorker
	dispatcher    routing.Dispatcher
	state         Control_State
	controlAccess sync.Mutex
	lastControl   time.Time
}

func NewBridgeWorker(ctx context.Context, domain string, tag string, d routing.Dispatcher) (*BridgeWorker, error) {
	bridgeCtx := session.ContextWithInbound(ctx, &session.Inbound{
		Tag: tag,
	})
	link, err := d.Dispatch(bridgeCtx, net.Destination{
		Network: net.Network_TCP,
		Address: net.DomainAddress(domain),
		Port:    0,
	})
	if err != nil {
		return nil, err
	}

	w := &BridgeWorker{
		dispatcher: d,
		tag:        tag,
	}

	worker, err := mux.NewServerWorker(ctx, w, link)
	if err != nil {
		return nil, err
	}
	w.worker = worker

	return w, nil
}

func (w *BridgeWorker) Type() interface{} {
	return routing.DispatcherType()
}

func (w *BridgeWorker) Start() error {
	return nil
}

func (w *BridgeWorker) Close() error {
	return nil
}

func (w *BridgeWorker) IsActive() bool {
	w.controlAccess.Lock()
	active := w.state == Control_ACTIVE
	w.controlAccess.Unlock()
	return active && !w.worker.Closed()
}

func (w *BridgeWorker) expireIfUnhealthy(now time.Time) bool {
	if w.worker.Closed() {
		return false
	}
	w.controlAccess.Lock()
	// Only newer portals send heartbeat IDs. Do not time out legacy portals.
	if w.state != Control_ACTIVE || w.lastControl.IsZero() || now.Sub(w.lastControl) < controlSilenceTimeout {
		w.controlAccess.Unlock()
		return false
	}
	w.state = Control_DRAIN
	w.controlAccess.Unlock()

	newError("bridge worker control heartbeat timed out").AtWarning().WriteToLog()
	go w.worker.Close()
	return true
}

func (w *BridgeWorker) Connections() uint32 {
	return w.worker.ActiveConnections()
}

func (w *BridgeWorker) handleInternalConn(link transport.Link) {
	go func() {
		defer func() {
			w.controlAccess.Lock()
			w.state = Control_DRAIN
			w.controlAccess.Unlock()
			common.Close(link.Writer)
		}()
		for {
			mb, err := link.Reader.ReadMultiBuffer()
			if err != nil {
				break
			}
			for _, b := range mb {
				var ctl Control
				if err := proto.Unmarshal(b.Bytes(), &ctl); err != nil {
					newError("failed to parse proto message").Base(err).WriteToLog()
					break
				}
				w.controlAccess.Lock()
				// A drained or expired worker must not become active again.
				if w.state == Control_ACTIVE {
					w.state = ctl.State
				}
				if ctl.HeartbeatId != 0 {
					w.lastControl = time.Now()
				}
				w.controlAccess.Unlock()
				// Legacy portals do not read replies or use the heartbeat deadline.
				if ctl.HeartbeatId != 0 {
					ack, err := proto.Marshal(&Control{HeartbeatId: ctl.HeartbeatId})
					if err != nil {
						newError("failed to encode control acknowledgement").Base(err).WriteToLog()
						buf.ReleaseMulti(mb)
						return
					}
					if err := link.Writer.WriteMultiBuffer(buf.MergeBytes(nil, ack)); err != nil {
						newError("failed to send control acknowledgement").Base(err).WriteToLog()
						buf.ReleaseMulti(mb)
						return
					}
				}
			}
			buf.ReleaseMulti(mb)
		}
	}()
}

func (w *BridgeWorker) Dispatch(ctx context.Context, dest net.Destination) (*transport.Link, error) {
	if !isInternalDomain(dest) {
		ctx = session.ContextWithInbound(ctx, &session.Inbound{
			Tag: w.tag,
		})
		return w.dispatcher.Dispatch(ctx, dest)
	}

	opt := []pipe.Option{pipe.WithSizeLimit(16 * 1024)}
	uplinkReader, uplinkWriter := pipe.New(opt...)
	downlinkReader, downlinkWriter := pipe.New(opt...)

	w.handleInternalConn(transport.Link{
		Reader: downlinkReader,
		Writer: uplinkWriter,
	})

	return &transport.Link{
		Reader: uplinkReader,
		Writer: downlinkWriter,
	}, nil
}
