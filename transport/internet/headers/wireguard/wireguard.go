package wireguard

import (
	"context"

	"github.com/ghxhy/v2ray-core/v5/common"
)

type Wireguard struct{}

func (Wireguard) Size() int32 {
	return 4
}

// Serialize implements PacketHeader.
func (Wireguard) Serialize(b []byte) {
	b[0] = 0x04
	b[1] = 0x00
	b[2] = 0x00
	b[3] = 0x00
}

// NewWireguard returns a new VideoChat instance based on given config.
func NewWireguard(ctx context.Context, config interface{}) (interface{}, error) {
	return Wireguard{}, nil
}

func init() {
	common.Must(common.RegisterConfig((*WireguardConfig)(nil), NewWireguard))
}
