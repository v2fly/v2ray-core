package tagged

import (
	"context"

	"github.com/v2fly/v2ray-core/v4/common/net"
)

type DialFunc func(ctx context.Context, dest net.Destination, tag string) (net.Conn, error)

var Dialer DialFunc
