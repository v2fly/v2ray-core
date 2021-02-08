// +build !confonly

package dns

import (
	"context"

	"v2ray.com/core"
	"v2ray.com/core/common/net"
	"v2ray.com/core/features/dns"
)

type FakeDNSServer struct {
	fakeDNSEngine dns.FakeDNSEngine
}

func NewFakeDNSServer() *FakeDNSServer {
	return &FakeDNSServer{}
}

func (f FakeDNSServer) Name() string {
	return "FakeDNS"
}

func (f *FakeDNSServer) QueryIP(ctx context.Context, domain string, _ net.IP, _ IPOption) ([]net.IP, error) {
	if f.fakeDNSEngine == nil {
		if err := core.RequireFeatures(ctx, func(fd dns.FakeDNSEngine) {
			f.fakeDNSEngine = fd
		}); err != nil {
			return nil, newError("Unable to locate a fake DNS Engine").Base(err).AtError()
		}
	}
	ips := f.fakeDNSEngine.GetFakeIPForDomain(domain)

	netIP, err := toNetIP(ips)
	if err != nil {
		return nil, newError("Unable to convert IP to net ip").Base(err).AtError()
	}

	return netIP, nil
}
