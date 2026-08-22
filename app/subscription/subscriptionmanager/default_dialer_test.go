package subscriptionmanager

import (
	"testing"

	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	"github.com/v2fly/v2ray-core/v5/app/subscription"
	"github.com/v2fly/v2ray-core/v5/app/subscription/specs"
	"github.com/v2fly/v2ray-core/v5/common/serial"
)

func TestMaterializeAppliesDefaultDialerTag(t *testing.T) {
	manager := &SubscriptionManagerImpl{
		config: &subscription.Config{DefaultDialerTag: "packetbridge-bypass"},
	}
	outbound, err := manager.materialize("test", "subscription_test", &specs.SubscriptionServerConfig{
		Configuration: &specs.ServerConfiguration{},
	})
	if err != nil {
		t.Fatalf("materialize failed: %v", err)
	}

	senderInstance, err := serial.GetInstanceOf(outbound.SenderSettings)
	if err != nil {
		t.Fatalf("failed to decode sender settings: %v", err)
	}
	sender := senderInstance.(*proxyman.SenderConfig)
	if sender.ProxySettings == nil {
		t.Fatal("proxy settings were not applied")
	}
	if sender.ProxySettings.Tag != "packetbridge-bypass" {
		t.Fatalf("unexpected proxy tag: %q", sender.ProxySettings.Tag)
	}
	if !sender.ProxySettings.TransportLayerProxy {
		t.Fatal("transport-layer proxy was not enabled")
	}
}

func TestMaterializePreservesLegacyBehaviorWithoutDefault(t *testing.T) {
	manager := &SubscriptionManagerImpl{config: &subscription.Config{}}
	outbound, err := manager.materialize("test", "subscription_test", &specs.SubscriptionServerConfig{
		Configuration: &specs.ServerConfiguration{},
	})
	if err != nil {
		t.Fatalf("materialize failed: %v", err)
	}

	senderInstance, err := serial.GetInstanceOf(outbound.SenderSettings)
	if err != nil {
		t.Fatalf("failed to decode sender settings: %v", err)
	}
	if senderInstance.(*proxyman.SenderConfig).ProxySettings != nil {
		t.Fatal("unexpected proxy settings without default_dialer_tag")
	}
}

func TestImportSourceDefaultDialerPrecedence(t *testing.T) {
	original := &subscription.ImportSource{ImportUsingTag: ""}
	effective := importSourceWithDefaultDialer(original, "packetbridge-bypass")
	if effective.ImportUsingTag != "packetbridge-bypass" {
		t.Fatalf("default tag was not applied: %q", effective.ImportUsingTag)
	}
	if original.ImportUsingTag != "" {
		t.Fatal("source was mutated")
	}

	explicit := &subscription.ImportSource{ImportUsingTag: "explicit"}
	effective = importSourceWithDefaultDialer(explicit, "packetbridge-bypass")
	if effective.ImportUsingTag != "explicit" {
		t.Fatalf("explicit tag was overwritten: %q", effective.ImportUsingTag)
	}
}
