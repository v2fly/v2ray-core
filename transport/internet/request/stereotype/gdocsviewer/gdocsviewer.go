package gdocsviewer

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request/assembler/simple"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request/assembly"
	roundtripper "github.com/v2fly/v2ray-core/v5/transport/internet/request/roundtripper/gdocsviewer"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

const protocolName = "gdocsviewer"

const (
	defaultPollingResponseWaitMs = 10000
	defaultMaxPollingIntervalMs  = 10000
	defaultFailedRetryIntervalMs = 10000
)

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return nil, newError("gdocsviewer is a transport")
	}))

	common.Must(internet.RegisterTransportDialer(protocolName, dial))
	common.Must(internet.RegisterTransportListener(protocolName, listen))
}

func dial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	setting := streamSettings.ProtocolSettings.(*Config)
	requestConfig := buildClientRequestConfig(setting)
	constructedSetting := &internet.MemoryStreamConfig{
		ProtocolName:     "request",
		ProtocolSettings: requestConfig,
		SecurityType:     streamSettings.SecurityType,
		SecuritySettings: streamSettings.SecuritySettings,
		SocketSettings:   streamSettings.SocketSettings,
	}
	return internet.Dial(ctx, dest, constructedSetting)
}

func listen(ctx context.Context, address net.Address, port net.Port, streamSettings *internet.MemoryStreamConfig, callback internet.ConnHandler) (internet.Listener, error) {
	setting := streamSettings.ProtocolSettings.(*Config)
	requestConfig := buildServerRequestConfig(setting)
	constructedSetting := &internet.MemoryStreamConfig{
		ProtocolName:     "request",
		ProtocolSettings: requestConfig,
		SecurityType:     streamSettings.SecurityType,
		SecuritySettings: streamSettings.SecuritySettings,
		SocketSettings:   streamSettings.SocketSettings,
	}
	return internet.ListenTCP(ctx, address, port, constructedSetting, callback)
}

func buildClientRequestConfig(setting *Config) *assembly.Config {
	if setting == nil {
		setting = &Config{}
	}
	requestConfig := &assembly.Config{
		Assembler: serial.ToTypedMessage(&simple.ClientConfig{
			MaxWriteSize:             int32(maxRequestBytes(setting)),
			WaitSubsequentWriteMs:    10,
			InitialPollingIntervalMs: minRequestIntervalMs(setting),
			MaxPollingIntervalMs:     maxPollingIntervalMs(setting),
			MinPollingIntervalMs:     minRequestIntervalMs(setting),
			BackoffFactor:            1.5,
			FailedRetryIntervalMs:    defaultFailedRetryIntervalMs,
		}),
		Roundtripper: serial.ToTypedMessage(&roundtripper.ClientConfig{
			ViewerUrl:                 setting.ViewerUrl,
			TextUrl:                   setting.TextUrl,
			OriginUrl:                 setting.OriginUrl,
			ViewerHostHeader:          setting.ViewerHostHeader,
			UserAgent:                 setting.UserAgent,
			AllowHttp:                 setting.AllowHttp,
			H2PoolSize:                setting.H2PoolSize,
			MaxViewerBodyBytes:        setting.MaxViewerBodyBytes,
			MinRequestIntervalMs:      setting.MinRequestIntervalMs,
			SharedKey:                 setting.SharedKey,
			OriginUrlReplacementRules: convertOriginURLReplacementRules(setting.OriginUrlReplacementRules),
			RequestHeaders:            setting.RequestHeaders,
		}),
	}
	return requestConfig
}

func convertOriginURLReplacementRules(rules []*OriginUrlReplacementRule) []*roundtripper.OriginUrlReplacementRule {
	if len(rules) == 0 {
		return nil
	}
	converted := make([]*roundtripper.OriginUrlReplacementRule, 0, len(rules))
	for _, rule := range rules {
		if rule == nil {
			converted = append(converted, nil)
			continue
		}
		converted = append(converted, &roundtripper.OriginUrlReplacementRule{
			Name:    rule.Name,
			Pattern: rule.Pattern,
		})
	}
	return converted
}

func buildServerRequestConfig(setting *Config) *assembly.Config {
	if setting == nil {
		setting = &Config{}
	}
	requestConfig := &assembly.Config{
		Assembler: serial.ToTypedMessage(&simple.ServerConfig{
			MaxWriteSize:          int32(maxResponseBytes(setting)),
			PollingResponseWaitMs: defaultPollingResponseWaitMs,
		}),
		Roundtripper: serial.ToTypedMessage(&roundtripper.ServerConfig{
			PathPrefix:       setting.PathPrefix,
			MaxRequestBytes:  setting.MaxRequestBytes,
			MaxResponseBytes: setting.MaxResponseBytes,
			SharedKey:        setting.SharedKey,
		}),
	}
	return requestConfig
}

func maxPollingIntervalMs(config *Config) int32 {
	minInterval := minRequestIntervalMs(config)
	if minInterval > defaultMaxPollingIntervalMs {
		return minInterval
	}
	return defaultMaxPollingIntervalMs
}

func maxRequestBytes(config *Config) int {
	if config != nil && config.MaxRequestBytes > 0 {
		return int(config.MaxRequestBytes)
	}
	return 1100
}

func maxResponseBytes(config *Config) int {
	if config != nil && config.MaxResponseBytes > 0 {
		return int(config.MaxResponseBytes)
	}
	return 64 * 1024
}

func minRequestIntervalMs(config *Config) int32 {
	if config != nil && config.MinRequestIntervalMs > 0 {
		return config.MinRequestIntervalMs
	}
	return 100
}
