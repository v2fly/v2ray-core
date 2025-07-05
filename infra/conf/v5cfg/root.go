package v5cfg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/dispatcher"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	"github.com/v2fly/v2ray-core/v5/common/platform"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v5/infra/conf/geodata"
	"github.com/v2fly/v2ray-core/v5/infra/conf/synthetic/log"
)

func (c RootConfig) BuildV5(ctx context.Context) (proto.Message, error) {
	config := &core.Config{
		App: []*anypb.Any{
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
		},
	}

	var logConfMsg *anypb.Any
	if c.LogConfig != nil {
		logConfMsgUnpacked, err := loadHeterogeneousConfigFromRawJSON("service", "log", c.LogConfig)
		if err != nil {
			return nil, newError("failed to parse Log config").Base(err)
		}
		logConfMsg = serial.ToTypedMessage(logConfMsgUnpacked)
	} else {
		logConfMsg = serial.ToTypedMessage(log.DefaultLogConfig())
	}
	// let logger module be the first App to start,
	// so that other modules could print log during initiating
	config.App = append([]*anypb.Any{logConfMsg}, config.App...)

	if c.RouterConfig != nil {
		routerConfig, err := loadHeterogeneousConfigFromRawJSON("service", "router", c.RouterConfig)
		if err != nil {
			return nil, newError("failed to parse Router config").Base(err)
		}
		config.App = append(config.App, serial.ToTypedMessage(routerConfig))
	}

	if c.DNSConfig != nil {
		dnsApp, err := loadHeterogeneousConfigFromRawJSON("service", "dns", c.DNSConfig)
		if err != nil {
			return nil, newError("failed to parse DNS config").Base(err)
		}
		config.App = append(config.App, serial.ToTypedMessage(dnsApp))
	}

	for _, rawInboundConfig := range c.Inbounds {
		ic, err := rawInboundConfig.BuildV5(ctx)
		if err != nil {
			return nil, err
		}
		config.Inbound = append(config.Inbound, ic.(*core.InboundHandlerConfig))
	}

	for _, rawOutboundConfig := range c.Outbounds {
		ic, err := rawOutboundConfig.BuildV5(ctx)
		if err != nil {
			return nil, err
		}
		config.Outbound = append(config.Outbound, ic.(*core.OutboundHandlerConfig))
	}

	for serviceName, service := range c.Services {
		servicePackedConfig, err := loadHeterogeneousConfigFromRawJSON("service", serviceName, service)
		if err != nil {
			return nil, newError(fmt.Sprintf("failed to parse %v config in Services", serviceName)).Base(err)
		}
		config.App = append(config.App, serial.ToTypedMessage(servicePackedConfig))
	}
	return config, nil
}

func loadJSONConfig(data []byte) (*core.Config, error) {
	rootConfig := &RootConfig{}

	rootConfDecoder := json.NewDecoder(bytes.NewReader(data))
	rootConfDecoder.DisallowUnknownFields()
	err := rootConfDecoder.Decode(rootConfig)
	if err != nil {
		return nil, newError("unable to load json").Base(err)
	}

	buildctx := cfgcommon.NewConfigureLoadingContext(context.Background())

	geoloadername := platform.NewEnvFlag("v2ray.conf.geoloader").GetValue(func() string {
		return "standard"
	})

	if loader, err := geodata.GetGeoDataLoader(geoloadername); err == nil {
		cfgcommon.SetGeoDataLoader(buildctx, loader)
	} else {
		return nil, newError("unable to create geo data loader ").Base(err)
	}

	message, err := rootConfig.BuildV5(buildctx)
	if err != nil {
		return nil, newError("unable to build config").Base(err)
	}
	return message.(*core.Config), nil
}
