// +build android

package internet

import (
	"context"
	"fmt"
	gonet "net"
	"testing"

	"github.com/v2fly/v2ray-core/v4/common/net"
)

func TestBootstrapDNS(t *testing.T) {
	if ips, err := gonet.LookupIP("www.google.com"); len(ips) == 0 {
		t.Errorf("failed to lookupIP with BootstrapDNS, %v", err)
	}
}

func TestBootstrapDNSWithV2raySystemDialer(t *testing.T) {
	const bootstrapDNS = "8.8.4.4:53"
	bootstrapDialer := func(ctx context.Context, network, _ string) (gonet.Conn, error) {
		dest, err := net.ParseDestination(fmt.Sprintf("%s:%s", network, bootstrapDNS))
		if err != nil {
			return nil, err
		}
		return DialSystem(ctx, dest, nil)
	}
	UseAlternativeBootstrapDNS(bootstrapDialer)

	TestBootstrapDNS(t)
}
