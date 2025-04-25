package hysteria2

import (
	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/transport/internet"
)

//go:generate go run github.com/ghxhy/v2ray-core/v5/common/errors/errorgen

const (
	protocolName = "hysteria2"
)

func init() {
	common.Must(internet.RegisterProtocolConfigCreator(protocolName, func() interface{} {
		return new(Config)
	}))
}
