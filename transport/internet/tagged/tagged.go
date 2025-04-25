package tagged

import (
	"context"

	"github.com/ghxhy/v2ray-core/v5/common/net"
)

type DialFunc func(ctx context.Context, dest net.Destination, tag string) (net.Conn, error)

var Dialer DialFunc
