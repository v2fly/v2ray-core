package webrtc

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	pionwebrtc "github.com/pion/webrtc/v4"
)

func TestCandidateBlastIP(t *testing.T) {
	raw, err := json.Marshal(pionwebrtc.ICECandidateInit{
		Candidate: "candidate:2333951359 1 udp 1694498815 80.233.61.13 21946 typ srflx raddr 192.0.2.1 rport 40002 ufrag test",
	})
	if err != nil {
		t.Fatal(err)
	}

	ip, err := candidateBlossomIP(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := ip.String(), "80.233.61.13"; got != want {
		t.Fatalf("candidateBlossomIP() = %s, want %s", got, want)
	}
}

func TestPortBlossomPayload(t *testing.T) {
	payload := portBlossomPayload(21946)
	if got, want := len(payload), 2; got != want {
		t.Fatalf("len(portBlossomPayload()) = %d, want %d", got, want)
	}
	if got, want := int(payload[0])<<8|int(payload[1]), 21946; got != want {
		t.Fatalf("portBlossomPayload() = %d, want %d", got, want)
	}
}

func TestPacketConnSupportsIP(t *testing.T) {
	tests := []struct {
		name    string
		localIP net.IP
		target  net.IP
		want    bool
	}{
		{
			name:    "ipv4 matches ipv4",
			localIP: net.ParseIP("0.0.0.0"),
			target:  net.ParseIP("198.51.100.1"),
			want:    true,
		},
		{
			name:    "wildcard ipv4 rejects ipv6",
			localIP: net.ParseIP("0.0.0.0"),
			target:  net.ParseIP("2001:db8::1"),
			want:    false,
		},
		{
			name:    "bound ipv4 rejects ipv6",
			localIP: net.ParseIP("192.0.2.1"),
			target:  net.ParseIP("2001:db8::1"),
			want:    false,
		},
		{
			name:    "bound ipv6 matches ipv6",
			localIP: net.ParseIP("2001:db8::2"),
			target:  net.ParseIP("2001:db8::1"),
			want:    true,
		},
		{
			name:    "wildcard ipv6 rejects ipv4",
			localIP: net.ParseIP("::"),
			target:  net.ParseIP("198.51.100.1"),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := fakePacketConn{addr: &net.UDPAddr{IP: tt.localIP, Port: 29000}}
			if got := packetConnSupportsIP(conn, tt.target); got != tt.want {
				t.Fatalf("packetConnSupportsIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

type fakePacketConn struct {
	addr net.Addr
}

func (f fakePacketConn) ReadFrom(_ []byte) (int, net.Addr, error)  { return 0, nil, nil }
func (f fakePacketConn) WriteTo(_ []byte, _ net.Addr) (int, error) { return 0, nil }
func (f fakePacketConn) Close() error                              { return nil }
func (f fakePacketConn) LocalAddr() net.Addr                       { return f.addr }
func (f fakePacketConn) SetDeadline(_ time.Time) error             { return nil }
func (f fakePacketConn) SetReadDeadline(_ time.Time) error         { return nil }
func (f fakePacketConn) SetWriteDeadline(_ time.Time) error        { return nil }
