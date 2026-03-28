package dtls

import (
	"testing"

	v2net "github.com/v2fly/v2ray-core/v5/common/net"
)

func TestDTLSConnWrappedCloseIsIdempotent(t *testing.T) {
	parent := &Listener{
		sessions: make(map[ConnectionID]*dTLSConnWrapped),
	}
	src := v2net.UDPDestination(v2net.LocalHostIP, 12345)

	finishCalls := 0
	wrapped := &dTLSConnWrapped{
		unencryptedConn: &dTLSConn{
			src:    src,
			parent: parent,
			finish: func() {
				finishCalls++
			},
		},
	}
	parent.sessions[ConnectionID{Remote: src.Address, Port: src.Port}] = wrapped

	if err := wrapped.Close(); err != nil {
		t.Fatal(err)
	}
	if err := wrapped.Close(); err != nil {
		t.Fatal(err)
	}
	if finishCalls != 1 {
		t.Fatalf("unexpected finish call count: %d", finishCalls)
	}
	if len(parent.sessions) != 0 {
		t.Fatalf("expected session removal, got %d sessions", len(parent.sessions))
	}
}
