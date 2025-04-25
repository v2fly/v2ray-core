package dtls

import (
	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/transport/internet"
)

//go:generate go run github.com/ghxhy/v2ray-core/v5/common/errors/errorgen

const protocolName = "dtls"

func init() {
	common.Must(internet.RegisterProtocolConfigCreator(protocolName, func() interface{} {
		return new(Config)
	}))
}
