//go:build !confonly
// +build !confonly

package router

import (
	"context"
	"encoding/json"

	"github.com/golang/protobuf/jsonpb"

	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/features/outbound"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/infra/conf/v5cfg"
)

type Rule struct {
	Tag       string
	Balancer  *Balancer
	Condition Condition
}

func (r *Rule) GetTag() (string, error) {
	if r.Balancer != nil {
		return r.Balancer.PickOutbound()
	}
	return r.Tag, nil
}

// Apply checks rule matching of current routing context.
func (r *Rule) Apply(ctx routing.Context) bool {
	return r.Condition.Apply(ctx)
}

func (rr *RoutingRule) BuildCondition() (Condition, error) {
	conds := NewConditionChan()

	if len(rr.Domain) > 0 {
		cond, err := NewDomainMatcher(rr.DomainMatcher, rr.Domain)
		if err != nil {
			return nil, newError("failed to build domain condition").Base(err)
		}
		conds.Add(cond)
	}

	var geoDomains []*routercommon.Domain
	for _, geo := range rr.GeoDomain {
		geoDomains = append(geoDomains, geo.Domain...)
	}
	if len(geoDomains) > 0 {
		cond, err := NewDomainMatcher(rr.DomainMatcher, geoDomains)
		if err != nil {
			return nil, newError("failed to build geo domain condition").Base(err)
		}
		conds.Add(cond)
	}

	if len(rr.UserEmail) > 0 {
		conds.Add(NewUserMatcher(rr.UserEmail))
	}

	if len(rr.InboundTag) > 0 {
		conds.Add(NewInboundTagMatcher(rr.InboundTag))
	}

	if rr.PortList != nil {
		conds.Add(NewPortMatcher(rr.PortList, false))
	} else if rr.PortRange != nil {
		conds.Add(NewPortMatcher(&net.PortList{Range: []*net.PortRange{rr.PortRange}}, false))
	}

	if rr.SourcePortList != nil {
		conds.Add(NewPortMatcher(rr.SourcePortList, true))
	}

	if len(rr.Networks) > 0 {
		conds.Add(NewNetworkMatcher(rr.Networks))
	} else if rr.NetworkList != nil {
		conds.Add(NewNetworkMatcher(rr.NetworkList.Network))
	}

	if len(rr.Geoip) > 0 {
		cond, err := NewMultiGeoIPMatcher(rr.Geoip, false)
		if err != nil {
			return nil, err
		}
		conds.Add(cond)
	} else if len(rr.Cidr) > 0 {
		cond, err := NewMultiGeoIPMatcher([]*routercommon.GeoIP{{Cidr: rr.Cidr}}, false)
		if err != nil {
			return nil, err
		}
		conds.Add(cond)
	}

	if len(rr.SourceGeoip) > 0 {
		cond, err := NewMultiGeoIPMatcher(rr.SourceGeoip, true)
		if err != nil {
			return nil, err
		}
		conds.Add(cond)
	} else if len(rr.SourceCidr) > 0 {
		cond, err := NewMultiGeoIPMatcher([]*routercommon.GeoIP{{Cidr: rr.SourceCidr}}, true)
		if err != nil {
			return nil, err
		}
		conds.Add(cond)
	}

	if len(rr.Protocol) > 0 {
		conds.Add(NewProtocolMatcher(rr.Protocol))
	}

	if len(rr.Attributes) > 0 {
		cond, err := NewAttributeMatcher(rr.Attributes)
		if err != nil {
			return nil, err
		}
		conds.Add(cond)
	}

	if conds.Len() == 0 {
		return nil, newError("this rule has no effective fields").AtWarning()
	}

	return conds, nil
}

// Build builds the balancing rule
func (br *BalancingRule) Build(ohm outbound.Manager, dispatcher routing.Dispatcher) (*Balancer, error) {
	switch br.Strategy {
	case "leastping":
		i, err := serial.GetInstanceOf(br.StrategySettings)
		if err != nil {
			return nil, err
		}
		s, ok := i.(*StrategyLeastPingConfig)
		if !ok {
			return nil, newError("not a StrategyLeastPingConfig").AtError()
		}
		return &Balancer{
			selectors: br.OutboundSelector,
			strategy:  &LeastPingStrategy{config: s},
			ohm:       ohm,
		}, nil
	case "leastload":
		i, err := serial.GetInstanceOf(br.StrategySettings)
		if err != nil {
			return nil, err
		}
		s, ok := i.(*StrategyLeastLoadConfig)
		if !ok {
			return nil, newError("not a StrategyLeastLoadConfig").AtError()
		}
		leastLoadStrategy := NewLeastLoadStrategy(s)
		return &Balancer{
			selectors: br.OutboundSelector,
			ohm:       ohm, fallbackTag: br.FallbackTag,
			strategy: leastLoadStrategy,
		}, nil
	case "random":
		fallthrough
	case "":
		return &Balancer{
			selectors: br.OutboundSelector,
			ohm:       ohm, fallbackTag: br.FallbackTag,
			strategy: &RandomStrategy{},
		}, nil
	default:
		return nil, newError("unrecognized balancer type")
	}
}

func (br *BalancingRule) UnmarshalJSONPB(unmarshaler *jsonpb.Unmarshaler, bytes []byte) error {
	type BalancingRuleStub struct {
		Tag              string          `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
		OutboundSelector []string        `protobuf:"bytes,2,rep,name=outbound_selector,json=outboundSelector,proto3" json:"outbound_selector,omitempty"`
		Strategy         string          `protobuf:"bytes,3,opt,name=strategy,proto3" json:"strategy,omitempty"`
		StrategySettings json.RawMessage `protobuf:"bytes,4,opt,name=strategy_settings,json=strategySettings,proto3" json:"strategy_settings,omitempty"`
		FallbackTag      string          `protobuf:"bytes,5,opt,name=fallback_tag,json=fallbackTag,proto3" json:"fallback_tag,omitempty"`
	}

	var stub BalancingRuleStub
	if err := json.Unmarshal(bytes, &stub); err != nil {
		return err
	}
	if stub.Strategy == "" {
		stub.Strategy = "random"
	}
	settingsPack, err := v5cfg.LoadHeterogeneousConfigFromRawJSON(context.TODO(), "balancer", stub.Strategy, stub.StrategySettings)
	if err != nil {
		return err
	}
	br.StrategySettings = serial.ToTypedMessage(settingsPack)

	br.Tag = stub.Tag
	br.Strategy = stub.Strategy
	br.OutboundSelector = stub.OutboundSelector
	br.FallbackTag = stub.FallbackTag

	return nil
}
