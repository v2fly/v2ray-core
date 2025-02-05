//go:build !confonly
// +build !confonly

package observatory

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/v2fly/v2ray-core/v5/app/persistentstorage"
	"github.com/v2fly/v2ray-core/v5/app/persistentstorage/protostorage"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"

	"github.com/golang/protobuf/proto"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/common/signal/done"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/features/extension"
	"github.com/v2fly/v2ray-core/v5/features/outbound"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tagged"
)

type Observer struct {
	config *Config
	ctx    context.Context

	statusLock sync.Mutex
	status     []*OutboundStatus

	finished *done.Instance

	ohm            outbound.Manager
	persistStorage persistentstorage.ScopedPersistentStorage

	persistOutboundStatusProtoStorage protostorage.ProtoPersistentStorage
}

func (o *Observer) GetObservation(ctx context.Context) (proto.Message, error) {
	return &ObservationResult{Status: o.status}, nil
}

func (o *Observer) Type() interface{} {
	return extension.ObservatoryType()
}

func (o *Observer) Start() error {
	if o.config != nil && len(o.config.SubjectSelector) != 0 {
		if o.config.PersistentProbeResult {
			appEnvironment := envctx.EnvironmentFromContext(o.ctx).(environment.AppEnvironment)
			o.persistStorage = appEnvironment.PersistentStorage()

			outboundStatusStorage, err := o.persistStorage.NarrowScope(o.ctx, []byte("outbound_status"))
			if err != nil {
				return newError("failed to get persistent storage for outbound_status").Base(err)
			}
			o.persistOutboundStatusProtoStorage = outboundStatusStorage.(protostorage.ProtoPersistentStorage)
			list, err := outboundStatusStorage.List(o.ctx, []byte(""))
			if err != nil {
				newError("failed to list persisted outbound status").Base(err).WriteToLog()
			} else {
				for _, v := range list {
					o.loadOutboundStatus(string(v))
				}
			}
		}
		o.finished = done.New()
		go o.background()
	}
	return nil
}

func (o *Observer) Close() error {
	if o.finished != nil {
		return o.finished.Close()
	}
	return nil
}

func (o *Observer) background() {
	for !o.finished.Done() {
		hs, ok := o.ohm.(outbound.HandlerSelector)
		if !ok {
			newError("outbound.Manager is not a HandlerSelector").WriteToLog()
			return
		}

		outbounds := hs.Select(o.config.SubjectSelector)
		sort.Strings(outbounds)

		o.updateStatus(outbounds)

		slept := false
		for _, v := range outbounds {
			result := o.probe(v)
			o.updateStatusForResult(v, &result)
			if o.finished.Done() {
				return
			}
			sleepTime := time.Second * 10
			if o.config.ProbeInterval != 0 {
				sleepTime = time.Duration(o.config.ProbeInterval)
			}
			time.Sleep(sleepTime)
			slept = true
		}
		if !slept {
			sleepTime := time.Second * 10
			if o.config.ProbeInterval != 0 {
				sleepTime = time.Duration(o.config.ProbeInterval)
			}
			time.Sleep(sleepTime)
		}
	}
}

func (o *Observer) updateStatus(outbounds []string) {
	o.statusLock.Lock()
	defer o.statusLock.Unlock()
	// TODO should remove old inbound that is removed
	_ = outbounds
}

func (o *Observer) probe(outbound string) ProbeResult {
	errorCollectorForRequest := newErrorCollector()

	httpTransport := http.Transport{
		Proxy: func(*http.Request) (*url.URL, error) {
			return nil, nil
		},
		DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
			var connection net.Conn
			taskErr := task.Run(ctx, func() error {
				// MUST use V2Fly's built in context system
				dest, err := v2net.ParseDestination(network + ":" + addr)
				if err != nil {
					return newError("cannot understand address").Base(err)
				}
				trackedCtx := session.TrackedConnectionError(o.ctx, errorCollectorForRequest)
				conn, err := tagged.Dialer(trackedCtx, dest, outbound)
				if err != nil {
					return newError("cannot dial remote address ", dest).Base(err)
				}
				connection = conn
				return nil
			})
			if taskErr != nil {
				return nil, newError("cannot finish connection").Base(taskErr)
			}
			return connection, nil
		},
		TLSHandshakeTimeout: time.Second * 5,
	}
	httpClient := &http.Client{
		Transport: &httpTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar:     nil,
		Timeout: time.Second * 5,
	}
	var GETTime time.Duration
	err := task.Run(o.ctx, func() error {
		startTime := time.Now()
		probeURL := "https://api.v2fly.org/checkConnection.svgz"
		if o.config.ProbeUrl != "" {
			probeURL = o.config.ProbeUrl
		}
		response, err := httpClient.Get(probeURL)
		if err != nil {
			return newError("outbound failed to relay connection").Base(err)
		}
		if response.Body != nil {
			response.Body.Close()
		}
		endTime := time.Now()
		GETTime = endTime.Sub(startTime)
		return nil
	})
	if err != nil {
		fullerr := newError("underlying connection failed").Base(errorCollectorForRequest.UnderlyingError())
		fullerr = newError("with outbound handler report").Base(fullerr)
		fullerr = newError("GET request failed:", err).Base(fullerr)
		fullerr = newError("the outbound ", outbound, " is dead:").Base(fullerr)
		fullerr = fullerr.AtInfo()
		fullerr.WriteToLog()
		return ProbeResult{Alive: false, LastErrorReason: fullerr.Error()}
	}
	newError("the outbound ", outbound, " is alive:", GETTime.Seconds()).AtInfo().WriteToLog()
	return ProbeResult{Alive: true, Delay: GETTime.Milliseconds()}
}

func (o *Observer) updateStatusForResult(outbound string, result *ProbeResult) {
	o.statusLock.Lock()
	defer o.statusLock.Unlock()
	var status *OutboundStatus
	if location := o.findStatusLocationLockHolderOnly(outbound); location != -1 {
		status = o.status[location]
	} else {
		status = &OutboundStatus{}
		o.status = append(o.status, status)
	}

	status.LastTryTime = time.Now().Unix()
	status.OutboundTag = outbound
	status.Alive = result.Alive
	if result.Alive {
		status.Delay = result.Delay
		status.LastSeenTime = status.LastTryTime
		status.LastErrorReason = ""
	} else {
		status.LastErrorReason = result.LastErrorReason
		status.Delay = 99999999
	}
	if o.config.PersistentProbeResult {
		err := o.persistOutboundStatusProtoStorage.PutProto(o.ctx, outbound, status)
		if err != nil {
			newError("failed to persist outbound status").Base(err).WriteToLog()
		}
	}
}

func (o *Observer) findStatusLocationLockHolderOnly(outbound string) int {
	for i, v := range o.status {
		if v.OutboundTag == outbound {
			return i
		}
	}
	return -1
}

func (o *Observer) loadOutboundStatus(name string) {
	if o.persistOutboundStatusProtoStorage == nil {
		return
	}
	status := &OutboundStatus{}
	err := o.persistOutboundStatusProtoStorage.GetProto(o.ctx, name, status)
	if err != nil {
		newError("failed to load outbound status").Base(err).WriteToLog()
		return
	}
	o.status = append(o.status, status)
}

func New(ctx context.Context, config *Config) (*Observer, error) {
	obs := &Observer{
		config: config,
		ctx:    ctx,
	}

	err := core.RequireFeatures(ctx, func(om outbound.Manager) {
		obs.ohm = om
	})
	if err != nil {
		return nil, newError("Cannot get depended features").Base(err)
	}

	return obs, nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*Config))
	}))
}
