package localdns

import (
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/features/dns"
)

// Client is an implementation of dns.Client, which queries localhost for DNS.
type Client struct {
	ipOption *dns.IPOption
}

// Type implements common.HasType.
func (*Client) Type() interface{} {
	return dns.ClientType()
}

// Start implements common.Runnable.
func (*Client) Start() error { return nil }

// Close implements common.Closable.
func (*Client) Close() error { return nil }

// GetIPOption implements Client.
func (c *Client) GetIPOption() *dns.IPOption {
	return c.ipOption
}

// SetIPOption implements Client.
func (c *Client) SetIPOption(isIPv4Enable, isIPv6Enable bool) {
	c.ipOption.IPv4Enable = isIPv4Enable
	c.ipOption.IPv6Enable = isIPv6Enable
}

// SetFakeDNSOption implements Client.
func (c *Client) SetFakeDNSOption(isFakeEnable bool) {
	c.ipOption.FakeEnable = isFakeEnable
}

// LookupIP implements Client.
func (c *Client) LookupIP(host string) ([]net.IP, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	parsedIPs := make([]net.IP, 0, len(ips))
	ipv4 := make([]net.IP, 0, len(ips))
	ipv6 := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		parsed := net.IPAddress(ip)
		if parsed != nil {
			parsedIPs = append(parsedIPs, parsed.IP())
		}
		if len(ip) == net.IPv4len {
			ipv4 = append(ipv4, ip)
		}
		if len(ip) == net.IPv6len {
			ipv6 = append(ipv6, ip)
		}
	}
	switch {
	case c.ipOption.IPv4Enable && c.ipOption.IPv6Enable:
		if len(parsedIPs) > 0 {
			return parsedIPs, nil
		}
	case c.ipOption.IPv4Enable:
		if len(ipv4) > 0 {
			return ipv4, nil
		}
	case c.ipOption.IPv6Enable:
		if len(ipv6) > 0 {
			return ipv6, nil
		}
	}
	return nil, dns.ErrEmptyResponse
}

// New create a new dns.Client that queries localhost for DNS.
func New(option *dns.IPOption) *Client {
	return &Client{
		ipOption: option,
	}
}
