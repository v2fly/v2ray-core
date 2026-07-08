//go:build linux && !confonly
// +build linux,!confonly

package socks5ify

import (
	"fmt"
	stdnet "net"

	"golang.org/x/sys/unix"
	"google.golang.org/protobuf/types/known/anypb"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/dispatcher"
	applog "github.com/v2fly/v2ray-core/v5/app/log"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	tunapp "github.com/v2fly/v2ray-core/v5/app/tun"
	clog "github.com/v2fly/v2ray-core/v5/common/log"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/proxy/socks"
)

func startV2RayWithTun(opts parentOptions, child childConfig, tunFD int) (*core.Instance, error) {
	if int(int32(tunFD)) != tunFD {
		_ = unix.Close(tunFD)
		return nil, fmt.Errorf("invalid TUN fd %d", tunFD)
	}

	config := buildCoreConfig(opts, child, int32(tunFD))
	server, err := core.New(config)
	if err != nil {
		_ = unix.Close(tunFD)
		return nil, err
	}
	if err := server.Start(); err != nil {
		_ = server.Close()
		return nil, err
	}
	return server, nil
}

func buildCoreConfig(opts parentOptions, child childConfig, tunFD int32) *core.Config {
	tunConfig := &tunapp.Config{
		Name:                  child.TunName,
		Mtu:                   uint32(child.MTU),
		PreopenedFd:           &tunFD,
		Tag:                   "socks5ify-tun",
		PacketEncoding:        packetaddr.PacketAddrType_Packet,
		EnablePromiscuousMode: true,
		EnableSpoofing:        true,
		Ips: []*routercommon.CIDR{
			cidr(child.IPv4.Host, child.IPv4.Prefix),
		},
		Routes: []*routercommon.CIDR{
			cidr("0.0.0.0", 0),
		},
	}
	if child.IPv6 {
		tunConfig.Ips = append(tunConfig.Ips, cidr(child.IPv6Config.Host, child.IPv6Config.Prefix))
		tunConfig.Routes = append(tunConfig.Routes, cidr("::", 0))
	}
	apps := []*anypb.Any{
		serial.ToTypedMessage(&dispatcher.Config{}),
		serial.ToTypedMessage(&proxyman.InboundConfig{}),
		serial.ToTypedMessage(&proxyman.OutboundConfig{}),
		serial.ToTypedMessage(tunConfig),
	}
	if opts.Quiet {
		apps = append([]*anypb.Any{serial.ToTypedMessage(quietLogConfig())}, apps...)
	}

	socksConfig := &socks.ClientConfig{
		Version: socks.Version_SOCKS5,
		Server: []*protocol.ServerEndpoint{
			{
				Address: v2net.NewIPOrDomain(v2net.ParseAddress(opts.SOCKS.Host)),
				Port:    opts.SOCKS.Port,
				User:    socksUsers(opts.SOCKS),
			},
		},
	}

	return &core.Config{
		App: apps,
		Outbound: []*core.OutboundHandlerConfig{
			{
				Tag:           "socks5ify-socks",
				ProxySettings: serial.ToTypedMessage(socksConfig),
			},
		},
	}
}

func quietLogConfig() *applog.Config {
	return &applog.Config{
		Error: &applog.LogSpecification{
			Type:  applog.LogType_Console,
			Level: clog.Severity_Error,
		},
		Access: &applog.LogSpecification{
			Type: applog.LogType_None,
		},
	}
}

func socksUsers(server socksServer) []*protocol.User {
	if server.Username == "" && server.Password == "" {
		return nil
	}
	return []*protocol.User{
		{
			Account: serial.ToTypedMessage(&socks.Account{
				Username: server.Username,
				Password: server.Password,
			}),
		},
	}
}

func cidr(ipText string, prefix int) *routercommon.CIDR {
	ip := stdnet.ParseIP(ipText)
	if ip4 := ip.To4(); ip4 != nil {
		return &routercommon.CIDR{Ip: ip4, Prefix: uint32(prefix)}
	}
	return &routercommon.CIDR{Ip: ip.To16(), Prefix: uint32(prefix)}
}
