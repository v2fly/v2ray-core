// +build android

package conf

import (
	"context"
	"fmt"
	gonet "net"
	"testing"

	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
)

func TestBootstrapDNS(t *testing.T) {
	if ips, err := gonet.LookupIP("www.google.com"); len(ips) == 0 {
		t.Errorf("failed to lookupIP with BootstrapDNS, %v", err)
	}
}

func TestBootstrapDNSWithV2raySystemDialer(t *testing.T) {
	BootstrapDialer := func(ctx context.Context, network, _ string) (gonet.Conn, error) {
		dest, err := net.ParseDestination(fmt.Sprintf("%s:%s", network, BootstrapDNS))
		if err != nil {
			return nil, err
		}
		return internet.DialSystem(ctx, dest, nil)
	}
	UseAlternativeBootstrapDNS(BootstrapDialer)

	TestBootstrapDNS(t *testing.T)
}
