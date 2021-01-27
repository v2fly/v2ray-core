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

func (f *FakeDNSServer) QueryIP(ctx context.Context, domain string, clientIP net.IP, option IPOption) ([]net.IP, error) {
	if f.fakeDNSEngine == nil {
		var fakeDNSEngine dns.FakeDNSEngine
		if err := core.RequireFeatures(ctx, func(fdns dns.FakeDNSEngine) {
			fakeDNSEngine = fdns
		}); err != nil {
			return nil, newError("Unable to locate a fake DNS Engine").Base(err).AtError()
		}
	}
	ips := f.fakeDNSEngine.GetFakeIPForDomain(domain)

	netip, err := toNetIP(ips)
	if err != nil {
		return nil, newError("Unable to convert IP to net ip").Base(err).AtError()
	}

	return netip, nil

}
