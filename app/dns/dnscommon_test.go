package dns

import (
	"math/rand"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/miekg/dns"
	"golang.org/x/net/dns/dnsmessage"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/net"
	dns_feature "github.com/v2fly/v2ray-core/v5/features/dns"
)

func Test_parseResponse(t *testing.T) {
	var p [][]byte

	ans := new(dns.Msg)
	ans.Id = 0
	p = append(p, common.Must2(ans.Pack()).([]byte))

	p = append(p, []byte{})

	ans = new(dns.Msg)
	ans.Id = 1
	ans.Answer = append(ans.Answer,
		common.Must2(dns.NewRR("google.com. IN CNAME m.test.google.com")).(dns.RR),
		common.Must2(dns.NewRR("google.com. IN CNAME fake.google.com")).(dns.RR),
		common.Must2(dns.NewRR("google.com. IN A 8.8.8.8")).(dns.RR),
		common.Must2(dns.NewRR("google.com. IN A 8.8.4.4")).(dns.RR),
	)
	p = append(p, common.Must2(ans.Pack()).([]byte))

	ans = new(dns.Msg)
	ans.Id = 2
	ans.Answer = append(ans.Answer,
		common.Must2(dns.NewRR("google.com. IN CNAME m.test.google.com")).(dns.RR),
		common.Must2(dns.NewRR("google.com. IN CNAME fake.google.com")).(dns.RR),
		common.Must2(dns.NewRR("google.com. IN CNAME m.test.google.com")).(dns.RR),
		common.Must2(dns.NewRR("google.com. IN CNAME test.google.com")).(dns.RR),
		common.Must2(dns.NewRR("google.com. IN AAAA 2001::123:8888")).(dns.RR),
		common.Must2(dns.NewRR("google.com. IN AAAA 2001::123:8844")).(dns.RR),
	)
	p = append(p, common.Must2(ans.Pack()).([]byte))

	tests := []struct {
		name    string
		want    *IPRecord
		wantErr bool
	}{
		{
			"empty",
			&IPRecord{0, []net.Address(nil), time.Time{}, dnsmessage.RCodeSuccess},
			false,
		},
		{
			"error",
			nil,
			true,
		},
		{
			"a record",
			&IPRecord{
				1,
				[]net.Address{net.ParseAddress("8.8.8.8"), net.ParseAddress("8.8.4.4")},
				time.Time{},
				dnsmessage.RCodeSuccess,
			},
			false,
		},
		{
			"aaaa record",
			&IPRecord{2, []net.Address{net.ParseAddress("2001::123:8888"), net.ParseAddress("2001::123:8844")}, time.Time{}, dnsmessage.RCodeSuccess},
			false,
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseResponse(p[i])
			if (err != nil) != tt.wantErr {
				t.Errorf("handleResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != nil {
				// reset the time
				got.Expire = time.Time{}
			}
			if cmp.Diff(got, tt.want) != "" {
				t.Errorf("%v", cmp.Diff(got, tt.want))
				// t.Errorf("handleResponse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_buildReqMsgs(t *testing.T) {
	stubID := func() uint16 {
		return uint16(rand.Uint32())
	}
	type args struct {
		domain  string
		option  dns_feature.IPOption
		reqOpts *dnsmessage.Resource
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"dual stack", args{"test.com", dns_feature.IPOption{
			IPv4Enable: true,
			IPv6Enable: true,
			FakeEnable: false,
		}, nil}, 2},
		{"ipv4 only", args{"test.com", dns_feature.IPOption{
			IPv4Enable: true,
			IPv6Enable: false,
			FakeEnable: false,
		}, nil}, 1},
		{"ipv6 only", args{"test.com", dns_feature.IPOption{
			IPv4Enable: false,
			IPv6Enable: true,
			FakeEnable: false,
		}, nil}, 1},
		{"none/error", args{"test.com", dns_feature.IPOption{
			IPv4Enable: false,
			IPv6Enable: false,
			FakeEnable: false,
		}, nil}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildReqMsgs(tt.args.domain, tt.args.option, stubID, tt.args.reqOpts); !(len(got) == tt.want) {
				t.Errorf("buildReqMsgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_genEDNS0Options(t *testing.T) {
	type args struct {
		clientIP net.IP
	}
	tests := []struct {
		name string
		args args
		want *dnsmessage.Resource
	}{
		// TODO: Add test cases.
		{"ipv4", args{net.ParseIP("4.3.2.1")}, nil},
		{"ipv6", args{net.ParseIP("2001::4321")}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genEDNS0Options(tt.args.clientIP); got == nil {
				t.Errorf("genEDNS0Options() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFqdn(t *testing.T) {
	type args struct {
		domain string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"with fqdn", args{"www.v2fly.org."}, "www.v2fly.org."},
		{"without fqdn", args{"www.v2fly.org"}, "www.v2fly.org."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Fqdn(tt.args.domain); got != tt.want {
				t.Errorf("Fqdn() = %v, want %v", got, tt.want)
			}
		})
	}
}
