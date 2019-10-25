// +build !confonly

package tcp

import (
	"v2ray.com/core/v4/common"
	"v2ray.com/core/v4/transport/internet"
)

const protocolName = "tcp"

func init() {
	common.Must(internet.RegisterProtocolConfigCreator(protocolName, func() interface{} {
		return new(Config)
	}))
}
