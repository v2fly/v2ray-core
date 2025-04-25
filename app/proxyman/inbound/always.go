package inbound

import (
	"context"

	core "github.com/ghxhy/v2ray-core/v5"
	"github.com/ghxhy/v2ray-core/v5/app/proxyman"
	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/dice"
	"github.com/ghxhy/v2ray-core/v5/common/errors"
	"github.com/ghxhy/v2ray-core/v5/common/mux"
	"github.com/ghxhy/v2ray-core/v5/common/net"
	"github.com/ghxhy/v2ray-core/v5/features/policy"
	"github.com/ghxhy/v2ray-core/v5/features/stats"
	"github.com/ghxhy/v2ray-core/v5/proxy"
	"github.com/ghxhy/v2ray-core/v5/transport/internet"
)

func getStatCounter(v *core.Instance, tag string) (stats.Counter, stats.Counter) {
	var uplinkCounter stats.Counter
	var downlinkCounter stats.Counter

	policy := v.GetFeature(policy.ManagerType()).(policy.Manager)
	if len(tag) > 0 && policy.ForSystem().Stats.InboundUplink {
		statsManager := v.GetFeature(stats.ManagerType()).(stats.Manager)
		name := "inbound>>>" + tag + ">>>traffic>>>uplink"
		c, _ := stats.GetOrRegisterCounter(statsManager, name)
		if c != nil {
			uplinkCounter = c
		}
	}
	if len(tag) > 0 && policy.ForSystem().Stats.InboundDownlink {
		statsManager := v.GetFeature(stats.ManagerType()).(stats.Manager)
		name := "inbound>>>" + tag + ">>>traffic>>>downlink"
		c, _ := stats.GetOrRegisterCounter(statsManager, name)
		if c != nil {
			downlinkCounter = c
		}
	}

	return uplinkCounter, downlinkCounter
}

type AlwaysOnInboundHandler struct {
	proxy   proxy.Inbound
	workers []worker
	mux     *mux.Server
	tag     string
}

func NewAlwaysOnInboundHandler(ctx context.Context, tag string, receiverConfig *proxyman.ReceiverConfig, proxyConfig interface{}) (*AlwaysOnInboundHandler, error) {
	rawProxy, err := common.CreateObject(ctx, proxyConfig)
	if err != nil {
		return nil, err
	}
	p, ok := rawProxy.(proxy.Inbound)
	if !ok {
		return nil, newError("not an inbound proxy.")
	}

	h := &AlwaysOnInboundHandler{
		proxy: p,
		mux:   mux.NewServer(ctx),
		tag:   tag,
	}

	uplinkCounter, downlinkCounter := getStatCounter(core.MustFromContext(ctx), tag)

	nl := p.Network()
	pr := receiverConfig.PortRange
	address := receiverConfig.Listen.AsAddress()
	if address == nil {
		address = net.AnyIP
	}

	mss, err := internet.ToMemoryStreamConfig(receiverConfig.StreamSettings)
	if err != nil {
		return nil, newError("failed to parse stream config").Base(err).AtWarning()
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
	if pr == nil {
		if net.HasNetwork(nl, net.Network_UNIX) {
			newError("creating unix domain socket worker on ", address).AtDebug().WriteToLog()

			worker := &dsWorker{
				address:         address,
				proxy:           p,
				stream:          mss,
				tag:             tag,
				dispatcher:      h.mux,
				sniffingConfig:  receiverConfig.GetEffectiveSniffingSettings(),
				uplinkCounter:   uplinkCounter,
				downlinkCounter: downlinkCounter,
				ctx:             ctx,
			}
			h.workers = append(h.workers, worker)
		}
	}
	if pr != nil {
		for port := pr.From; port <= pr.To; port++ {
			if net.HasNetwork(nl, net.Network_TCP) {
				newError("creating stream worker on ", address, ":", port).AtDebug().WriteToLog()

				worker := &tcpWorker{
					address:         address,
					port:            net.Port(port),
					proxy:           p,
					stream:          mss,
					recvOrigDest:    receiverConfig.ReceiveOriginalDestination,
					tag:             tag,
					dispatcher:      h.mux,
					sniffingConfig:  receiverConfig.GetEffectiveSniffingSettings(),
					uplinkCounter:   uplinkCounter,
					downlinkCounter: downlinkCounter,
					ctx:             ctx,
				}
				h.workers = append(h.workers, worker)
			}

			if net.HasNetwork(nl, net.Network_UDP) {
				worker := &udpWorker{
					ctx:             ctx,
					tag:             tag,
					proxy:           p,
					address:         address,
					port:            net.Port(port),
					dispatcher:      h.mux,
					sniffingConfig:  receiverConfig.GetEffectiveSniffingSettings(),
					uplinkCounter:   uplinkCounter,
					downlinkCounter: downlinkCounter,
					stream:          mss,
				}
				h.workers = append(h.workers, worker)
			}
		}
	}

	return h, nil
}

// Start implements common.Runnable.
func (h *AlwaysOnInboundHandler) Start() error {
	for _, worker := range h.workers {
		if err := worker.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Close implements common.Closable.
func (h *AlwaysOnInboundHandler) Close() error {
	var errs []error
	for _, worker := range h.workers {
		errs = append(errs, worker.Close())
	}
	errs = append(errs, h.mux.Close())
	if err := errors.Combine(errs...); err != nil {
		return newError("failed to close all resources").Base(err)
	}
	return nil
}

func (h *AlwaysOnInboundHandler) GetRandomInboundProxy() (interface{}, net.Port, int) {
	if len(h.workers) == 0 {
		return nil, 0, 0
	}
	w := h.workers[dice.Roll(len(h.workers))]
	return w.Proxy(), w.Port(), 9999
}

func (h *AlwaysOnInboundHandler) Tag() string {
	return h.tag
}

func (h *AlwaysOnInboundHandler) GetInbound() proxy.Inbound {
	return h.proxy
}
