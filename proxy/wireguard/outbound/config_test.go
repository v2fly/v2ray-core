package outbound

import (
	"testing"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/common/protoext"
)

func TestConfigProtoJSONOutboundPacketEncoding(t *testing.T) {
	tests := []struct {
		name string
		json string
		want packetaddr.PacketAddrType
	}{
		{name: "omitted", json: `{}`, want: packetaddr.PacketAddrType_None},
		{name: "none", json: `{"outboundPacketEncoding":"None"}`, want: packetaddr.PacketAddrType_None},
		{name: "packet", json: `{"outboundPacketEncoding":"Packet"}`, want: packetaddr.PacketAddrType_Packet},
		{name: "stream", json: `{"outboundPacketEncoding":"Stream"}`, want: packetaddr.PacketAddrType_Stream},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := new(Config)
			if err := protojson.Unmarshal([]byte(test.json), config); err != nil {
				t.Fatalf("failed to unmarshal config: %v", err)
			}
			if got := config.GetOutboundPacketEncoding(); got != test.want {
				t.Fatalf("GetOutboundPacketEncoding() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestConfigProtoJSONOutboundNoReceiveTimeoutPresence(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		wantPresent bool
		want        uint32
	}{
		{name: "omitted", json: `{}`, wantPresent: false},
		{name: "disabled", json: `{"outboundNoReceiveTimeoutSec":0}`, wantPresent: true, want: 0},
		{name: "custom", json: `{"outboundNoReceiveTimeoutSec":120}`, wantPresent: true, want: 120},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := new(Config)
			if err := protojson.Unmarshal([]byte(test.json), config); err != nil {
				t.Fatalf("failed to unmarshal config: %v", err)
			}

			if got := config.OutboundNoReceiveTimeoutSec != nil; got != test.wantPresent {
				t.Fatalf("timeout presence = %v, want %v", got, test.wantPresent)
			}
			if got := config.GetOutboundNoReceiveTimeoutSec(); got != test.want {
				t.Fatalf("GetOutboundNoReceiveTimeoutSec() = %d, want %d", got, test.want)
			}
		})
	}
}

func TestConfigProtoJSONRestricted(t *testing.T) {
	tests := []struct {
		name string
		json string
		want bool
	}{
		{name: "omitted", json: `{}`},
		{name: "false", json: `{"restricted":false}`},
		{name: "true", json: `{"restricted":true}`, want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := new(Config)
			if err := protojson.Unmarshal([]byte(test.json), config); err != nil {
				t.Fatalf("failed to unmarshal config: %v", err)
			}
			if got := config.GetRestricted(); got != test.want {
				t.Fatalf("GetRestricted() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestConfigRestrictedLoadOption(t *testing.T) {
	options, err := protoext.GetMessageOptions((&Config{}).ProtoReflect().Descriptor())
	if err != nil {
		t.Fatalf("failed to get message options: %v", err)
	}
	if got, want := options.GetAllowRestrictedModeLoadIfSet(), "restricted"; got != want {
		t.Fatalf("allow_restricted_mode_load_if_set = %q, want %q", got, want)
	}
}

func TestConfigProtoJSONRestrictedSystemNetworkRejected(t *testing.T) {
	config := new(Config)
	if err := protojson.Unmarshal([]byte(`{"restricted":true,"listenOnSystemNetwork":true}`), config); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}
	if err := validateWireguardConfig(config); err == nil {
		t.Fatal("validateWireguardConfig() accepted restricted system-network config decoded from JSON")
	}
}
