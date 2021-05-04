package outbound_test

import (
	"context"
	"testing"
	_ "unsafe"

	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/app/policy"
	. "github.com/v2fly/v2ray-core/v4/app/proxyman/outbound"
	"github.com/v2fly/v2ray-core/v4/app/stats"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/serial"
	"github.com/v2fly/v2ray-core/v4/features/outbound"
	"github.com/v2fly/v2ray-core/v4/proxy/freedom"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
)

func TestInterfaces(t *testing.T) {
	_ = (outbound.Handler)(new(Handler))
	_ = (outbound.Manager)(new(Manager))
}

//go:linkname toContext github.com/v2fly/v2ray-core/v4.toContext
func toContext(ctx context.Context, v *core.Instance) context.Context

func TestOutboundWithoutStatCounter(t *testing.T) {
	config := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&stats.Config{}),
			serial.ToTypedMessage(&policy.Config{
				System: &policy.SystemPolicy{
					Stats: &policy.SystemPolicy_Stats{
						InboundUplink: true,
					},
				},
			}),
		},
	}

	v, _ := core.New(config)
	v.AddFeature((outbound.Manager)(new(Manager)))
	ctx := toContext(context.Background(), v)
	h, _ := NewHandler(ctx, &core.OutboundHandlerConfig{
		Tag:           "tag",
		ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
	})
	conn, _ := h.(*Handler).Dial(ctx, net.TCPDestination(net.DomainAddress("localhost"), 13146))
	_, ok := conn.(*internet.StatCouterConnection)
	if ok {
		t.Errorf("Expected conn to not be StatCouterConnection")
	}
}

func TestOutboundWithStatCounter(t *testing.T) {
	config := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&stats.Config{}),
			serial.ToTypedMessage(&policy.Config{
				System: &policy.SystemPolicy{
					Stats: &policy.SystemPolicy_Stats{
						OutboundUplink:   true,
						OutboundDownlink: true,
					},
				},
			}),
		},
	}

	v, _ := core.New(config)
	v.AddFeature((outbound.Manager)(new(Manager)))
	ctx := toContext(context.Background(), v)
	h, _ := NewHandler(ctx, &core.OutboundHandlerConfig{
		Tag:           "tag",
		ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
	})
	conn, _ := h.(*Handler).Dial(ctx, net.TCPDestination(net.DomainAddress("localhost"), 13146))
	_, ok := conn.(*internet.StatCouterConnection)
	if !ok {
		t.Errorf("Expected conn to be StatCouterConnection")
	}
}
