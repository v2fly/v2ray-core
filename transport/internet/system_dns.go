package internet

import (
	"net"
)

type DNSResolverFunc func() *net.Resolver

var NewDNSResolver DNSResolverFunc = func() *net.Resolver {
	return nil
}
