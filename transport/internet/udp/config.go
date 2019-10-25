package udp

import (
	"v2ray.com/core/v4/common"
	"v2ray.com/core/v4/transport/internet"
)

func init() {
	common.Must(internet.RegisterProtocolConfigCreator(protocolName, func() interface{} {
		return new(Config)
	}))
}
