package v5cfg

import (
	"encoding/json"

	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon/muxcfg"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon/proxycfg"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon/sniffer"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon/socketcfg"
)

type RootConfig struct {
	LogConfig    json.RawMessage            `json:"log"`
	DNSConfig    json.RawMessage            `json:"dns"`
	RouterConfig json.RawMessage            `json:"router"`
	Inbounds     []InboundConfig            `json:"inbounds"`
	Outbounds    []OutboundConfig           `json:"outbounds"`
	Services     map[string]json.RawMessage `json:"services"`
	Extensions   []json.RawMessage          `json:"extension"`
}

type InboundConfig struct {
	Protocol       string                  `json:"protocol"`
	PortRange      *cfgcommon.PortRange    `json:"port"`
	ListenOn       *cfgcommon.Address      `json:"listen"`
	Settings       json.RawMessage         `json:"settings"`
	Tag            string                  `json:"tag"`
	SniffingConfig *sniffer.SniffingConfig `json:"sniffing"`
	StreamSetting  *StreamConfig           `json:"streamSettings"`
}

type OutboundConfig struct {
	Protocol      string                `json:"protocol"`
	SendThrough   *cfgcommon.Address    `json:"sendThrough"`
	Tag           string                `json:"tag"`
	Settings      json.RawMessage       `json:"settings"`
	StreamSetting *StreamConfig         `json:"streamSettings"`
	ProxySettings *proxycfg.ProxyConfig `json:"proxySettings"`
	MuxSettings   *muxcfg.MuxConfig     `json:"mux"`
}

type StreamConfig struct {
	Transport         string                 `json:"transport"`
	TransportSettings json.RawMessage        `json:"transportSettings"`
	Security          string                 `json:"security"`
	SecuritySettings  json.RawMessage        `json:"securitySettings"`
	SocketSettings    socketcfg.SocketConfig `json:"socketSettings"`
}
