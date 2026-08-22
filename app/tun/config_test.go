package tun

import (
	"context"
	"testing"

	"github.com/v2fly/v2ray-core/v5/common/session"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestUDPBridgeConfigJSON(t *testing.T) {
	var config Config
	if err := protojson.Unmarshal([]byte(`{
		"name": "packetbridge",
		"mtu": 1500,
		"udp_bridge": {
			"listen_address": "127.0.0.1",
			"listen_port": 9090,
			"peer_address": "127.0.0.1",
			"peer_port": 9091,
			"queue_size": 512
		}
	}`), &config); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	if config.UdpBridge == nil {
		t.Fatal("udp_bridge was not decoded")
	}
	if config.UdpBridge.ListenPort != 9090 || config.UdpBridge.PeerPort != 9091 {
		t.Fatalf("unexpected bridge ports: %+v", config.UdpBridge)
	}
}

func TestUDPBridgeRejectsPreopenedFD(t *testing.T) {
	fd := int32(3)
	tun := new(TUN)
	err := tun.Init(context.Background(), &Config{
		PreopenedFd: &fd,
		UdpBridge:   &UDPBridgeConfig{},
	}, nil, nil)
	if err == nil {
		t.Fatal("expected preopened_fd/udp_bridge conflict")
	}
}

func TestPacketEncodingContextCarriesTUNInboundTag(t *testing.T) {
	tun := &TUN{
		ctx:    context.Background(),
		config: &Config{Tag: "packet-bridge-in"},
	}

	inbound := session.InboundFromContext(tun.packetEncodingContext())
	if inbound == nil {
		t.Fatal("packet encoding context has no inbound session")
	}
	if inbound.Tag != tun.config.Tag {
		t.Fatalf("packet encoding context inbound tag = %q, want %q", inbound.Tag, tun.config.Tag)
	}
}
