package specs_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/v2fly/v2ray-core/v5/app/subscription/specs"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	wireguard "github.com/v2fly/v2ray-core/v5/proxy/wireguard/outbound"
)

const wireguardConfigFullName = "v2ray.core.proxy.wireguard.outbound.Config"

func TestWireguardRestrictedSubscriptionRequiresOptIn(t *testing.T) {
	tests := []struct {
		name     string
		settings string
	}{
		{name: "omitted", settings: `{}`},
		{name: "false", settings: `{"restricted":false}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseWireguardSubscription(t, "wireguard", test.settings)
			if err == nil {
				t.Fatal("restricted subscription load unexpectedly succeeded")
			}
			if !strings.Contains(err.Error(), "component has not opted in for load in restricted mode") {
				t.Fatalf("restricted subscription load error = %q, want restricted-mode opt-in error", err)
			}
		})
	}
}

func TestWireguardRestrictedSubscriptionPreservesSettings(t *testing.T) {
	tests := []struct {
		name           string
		encodingJSON   string
		wantEncoding   packetaddr.PacketAddrType
		timeoutJSON    string
		wantTimeout    uint32
		wantTimeoutSet bool
	}{
		{
			name:         "default encoding and timeout",
			wantEncoding: packetaddr.PacketAddrType_None,
		},
		{
			name:           "packet encoding and disabled timeout",
			encodingJSON:   `,"outboundPacketEncoding":"Packet"`,
			wantEncoding:   packetaddr.PacketAddrType_Packet,
			timeoutJSON:    `,"outboundNoReceiveTimeoutSec":0`,
			wantTimeoutSet: true,
		},
		{
			name:           "stream encoding and custom timeout",
			encodingJSON:   `,"outboundPacketEncoding":"Stream"`,
			wantEncoding:   packetaddr.PacketAddrType_Stream,
			timeoutJSON:    `,"outboundNoReceiveTimeoutSec":137`,
			wantTimeout:    137,
			wantTimeoutSet: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			settings := fmt.Sprintf(
				`{"restricted":true,"wgDevice":{"mtu":1412},"domainStrategy":"USE_IP4"%s%s}`,
				test.encodingJSON,
				test.timeoutJSON,
			)
			server, err := parseWireguardSubscription(t, "wireguard", settings)
			if err != nil {
				t.Fatalf("restricted subscription load failed: %v", err)
			}

			message, err := serial.GetInstanceOf(server.Configuration.ProtocolSettings)
			if err != nil {
				t.Fatalf("failed to unpack protocol settings: %v", err)
			}
			config, ok := message.(*wireguard.Config)
			if !ok {
				t.Fatalf("protocol settings type = %T, want *outbound.Config", message)
			}
			if !config.GetRestricted() {
				t.Fatal("restricted flag was not preserved")
			}
			if got := config.GetOutboundPacketEncoding(); got != test.wantEncoding {
				t.Fatalf("outbound packet encoding = %v, want %v", got, test.wantEncoding)
			}
			if got := config.GetDomainStrategy(); got != wireguard.Config_USE_IP4 {
				t.Fatalf("domain strategy = %v, want %v", got, wireguard.Config_USE_IP4)
			}
			if config.GetWgDevice() == nil || config.GetWgDevice().GetMtu() != 1412 {
				t.Fatalf("WireGuard device settings were not preserved: %+v", config.GetWgDevice())
			}
			if got := config.OutboundNoReceiveTimeoutSec != nil; got != test.wantTimeoutSet {
				t.Fatalf("timeout presence = %v, want %v", got, test.wantTimeoutSet)
			}
			if got := config.GetOutboundNoReceiveTimeoutSec(); got != test.wantTimeout {
				t.Fatalf("timeout = %d, want %d", got, test.wantTimeout)
			}
		})
	}
}

func TestWireguardRestrictedSubscriptionFullNameCannotBypassOptIn(t *testing.T) {
	protocol := "#" + wireguardConfigFullName
	if _, err := parseWireguardSubscription(t, protocol, `{}`); err == nil {
		t.Fatal("full protobuf name bypassed restricted loading enforcement")
	} else if !strings.Contains(err.Error(), "component has not opted in for load in restricted mode") {
		t.Fatalf("full protobuf name error = %q, want restricted-mode opt-in error", err)
	}

	server, err := parseWireguardSubscription(t, protocol, `{"restricted":true}`)
	if err != nil {
		t.Fatalf("full protobuf name with explicit opt-in failed: %v", err)
	}
	if got := serial.V2Type(server.Configuration.ProtocolSettings); got != wireguardConfigFullName {
		t.Fatalf("protocol settings type = %q, want %q", got, wireguardConfigFullName)
	}
}

func parseWireguardSubscription(t *testing.T, protocol, settings string) (*specs.SubscriptionServerConfig, error) {
	t.Helper()
	raw := []byte(fmt.Sprintf(`{"protocol":%q,"settings":%s}`, protocol, settings))
	parser := specs.NewOutboundParser()
	outbound, err := parser.ParseOutboundConfig(raw)
	if err != nil {
		return nil, err
	}
	return parser.ToSubscriptionServerConfig(outbound)
}
