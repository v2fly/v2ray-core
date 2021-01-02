package conf

import (
	"context"
	"net"
	"testing"
)

func TestBootstrapDNS(t *testing.T) {
	defaultNS := "0.0.0.0:53"
	const domain = "github.com"
	var dialer net.Dialer
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(context context.Context, _, _ string) (net.Conn, error) {
			conn, err := dialer.DialContext(context, "udp", defaultNS)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
	if ips, _ := net.LookupIP(domain); len(ips) > 0 {
		t.Error("set BootstrapDNS failed")
	}
	defaultNS = "8.8.8.8:53"
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(context context.Context, network, address string) (net.Conn, error) {
			conn, err := dialer.DialContext(context, "udp", defaultNS)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
	if ips, err := net.LookupIP(domain); len(ips) == 0 {
		t.Error("set BootstrapDNS failed: ", err)
	}
}
