// +build !confonly

package quic

import (
	"github.com/v2fly/v2ray-core/common"
	"github.com/v2fly/v2ray-core/transport/internet"
)

//go:generate errorgen

// Here is some modification needs to be done before update quic vendor.
// * use bytespool in buffer_pool.go
// * set MaxReceivePacketSize to 1452 - 32 (16 bytes auth, 16 bytes head)
//
//

const protocolName = "quic"
const internalDomain = "quic.internal.v2fly.org"

func init() {
	common.Must(internet.RegisterProtocolConfigCreator(protocolName, func() interface{} {
		return new(Config)
	}))
}
