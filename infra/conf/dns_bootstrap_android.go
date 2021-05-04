// +build android

package conf

import (
	"context"
	"net"
)

type DialerFunc func(context.Context, string, string) (net.Conn, error)

var BootstrapDialer DialerFunc = func(ctx context.Context, network, _ string) (net.Conn, error) {
	var dialer net.Dialer
	return dialer.DialContext(ctx, network, BootstrapDNS)
}

var BootstrapDNS = "8.8.8.8:53"

func init() {
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: BootstrapDialer,
	}
	newError("Android Bootstrap DNS: ", BootstrapDNS).AtWarning().WriteToLog()
}
