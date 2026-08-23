package tun

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/v2fly/v2ray-core/v5/common/session"
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

func TestPacketEncodingBypassPortsJSON(t *testing.T) {
	var config Config
	if err := protojson.Unmarshal([]byte(`{
		"packet_encoding": "Stream",
		"packet_encoding_bypass_ports": [53, 123]
	}`), &config); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	if got := config.PacketEncodingBypassPorts; len(got) != 2 || got[0] != 53 || got[1] != 123 {
		t.Fatalf("packet_encoding_bypass_ports = %v, want [53 123]", got)
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

func TestPacketEncodingBypassPortsValidation(t *testing.T) {
	for _, port := range []uint32{0, 65536} {
		t.Run(fmt.Sprint(port), func(t *testing.T) {
			tun := new(TUN)
			err := tun.Init(context.Background(), &Config{
				PacketEncodingBypassPorts: []uint32{port},
			}, nil, nil)
			if err == nil {
				t.Fatalf("expected packet_encoding_bypass_ports value %d to be rejected", port)
			}
		})
	}
}
