package dns_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	. "v2ray.com/core/app/dns"
	"v2ray.com/core/common"
	"v2ray.com/core/common/net"
)

func TestQUICNameServer(t *testing.T) {
	url, err := url.Parse("quic://dns.adguard.com")
	common.Must(err)
	s, err := NewQUICNameServer(url)
	common.Must(err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	ips, err := s.QueryIP(ctx, "google.com", net.IP(nil), IPOption{
		IPv4Enable: true,
		IPv6Enable: true,
	})
	cancel()
	common.Must(err)
	if len(ips) == 0 {
		t.Error("expect some ips, but got 0")
	}
}
