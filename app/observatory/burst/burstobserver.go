package burst

import (
	"context"
	"sync"

	"github.com/golang/protobuf/proto"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/observatory"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/signal/done"
	"github.com/v2fly/v2ray-core/v5/features/extension"
	"github.com/v2fly/v2ray-core/v5/features/outbound"
)

type Observer struct {
	config *Config
	ctx    context.Context

	statusLock sync.Mutex // nolint: structcheck
	hp         *HealthPing

	finished *done.Instance

	ohm outbound.Manager
}

func (o *Observer) GetObservation(ctx context.Context) (proto.Message, error) {
	return &observatory.ObservationResult{Status: o.createResult()}, nil
}

func (o *Observer) createResult() []*observatory.OutboundStatus {
	var result []*observatory.OutboundStatus
	o.hp.access.Lock()
	defer o.hp.access.Unlock()
	for name, value := range o.hp.Results {
		status := observatory.OutboundStatus{
			Alive:           value.getStatistics().All != value.getStatistics().Fail,
			Delay:           value.getStatistics().Average.Milliseconds(),
			LastErrorReason: "",
			OutboundTag:     name,
			LastSeenTime:    0,
			LastTryTime:     0,
			HealthPing: &observatory.HealthPingMeasurementResult{
				All:       int64(value.getStatistics().All),
				Fail:      int64(value.getStatistics().Fail),
				Deviation: int64(value.getStatistics().Deviation),
				Average:   int64(value.getStatistics().Average),
				Max:       int64(value.getStatistics().Max),
				Min:       int64(value.getStatistics().Min),
			},
		}
		result = append(result, &status)
	}
	return result
}

func (o *Observer) Type() interface{} {
	return extension.ObservatoryType()
}

func (o *Observer) Start() error {
	if o.config != nil && len(o.config.SubjectSelector) != 0 {
		o.finished = done.New()
		o.hp.StartScheduler(func() ([]string, error) {
			hs, ok := o.ohm.(outbound.HandlerSelector)
			if !ok {
				return nil, newError("outbound.Manager is not a HandlerSelector")
			}

			outbounds := hs.Select(o.config.SubjectSelector)
			return outbounds, nil
		})
	}
	return nil
}

func (o *Observer) Close() error {
	if o.finished != nil {
		o.hp.StopScheduler()
		return o.finished.Close()
	}
	return nil
}

func New(ctx context.Context, config *Config) (*Observer, error) {
	var outboundManager outbound.Manager
	err := core.RequireFeatures(ctx, func(om outbound.Manager) {
		outboundManager = om
	})
	if err != nil {
		return nil, newError("Cannot get depended features").Base(err)
	}
	hp := NewHealthPing(ctx, config.PingConfig)
	return &Observer{
		config: config,
		ctx:    ctx,
		ohm:    outboundManager,
		hp:     hp,
	}, nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*Config))
	}))
}
