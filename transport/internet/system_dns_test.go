package internet

import (
	"context"
	"fmt"
	gonet "net"
	"testing"

	"github.com/v2fly/v2ray-core/v4/common/net"
)

func TestDNSResolver(t *testing.T) {
	resolver := NewDNSResolver()
	if ips, err := resolver.LookupIP(context.Background(), "tcp", "www.google.com"); err != nil {
		t.Errorf("failed to lookupIP with BootstrapDNS, %v, %v", ips, err)
	}
}

func TestSystemDNSResolver(t *testing.T) {
	NewDNSResolver = func() *gonet.Resolver {
		return &gonet.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, _ string) (gonet.Conn, error) {
				const systemDNS = "8.8.8.8:53"
				dest, err := net.ParseDestination(fmt.Sprintf("%s:%s", network, systemDNS))
				if err != nil {
					return nil, err
				}
				return DialSystem(ctx, dest, nil)
			},
		}
	}

	TestDNSResolver(t)
}
