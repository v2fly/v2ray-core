package inbound

import (
	"context"
	"sync"
	"time"

	core "github.com/ghxhy/v2ray-core/v5"
	"github.com/ghxhy/v2ray-core/v5/app/proxyman"
	"github.com/ghxhy/v2ray-core/v5/common/dice"
	"github.com/ghxhy/v2ray-core/v5/common/mux"
	"github.com/ghxhy/v2ray-core/v5/common/net"
	"github.com/ghxhy/v2ray-core/v5/common/task"
	"github.com/ghxhy/v2ray-core/v5/proxy"
	"github.com/ghxhy/v2ray-core/v5/transport/internet"
)

type DynamicInboundHandler struct {
	tag            string
	v              *core.Instance
	proxyConfig    interface{}
	receiverConfig *proxyman.ReceiverConfig
	streamSettings *internet.MemoryStreamConfig
	portMutex      sync.Mutex
	portsInUse     map[net.Port]bool
	workerMutex    sync.RWMutex
	worker         []worker
	lastRefresh    time.Time
	mux            *mux.Server
	task           *task.Periodic

	ctx context.Context
}

func NewDynamicInboundHandler(ctx context.Context, tag string, receiverConfig *proxyman.ReceiverConfig, proxyConfig interface{}) (*DynamicInboundHandler, error) {
	v := core.MustFromContext(ctx)
	h := &DynamicInboundHandler{
		tag:            tag,
		proxyConfig:    proxyConfig,
		receiverConfig: receiverConfig,
		portsInUse:     make(map[net.Port]bool),
		mux:            mux.NewServer(ctx),
		v:              v,
		ctx:            ctx,
	}

	mss, err := internet.ToMemoryStreamConfig(receiverConfig.StreamSettings)
	if err != nil {
		return nil, newError("failed to parse stream settings").Base(err).AtWarning()
	}
	if receiverConfig.ReceiveOriginalDestination {
		if mss.SocketSettings == nil {
			mss.SocketSettings = &internet.SocketConfig{}
		}
		if mss.SocketSettings.Tproxy == internet.SocketConfig_Off {
			mss.SocketSettings.Tproxy = internet.SocketConfig_Redirect
		}
		mss.SocketSettings.ReceiveOriginalDestAddress = true
	}

	h.streamSettings = mss

	h.task = &task.Periodic{
		Interval: time.Minute * time.Duration(h.receiverConfig.AllocationStrategy.GetRefreshValue()),
		Execute:  h.refresh,
	}

	return h, nil
}

func (h *DynamicInboundHandler) allocatePort() net.Port {
	from := int(h.receiverConfig.PortRange.From)
	delta := int(h.receiverConfig.PortRange.To) - from + 1

	h.portMutex.Lock()
	defer h.portMutex.Unlock()

	for {
		r := dice.Roll(delta)
		port := net.Port(from + r)
		_, used := h.portsInUse[port]
		if !used {
			h.portsInUse[port] = true
			return port
		}
	}
}

func (h *DynamicInboundHandler) closeWorkers(workers []worker) {
	ports2Del := make([]net.Port, len(workers))
	for idx, worker := range workers {
		ports2Del[idx] = worker.Port()
		if err := worker.Close(); err != nil {
			newError("failed to close worker").Base(err).WriteToLog()
		}
	}

	h.portMutex.Lock()
	for _, port := range ports2Del {
		delete(h.portsInUse, port)
	}
	h.portMutex.Unlock()
}

func (h *DynamicInboundHandler) refresh() error {
	h.lastRefresh = time.Now()

	timeout := time.Minute * time.Duration(h.receiverConfig.AllocationStrategy.GetRefreshValue()) * 2
	concurrency := h.receiverConfig.AllocationStrategy.GetConcurrencyValue()
	workers := make([]worker, 0, concurrency)

	address := h.receiverConfig.Listen.AsAddress()
	if address == nil {
		address = net.AnyIP
	}

	uplinkCounter, downlinkCounter := getStatCounter(h.v, h.tag)

	for i := uint32(0); i < concurrency; i++ {
		port := h.allocatePort()
		rawProxy, err := core.CreateObject(h.v, h.proxyConfig)
		if err != nil {
			newError("failed to create proxy instance").Base(err).AtWarning().WriteToLog()
			continue
		}
		p := rawProxy.(proxy.Inbound)
		nl := p.Network()
		if net.HasNetwork(nl, net.Network_TCP) {
			worker := &tcpWorker{
				tag:             h.tag,
				address:         address,
				port:            port,
				proxy:           p,
				stream:          h.streamSettings,
				recvOrigDest:    h.receiverConfig.ReceiveOriginalDestination,
				dispatcher:      h.mux,
				sniffingConfig:  h.receiverConfig.GetEffectiveSniffingSettings(),
				uplinkCounter:   uplinkCounter,
				downlinkCounter: downlinkCounter,
				ctx:             h.ctx,
			}
			if err := worker.Start(); err != nil {
				newError("failed to create TCP worker").Base(err).AtWarning().WriteToLog()
				continue
			}
			workers = append(workers, worker)
		}

		if net.HasNetwork(nl, net.Network_UDP) {
			worker := &udpWorker{
				ctx:             h.ctx,
				tag:             h.tag,
				proxy:           p,
				address:         address,
				port:            port,
				dispatcher:      h.mux,
				sniffingConfig:  h.receiverConfig.GetEffectiveSniffingSettings(),
				uplinkCounter:   uplinkCounter,
				downlinkCounter: downlinkCounter,
				stream:          h.streamSettings,
			}
			if err := worker.Start(); err != nil {
				newError("failed to create UDP worker").Base(err).AtWarning().WriteToLog()
				continue
			}
			workers = append(workers, worker)
		}
	}

	h.workerMutex.Lock()
	h.worker = workers
	h.workerMutex.Unlock()

	time.AfterFunc(timeout, func() {
		h.closeWorkers(workers)
	})

	return nil
}

func (h *DynamicInboundHandler) Start() error {
	return h.task.Start()
}

func (h *DynamicInboundHandler) Close() error {
	return h.task.Close()
}

func (h *DynamicInboundHandler) GetRandomInboundProxy() (interface{}, net.Port, int) {
	h.workerMutex.RLock()
	defer h.workerMutex.RUnlock()

	if len(h.worker) == 0 {
		return nil, 0, 0
	}
	w := h.worker[dice.Roll(len(h.worker))]
	expire := h.receiverConfig.AllocationStrategy.GetRefreshValue() - uint32(time.Since(h.lastRefresh)/time.Minute)
	return w.Proxy(), w.Port(), int(expire)
}

func (h *DynamicInboundHandler) Tag() string {
	return h.tag
}
