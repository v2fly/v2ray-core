// +build android

package internet

import (
	"context"
	"net"
)

func init() {
	NewDNSResolver = func() *net.Resolver {
		return &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
				const systemDNS = "8.8.8.8:53"
				var dialer net.Dialer
				return dialer.DialContext(ctx, network, systemDNS)
			},
		}
	}
}
