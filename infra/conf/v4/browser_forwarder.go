package v4

import (
	"strings"

	"github.com/golang/protobuf/proto"

	"github.com/ghxhy/v2ray-core/v5/app/browserforwarder"
)

type BrowserForwarderConfig struct {
	ListenAddr string `json:"listenAddr"`
	ListenPort int32  `json:"listenPort"`
}

func (b *BrowserForwarderConfig) Build() (proto.Message, error) {
	b.ListenAddr = strings.TrimSpace(b.ListenAddr)
	if b.ListenAddr != "" && b.ListenPort == 0 {
		b.ListenPort = 54321
	}
	return &browserforwarder.Config{
		ListenAddr: b.ListenAddr,
		ListenPort: b.ListenPort,
	}, nil
}
