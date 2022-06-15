//go:build !confonly
// +build !confonly

package router

import (
	"context"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/observatory"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/features"
	"github.com/v2fly/v2ray-core/v5/features/extension"
)

type FallbackStrategy struct {
	ctx         context.Context
	observatory extension.Observatory

	config *StrategyFallbackConfig
}

func (l *FallbackStrategy) GetPrincipleTarget(strings []string) []string {
	return []string{l.PickOutbound(strings)}
}

func (l *FallbackStrategy) InjectContext(ctx context.Context) {
	l.ctx = ctx
}

func (l *FallbackStrategy) PickOutbound(strings []string) string {
	if l.observatory == nil {
		common.Must(core.RequireFeatures(l.ctx, func(observatory extension.Observatory) error {
			if l.config.ObserverTag != "" {
				l.observatory = common.Must2(observatory.(features.TaggedFeatures).GetFeaturesByTag(l.config.ObserverTag)).(extension.Observatory)
			} else {
				l.observatory = observatory
			}
			return nil
		}))
	}

	observeReport, err := l.observatory.GetObservation(l.ctx)
	if err != nil {
		newError("cannot get observe report").Base(err).WriteToLog()
		return ""
	}
	outboundsList := outboundList(strings)
	result, ok := observeReport.(*observatory.ObservationResult)
	if !ok {
		return ""
	}
	status := result.Status
	for _, outbound := range outboundsList {
		for _, v := range status {
			if outbound == v.OutboundTag && v.Alive {
				println(outbound)
				return outbound
			}
		}
	}

	// No way to understand observeReport
	return ""
}

func init() {
	common.Must(common.RegisterConfig((*StrategyFallbackConfig)(nil), nil))
}
