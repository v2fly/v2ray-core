package encoding

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type testGunService struct {
	ctx context.Context
}

func (s testGunService) Context() context.Context {
	return s.ctx
}

func (testGunService) Send(*Hunk) error {
	return nil
}

func (testGunService) Recv() (*Hunk, error) {
	return nil, nil
}

func TestNewGunConnUsesPeerAddrByDefault(t *testing.T) {
	ctx := peer.NewContext(context.Background(), &peer.Peer{
		Addr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234},
	})
	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("x-forwarded-for", "203.0.113.8"))

	conn := NewGunConn(testGunService{ctx: ctx}, nil, false)
	if got := conn.RemoteAddr().String(); got != "127.0.0.1:1234" {
		t.Fatalf("unexpected remote addr: %s", got)
	}
}

func TestNewGunConnUsesXForwardedForWhenEnabled(t *testing.T) {
	ctx := peer.NewContext(context.Background(), &peer.Peer{
		Addr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234},
	})
	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("x-forwarded-for", "203.0.113.8, 198.51.100.9"))

	conn := NewGunConn(testGunService{ctx: ctx}, nil, true)
	if got := conn.RemoteAddr().String(); got != "203.0.113.8:0" {
		t.Fatalf("unexpected remote addr: %s", got)
	}
}
