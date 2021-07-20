package grpc_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/metadata"

	"github.com/v2fly/v2ray-core/v4/common/net"
	. "github.com/v2fly/v2ray-core/v4/common/protocol/grpc"
)

func TestParseXForwardedFor(t *testing.T) {
	md := metadata.Pairs("X-Forwarded-For", "129.78.138.66, 129.78.64.103")
	addrs := ParseXForwardedFor(md)
	if r := cmp.Diff(addrs, []net.Address{net.ParseAddress("129.78.138.66"), net.ParseAddress("129.78.64.103")}); r != "" {
		t.Error(r)
	}
}
