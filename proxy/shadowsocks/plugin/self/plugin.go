package self

import (
	"flag"
	"fmt"
	"os/user"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/app/dispatcher"
	vlog "github.com/v2fly/v2ray-core/v4/app/log"
	"github.com/v2fly/v2ray-core/v4/app/proxyman"
	clog "github.com/v2fly/v2ray-core/v4/common/log"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/platform/filesystem"
	"github.com/v2fly/v2ray-core/v4/common/protocol"
	"github.com/v2fly/v2ray-core/v4/common/serial"
	"github.com/v2fly/v2ray-core/v4/proxy/dokodemo"
	"github.com/v2fly/v2ray-core/v4/proxy/freedom"
	"github.com/v2fly/v2ray-core/v4/proxy/shadowsocks"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
	"github.com/v2fly/v2ray-core/v4/transport/internet/quic"
	"github.com/v2fly/v2ray-core/v4/transport/internet/tls"
	"github.com/v2fly/v2ray-core/v4/transport/internet/websocket"
)

//go:generate go run github.com/v2fly/self-core/v4/common/errors/errorgen

var _ shadowsocks.SIP003Plugin = (*Plugin)(nil)

func init() {
	shadowsocks.RegisterPlugin("v2ray-plugin", func() shadowsocks.SIP003Plugin {
		return &Plugin{}
	})
}

type Plugin struct {
	instance *core.Instance
}

func (v *Plugin) Init(localHost string, localPort string, remoteHost string, remotePort string, pluginOpts string, pluginArgs []string, _ *shadowsocks.MemoryAccount) error {
	opts := make(Args)

	opts.Add("localAddr", localHost)
	opts.Add("localPort", localPort)
	opts.Add("remoteAddr", remoteHost)
	opts.Add("remotePort", remotePort)

	if len(pluginOpts) > 0 {
		otherOpts, err := ParsePluginOptions(pluginOpts)
		if err != nil {
			return err
		}
		for k, v := range otherOpts {
			opts[k] = v
		}
	}

	config, err := v.init(opts, pluginArgs)
	if err != nil {
		return newError("create config for v2ray-plugin").Base(err)
	}

	instance, err := core.New(config)
	if err != nil {
		return newError("create core for v2ray-plugin").Base(err)
	}

	err = instance.Start()

	if err != nil {
		return newError("start core for v2ray-plugin").Base(err)
	}

	v.instance = instance

	return nil
}

func (v *Plugin) init(opts Args, pluginArgs []string) (*core.Config, error) {
	flag := flag.NewFlagSet("v2ray-plugin", flag.ContinueOnError)
	var (
		fastOpen   = flag.Bool("fast-open", false, "Enable TCP fast open.")
		localAddr  = flag.String("localAddr", "127.0.0.1", "local address to listen on.")
		localPort  = flag.String("localPort", "1984", "local port to listen on.")
		remoteAddr = flag.String("remoteAddr", "127.0.0.1", "remote address to forward.")
		remotePort = flag.String("remotePort", "1080", "remote port to forward.")
		path       = flag.String("path", "/", "URL path for websocket.")
		host       = flag.String("host", "cloudfront.com", "Hostname for server.")
		tlsEnabled = flag.Bool("tls", false, "Enable TLS.")
		cert       = flag.String("cert", "", "Path to TLS certificate file. Overrides certRaw. Default: ~/.acme.sh/{host}/fullchain.cer")
		certRaw    = flag.String("certRaw", "", "Raw TLS certificate content. Intended only for Android.")
		key        = flag.String("key", "", "(server) Path to TLS key file. Default: ~/.acme.sh/{host}/{host}.key")
		mode       = flag.String("mode", "websocket", "Transport mode: websocket, quic (enforced tls).")
		mux        = flag.Int("mux", 1, "Concurrent multiplexed connections (websocket client mode only).")
		server     = flag.Bool("server", false, "Run in server mode")
		logLevel   = flag.String("loglevel", "", "loglevel for self: debug, info, warning (default), error, none.")
		fwmark     = flag.Int("fwmark", 0, "Set SO_MARK option for outbound sockets.")
	)

	if err := flag.Parse(pluginArgs); err != nil {
		return nil, newError("failed to parse plugin args").Base(err)
	}

	if c, b := opts.Get("mode"); b {
		*mode = c
	}
	if c, b := opts.Get("mux"); b {
		if i, err := strconv.Atoi(c); err == nil {
			*mux = i
		} else {
			newError("failed to parse mux, use default value").AtWarning().WriteToLog()
		}
	}
	if _, b := opts.Get("tls"); b {
		*tlsEnabled = true
	}
	if c, b := opts.Get("host"); b {
		*host = c
	}
	if c, b := opts.Get("path"); b {
		*path = c
	}
	if c, b := opts.Get("cert"); b {
		*cert = c
	}
	if c, b := opts.Get("certRaw"); b {
		*certRaw = c
	}
	if c, b := opts.Get("key"); b {
		*key = c
	}
	if c, b := opts.Get("loglevel"); b {
		*logLevel = c
	}
	if _, b := opts.Get("server"); b {
		*server = true
	}
	if c, b := opts.Get("localAddr"); b {
		if *server {
			*remoteAddr = c
		} else {
			*localAddr = c
		}
	}
	if c, b := opts.Get("localPort"); b {
		if *server {
			*remotePort = c
		} else {
			*localPort = c
		}
	}
	if c, b := opts.Get("remoteAddr"); b {
		if *server {
			*localAddr = c
		} else {
			*remoteAddr = c
		}
	}
	if c, b := opts.Get("remotePort"); b {
		if *server {
			*localPort = c
		} else {
			*remotePort = c
		}
	}

	if _, b := opts.Get("fastOpen"); b {
		*fastOpen = true
	}

	if c, b := opts.Get("fwmark"); b {
		if i, err := strconv.Atoi(c); err == nil {
			*fwmark = i
		} else {
			newError("failed to parse fwmark, use default value").AtWarning().WriteToLog()
		}
	}

	lport, err := net.PortFromString(*localPort)
	if err != nil {
		return nil, newError("invalid localPort:", *localPort).Base(err)
	}
	rport, err := strconv.ParseUint(*remotePort, 10, 32)
	if err != nil {
		return nil, newError("invalid remotePort:", *remotePort).Base(err)
	}
	outboundProxy := serial.ToTypedMessage(&freedom.Config{
		DestinationOverride: &freedom.DestinationOverride{
			Server: &protocol.ServerEndpoint{
				Address: net.NewIPOrDomain(net.ParseAddress(*remoteAddr)),
				Port:    uint32(rport),
			},
		},
	})

	var transportSettings proto.Message
	var connectionReuse bool
	switch *mode {
	case "websocket":
		transportSettings = &websocket.Config{
			Path: *path,
			Header: []*websocket.Header{
				{Key: "Host", Value: *host},
			},
		}
		if *mux != 0 {
			connectionReuse = true
		}
	case "quic":
		transportSettings = &quic.Config{
			Security: &protocol.SecurityConfig{Type: protocol.SecurityType_NONE},
		}
		*tlsEnabled = true
	default:
		return nil, newError("unsupported mode:", *mode)
	}

	streamConfig := internet.StreamConfig{
		ProtocolName: *mode,
		TransportSettings: []*internet.TransportConfig{{
			ProtocolName: *mode,
			Settings:     serial.ToTypedMessage(transportSettings),
		}},
	}
	if *fastOpen || *fwmark != 0 {
		socketConfig := &internet.SocketConfig{}
		if *fastOpen {
			socketConfig.Tfo = internet.SocketConfig_Enable
		}
		if *fwmark != 0 {
			socketConfig.Mark = uint32(int32(*fwmark))
		}

		streamConfig.SocketSettings = socketConfig
	}
	if *tlsEnabled {
		tlsConfig := tls.Config{ServerName: *host}
		if *server {
			certificate := tls.Certificate{}
			if *cert == "" && *certRaw == "" {
				usr, err := user.Current()
				if err != nil {
					return nil, err
				}

				*cert = fmt.Sprintf("%s/.acme.sh/%s/fullchain.cer", usr.HomeDir, *host)
				newError("No TLS cert specified, trying ", *cert).AtWarning().WriteToLog()
			}

			if *cert != "" {
				certificate.Certificate, err = filesystem.ReadFile(*cert)
			}
			if *certRaw != "" {
				certHead := "-----BEGIN CERTIFICATE-----"
				certTail := "-----END CERTIFICATE-----"
				fixedCert := certHead + "\n" + *certRaw + "\n" + certTail
				certificate.Certificate = []byte(fixedCert)
			}
			if err != nil {
				return nil, newError("failed to read cert").Base(err)
			}
			if *key == "" {
				usr, err := user.Current()
				if err != nil {
					return nil, err
				}
				*key = fmt.Sprintf("%[1]s/.acme.sh/%[2]s/%[2]s.key", usr.HomeDir, *host)
				newError("No TLS key specified, trying ", *key).AtWarning().WriteToLog()
			}
			certificate.Key, err = filesystem.ReadFile(*key)
			if err != nil {
				return nil, newError("failed to read key file").Base(err)
			}
			tlsConfig.Certificate = []*tls.Certificate{&certificate}
		} else if *cert != "" || *certRaw != "" {
			certificate := tls.Certificate{Usage: tls.Certificate_AUTHORITY_VERIFY}
			if *cert != "" {
				certificate.Certificate, err = filesystem.ReadFile(*cert)
			}
			if *certRaw != "" {
				certHead := "-----BEGIN CERTIFICATE-----"
				certTail := "-----END CERTIFICATE-----"
				fixedCert := certHead + "\n" + *certRaw + "\n" + certTail
				certificate.Certificate = []byte(fixedCert)
			}
			if err != nil {
				return nil, newError("failed to read cert").Base(err)
			}
			tlsConfig.Certificate = []*tls.Certificate{&certificate}
		}
		streamConfig.SecurityType = serial.GetMessageType(&tlsConfig)
		streamConfig.SecuritySettings = []*anypb.Any{serial.ToTypedMessage(&tlsConfig)}
	}

	apps := []*anypb.Any{
		serial.ToTypedMessage(&dispatcher.Config{}),
		serial.ToTypedMessage(&proxyman.InboundConfig{}),
		serial.ToTypedMessage(&proxyman.OutboundConfig{}),
		serial.ToTypedMessage(LogConfig(*logLevel)),
	}

	var config *core.Config

	if *server {
		proxyAddress := net.LocalHostIP
		if connectionReuse {
			// This address is required when mux is used on client.
			// dokodemo is not aware of mux connections by itself.
			proxyAddress = net.ParseAddress("v1.mux.cool")
		}
		localAddrs := ParseLocalAddr(*localAddr)
		inbounds := make([]*core.InboundHandlerConfig, len(localAddrs))

		for i := 0; i < len(localAddrs); i++ {
			inbounds[i] = &core.InboundHandlerConfig{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange:      net.SinglePortRange(lport),
					Listen:         net.NewIPOrDomain(net.ParseAddress(localAddrs[i])),
					StreamSettings: &streamConfig,
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address:  net.NewIPOrDomain(proxyAddress),
					Networks: []net.Network{net.Network_TCP},
				}),
			}
		}

		config = &core.Config{
			Inbound: inbounds,
			Outbound: []*core.OutboundHandlerConfig{{
				ProxySettings: outboundProxy,
			}},
			App: apps,
		}
	} else {
		senderConfig := proxyman.SenderConfig{StreamSettings: &streamConfig}
		if connectionReuse {
			senderConfig.MultiplexSettings = &proxyman.MultiplexingConfig{Enabled: true, Concurrency: uint32(*mux)}
		}
		config = &core.Config{
			Inbound: []*core.InboundHandlerConfig{{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(lport),
					Listen:    net.NewIPOrDomain(net.ParseAddress(*localAddr)),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address:  net.NewIPOrDomain(net.LocalHostIP),
					Networks: []net.Network{net.Network_TCP},
				}),
			}},
			Outbound: []*core.OutboundHandlerConfig{{
				SenderSettings: serial.ToTypedMessage(&senderConfig),
				ProxySettings:  outboundProxy,
			}},
			App: apps,
		}
	}

	return config, nil
}

func LogConfig(logLevel string) *vlog.Config {
	config := &vlog.Config{
		Error: &vlog.LogSpecification{
			Type:  vlog.LogType_Console,
			Level: clog.Severity_Warning,
		},
		Access: &vlog.LogSpecification{
			Type: vlog.LogType_Console,
		},
	}
	level := strings.ToLower(logLevel)
	switch level {
	case "debug":
		config.Error.Level = clog.Severity_Debug
	case "info":
		config.Error.Level = clog.Severity_Info
	case "error":
		config.Error.Level = clog.Severity_Error
	case "none":
		config.Error.Type = vlog.LogType_None
		config.Access.Type = vlog.LogType_None
	}
	config.Access.Level = config.Error.Level
	return config
}

func ParseLocalAddr(localAddr string) []string {
	return strings.Split(localAddr, "|")
}

func (v *Plugin) Close() error {
	if v.instance == nil {
		return nil
	}
	return v.instance.Close()
}
