package router

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

import (
	"context"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/platform"
	"github.com/v2fly/v2ray-core/v5/features/dns"
	"github.com/v2fly/v2ray-core/v5/features/outbound"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	routing_dns "github.com/v2fly/v2ray-core/v5/features/routing/dns"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v5/infra/conf/geodata"
)

// Router is an implementation of routing.Router.
type Router struct {
	domainStrategy DomainStrategy
	rules          []*Rule
	balancers      map[string]*Balancer
	dns            dns.Client
}

// Route is an implementation of routing.Route.
type Route struct {
	routing.Context
	outboundGroupTags []string
	outboundTag       string
}

// Init initializes the Router.
func (r *Router) Init(ctx context.Context, config *Config, d dns.Client, ohm outbound.Manager, dispatcher routing.Dispatcher) error {
	r.domainStrategy = config.DomainStrategy
	r.dns = d

	r.balancers = make(map[string]*Balancer, len(config.BalancingRule))
	for _, rule := range config.BalancingRule {
		balancer, err := rule.Build(ohm, dispatcher)
		if err != nil {
			return err
		}
		balancer.InjectContext(ctx)
		r.balancers[rule.Tag] = balancer
	}

	r.rules = make([]*Rule, 0, len(config.Rule))
	for _, rule := range config.Rule {
		cond, err := rule.BuildCondition()
		if err != nil {
			return err
		}
		rr := &Rule{
			Condition: cond,
			Tag:       rule.GetTag(),
		}
		btag := rule.GetBalancingTag()
		if len(btag) > 0 {
			brule, found := r.balancers[btag]
			if !found {
				return newError("balancer ", btag, " not found")
			}
			rr.Balancer = brule
		}
		r.rules = append(r.rules, rr)
	}

	return nil
}

// PickRoute implements routing.Router.
func (r *Router) PickRoute(ctx routing.Context) (routing.Route, error) {
	rule, ctx, err := r.pickRouteInternal(ctx)
	if err != nil {
		return nil, err
	}
	tag, err := rule.GetTag()
	if err != nil {
		return nil, err
	}
	return &Route{Context: ctx, outboundTag: tag}, nil
}

func (r *Router) pickRouteInternal(ctx routing.Context) (*Rule, routing.Context, error) {
	// SkipDNSResolve is set from DNS module.
	// the DOH remote server maybe a domain name,
	// this prevents cycle resolving dead loop
	skipDNSResolve := ctx.GetSkipDNSResolve()

	if r.domainStrategy == DomainStrategy_IpOnDemand && !skipDNSResolve {
		ctx = routing_dns.ContextWithDNSClient(ctx, r.dns)
	}

	for _, rule := range r.rules {
		if rule.Apply(ctx) {
			return rule, ctx, nil
		}
	}

	if r.domainStrategy != DomainStrategy_IpIfNonMatch || len(ctx.GetTargetDomain()) == 0 || skipDNSResolve {
		return nil, ctx, common.ErrNoClue
	}

	ctx = routing_dns.ContextWithDNSClient(ctx, r.dns)

	// Try applying rules again if we have IPs.
	for _, rule := range r.rules {
		if rule.Apply(ctx) {
			return rule, ctx, nil
		}
	}

	return nil, ctx, common.ErrNoClue
}

// Start implements common.Runnable.
func (r *Router) Start() error {
	return nil
}

// Close implements common.Closable.
func (r *Router) Close() error {
	return nil
}

// Type implements common.HasType.
func (*Router) Type() interface{} {
	return routing.RouterType()
}

// GetOutboundGroupTags implements routing.Route.
func (r *Route) GetOutboundGroupTags() []string {
	return r.outboundGroupTags
}

// GetOutboundTag implements routing.Route.
func (r *Route) GetOutboundTag() string {
	return r.outboundTag
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		r := new(Router)
		if err := core.RequireFeatures(ctx, func(d dns.Client, ohm outbound.Manager, dispatcher routing.Dispatcher) error {
			return r.Init(ctx, config.(*Config), d, ohm, dispatcher)
		}); err != nil {
			return nil, err
		}
		return r, nil
	}))

	common.Must(common.RegisterConfig((*SimplifiedConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		ctx = cfgcommon.NewConfigureLoadingContext(ctx)

		geoloadername := platform.NewEnvFlag("v2ray.conf.geoloader").GetValue(func() string {
			return "standard"
		})

		if loader, err := geodata.GetGeoDataLoader(geoloadername); err == nil {
			cfgcommon.SetGeoDataLoader(ctx, loader)
		} else {
			return nil, newError("unable to create geo data loader ").Base(err)
		}

		cfgEnv := cfgcommon.GetConfigureLoadingEnvironment(ctx)
		geoLoader := cfgEnv.GetGeoLoader()

		simplifiedConfig := config.(*SimplifiedConfig)

		var routingRules []*RoutingRule

		for _, v := range simplifiedConfig.Rule {
			rule := new(RoutingRule)

			for _, geo := range v.Geoip {
				if geo.Code != "" {
					filepath := "geoip.dat"
					if geo.FilePath != "" {
						filepath = geo.FilePath
					} else {
						geo.CountryCode = geo.Code
					}
					var err error
					geo.Cidr, err = geoLoader.LoadIP(filepath, geo.Code)
					if err != nil {
						return nil, newError("unable to load geoip").Base(err)
					}
				}
			}
			rule.Geoip = v.Geoip

			for _, geo := range v.SourceGeoip {
				if geo.Code != "" {
					filepath := "geoip.dat"
					if geo.FilePath != "" {
						filepath = geo.FilePath
					} else {
						geo.CountryCode = geo.Code
					}
					var err error
					geo.Cidr, err = geoLoader.LoadIP(filepath, geo.Code)
					if err != nil {
						return nil, newError("unable to load geoip").Base(err)
					}
				}
			}
			rule.SourceGeoip = v.SourceGeoip

			for _, geo := range v.GeoDomain {
				if geo.Code != "" {
					filepath := "geosite.dat"
					if geo.FilePath != "" {
						filepath = geo.FilePath
					}
					var err error
					geo.Domain, err = geoLoader.LoadGeoSiteWithAttr(filepath, geo.Code)
					if err != nil {
						return nil, newError("unable to load geodomain").Base(err)
					}
				}
			}
			if v.PortList != "" {
				portList := &cfgcommon.PortList{}
				err := portList.UnmarshalText(v.PortList)
				if err != nil {
					return nil, err
				}
				rule.PortList = portList.Build()
			}
			if v.SourcePortList != "" {
				portList := &cfgcommon.PortList{}
				err := portList.UnmarshalText(v.SourcePortList)
				if err != nil {
					return nil, err
				}
				rule.SourcePortList = portList.Build()
			}
			rule.Domain = v.Domain
			rule.GeoDomain = v.GeoDomain
			rule.Networks = v.Networks.GetNetwork()
			rule.Protocol = v.Protocol
			rule.Attributes = v.Attributes
			rule.UserEmail = v.UserEmail
			rule.InboundTag = v.InboundTag
			rule.DomainMatcher = v.DomainMatcher
			switch s := v.TargetTag.(type) {
			case *SimplifiedRoutingRule_Tag:
				rule.TargetTag = &RoutingRule_Tag{s.Tag}
			case *SimplifiedRoutingRule_BalancingTag:
				rule.TargetTag = &RoutingRule_BalancingTag{s.BalancingTag}
			}
			routingRules = append(routingRules, rule)
		}

		fullConfig := &Config{
			DomainStrategy: simplifiedConfig.DomainStrategy,
			Rule:           routingRules,
			BalancingRule:  simplifiedConfig.BalancingRule,
		}
		return common.CreateObject(ctx, fullConfig)
	}))
}
