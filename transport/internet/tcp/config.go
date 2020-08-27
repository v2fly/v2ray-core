// +build !confonly

package tcp

import (
	"github.com/v2fly/v2ray-core/common"
	"github.com/v2fly/v2ray-core/transport/internet"
)

const protocolName = "tcp"

func init() {
	common.Must(internet.RegisterProtocolConfigCreator(protocolName, func() interface{} {
		return new(Config)
	}))
}
