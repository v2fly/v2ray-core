package dns

import (
	"github.com/v2fly/v2ray-core/v5/common/errors"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/features"
)

// IPOption is an object for IP query options.
type IPOption struct {
	IPv4Enable bool
	IPv6Enable bool
	FakeEnable bool
}

func (opt IPOption) With(other IPOption) IPOption {
	return IPOption{
		IPv4Enable: opt.IPv4Enable && other.IPv4Enable,
		IPv6Enable: opt.IPv6Enable && other.IPv6Enable,
		FakeEnable: opt.FakeEnable && other.FakeEnable,
	}
}

func (opt IPOption) IsValid() bool {
	return opt.IPv4Enable || opt.IPv6Enable
}

// Client is a V2Ray feature for querying DNS information.
//
// v2ray:api:stable
type Client interface {
	features.Feature

	// LookupIP returns IP address for the given domain. IPs may contain IPv4 and/or IPv6 addresses.
	LookupIP(domain string) ([]net.IP, error)
}

// IPv4Lookup is an optional feature for querying IPv4 addresses only.
//
// v2ray:api:beta
type IPv4Lookup interface {
	LookupIPv4(domain string) ([]net.IP, error)
}

// IPv6Lookup is an optional feature for querying IPv6 addresses only.
//
// v2ray:api:beta
type IPv6Lookup interface {
	LookupIPv6(domain string) ([]net.IP, error)
}

// LookupIPWithOption is a helper function for querying DNS information from a dns.Client with dns.IPOption.
//
// v2ray:api:beta
func LookupIPWithOption(client Client, domain string, option IPOption) ([]net.IP, error) {
	if option.FakeEnable {
		if clientWithFakeDNS, ok := client.(ClientWithFakeDNS); ok {
			client = clientWithFakeDNS.AsFakeDNSClient()
		}
	}
	if option.IPv4Enable && !option.IPv6Enable {
		if ipv4Lookup, ok := client.(IPv4Lookup); ok {
			return ipv4Lookup.LookupIPv4(domain)
		} else {
			return nil, errors.New("dns.Client doesn't implement IPv4Lookup")
		}
	}
	if option.IPv6Enable && !option.IPv4Enable {
		if ipv6Lookup, ok := client.(IPv6Lookup); ok {
			return ipv6Lookup.LookupIPv6(domain)
		} else {
			return nil, errors.New("dns.Client doesn't implement IPv6Lookup")
		}
	}
	return client.LookupIP(domain)
}

// ClientType returns the type of Client interface. Can be used for implementing common.HasType.
//
// v2ray:api:beta
func ClientType() interface{} {
	return (*Client)(nil)
}

// ErrEmptyResponse indicates that DNS query succeeded but no answer was returned.
var ErrEmptyResponse = errors.New("empty response")

type RCodeError uint16

func (e RCodeError) Error() string {
	return serial.Concat("rcode: ", uint16(e))
}

func RCodeFromError(err error) uint16 {
	if err == nil {
		return 0
	}
	cause := errors.Cause(err)
	if r, ok := cause.(RCodeError); ok {
		return uint16(r)
	}
	return 0
}
