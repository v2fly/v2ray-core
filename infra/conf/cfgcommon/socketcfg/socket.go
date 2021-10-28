package socketcfg

import (
	"strings"

	"github.com/v2fly/v2ray-core/v4/transport/internet"
)

type SocketConfig struct {
	Mark                 uint32 `json:"mark"`
	TFO                  *bool  `json:"tcpFastOpen"`
	TProxy               string `json:"tproxy"`
	AcceptProxyProtocol  bool   `json:"acceptProxyProtocol"`
	TCPKeepAliveInterval int32  `json:"tcpKeepAliveInterval"`
	TFOQueueLength       uint32 `json:"tcpFastOpenQueueLength"`
}

// Build implements Buildable.
func (c *SocketConfig) Build() (*internet.SocketConfig, error) {
	var tfoSettings internet.SocketConfig_TCPFastOpenState
	if c.TFO != nil {
		if *c.TFO {
			tfoSettings = internet.SocketConfig_Enable
		} else {
			tfoSettings = internet.SocketConfig_Disable
		}
	}

	tfoQueueLength := c.TFOQueueLength
	if tfoQueueLength == 0 {
		tfoQueueLength = 4096
	}

	var tproxy internet.SocketConfig_TProxyMode
	switch strings.ToLower(c.TProxy) {
	case "tproxy":
		tproxy = internet.SocketConfig_TProxy
	case "redirect":
		tproxy = internet.SocketConfig_Redirect
	default:
		tproxy = internet.SocketConfig_Off
	}

	return &internet.SocketConfig{
		Mark:                 c.Mark,
		Tfo:                  tfoSettings,
		TfoQueueLength:       tfoQueueLength,
		Tproxy:               tproxy,
		AcceptProxyProtocol:  c.AcceptProxyProtocol,
		TcpKeepAliveInterval: c.TCPKeepAliveInterval,
	}, nil
}
