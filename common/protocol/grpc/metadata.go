package grpc

import (
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

// ParseXForwardedFor parses X-Forwarded-For metadata in gRPC metadata, and return the IP list in it.
func ParseXForwardedFor(md metadata.MD) []net.Address {
	xff := md.Get("X-Forwarded-For")
	if len(xff) == 0 {
		return nil
	}
	list := strings.Split(xff[0], ",")
	addrs := make([]net.Address, 0, len(list))
	for _, proxy := range list {
		addrs = append(addrs, net.ParseAddress(proxy))
	}
	return addrs
}
