package conf

import (
	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v4/app/browserforwarder"
)

type BrowserForwarderConfig struct {
	ListenAddr string `json:"listenAddr"`
	ListenPort int32  `json:"listenPort"`
}

func (b BrowserForwarderConfig) Build() (proto.Message, error) {
	return &browserforwarder.Config{
		ListenAddr: b.ListenAddr,
		ListenPort: b.ListenPort,
	}, nil
}
