// +build android

package internet

import (
	"context"
	"net"
)

type DialerFunc func(context.Context, string, string) (net.Conn, error)

func UseAlternativeBootstrapDNS(dialer DialerFunc) {
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: dialer,
	}
}

func init() {
	const bootstrapDNS = "8.8.8.8:53"
	bootstrapDialer := func(ctx context.Context, network, _ string) (net.Conn, error) {
		var dialer net.Dialer
		return dialer.DialContext(ctx, network, bootstrapDNS)
	}
	UseAlternativeBootstrapDNS(bootstrapDialer)

	newError("Android Bootstrap DNS: ", bootstrapDNS).AtWarning().WriteToLog()
}
