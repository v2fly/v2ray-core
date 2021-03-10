// +build android

package conf

import (
	"context"
	"net"
)

func init() {
	const bootstrapDNS = "8.8.8.8:53"
	var dialer net.Dialer
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(context context.Context, _, _ string) (net.Conn, error) {
			conn, err := dialer.DialContext(context, "udp", bootstrapDNS)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
	newError("Bootstrap DNS: ", bootstrapDNS).AtWarning().WriteToLog()
}
