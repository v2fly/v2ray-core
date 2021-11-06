package all

import (
	// The following are necessary as they register handlers in their init functions.

	// Mandatory features. Can't remove unless there are replacements.
	_ "github.com/v2fly/v2ray-core/v4/app/dispatcher"
	_ "github.com/v2fly/v2ray-core/v4/app/proxyman/inbound"
	_ "github.com/v2fly/v2ray-core/v4/app/proxyman/outbound"

	// Default commander and all its services. This is an optional feature.
	_ "github.com/v2fly/v2ray-core/v4/app/commander"
	_ "github.com/v2fly/v2ray-core/v4/app/log/command"
	_ "github.com/v2fly/v2ray-core/v4/app/proxyman/command"
	_ "github.com/v2fly/v2ray-core/v4/app/stats/command"

	// Developer preview services
	_ "github.com/v2fly/v2ray-core/v4/app/instman/command"
	_ "github.com/v2fly/v2ray-core/v4/app/observatory/command"

	// Other optional features.
	_ "github.com/v2fly/v2ray-core/v4/app/dns"
	_ "github.com/v2fly/v2ray-core/v4/app/dns/fakedns"
	_ "github.com/v2fly/v2ray-core/v4/app/log"
	_ "github.com/v2fly/v2ray-core/v4/app/policy"
	_ "github.com/v2fly/v2ray-core/v4/app/reverse"
	_ "github.com/v2fly/v2ray-core/v4/app/router"
	_ "github.com/v2fly/v2ray-core/v4/app/stats"

	// Fix dependency cycle caused by core import in internet package
	_ "github.com/v2fly/v2ray-core/v4/transport/internet/tagged/taggedimpl"

	// Developer preview features
	_ "github.com/v2fly/v2ray-core/v4/app/instman"
	_ "github.com/v2fly/v2ray-core/v4/app/observatory"
	_ "github.com/v2fly/v2ray-core/v4/app/restful-api"

	// Inbound and outbound proxies.
	_ "github.com/v2fly/v2ray-core/v4/proxy/blackhole"
	_ "github.com/v2fly/v2ray-core/v4/proxy/dns"
	_ "github.com/v2fly/v2ray-core/v4/proxy/dokodemo"
	_ "github.com/v2fly/v2ray-core/v4/proxy/freedom"
	_ "github.com/v2fly/v2ray-core/v4/proxy/http"
	_ "github.com/v2fly/v2ray-core/v4/proxy/shadowsocks"
	_ "github.com/v2fly/v2ray-core/v4/proxy/shadowsocks/plugin/external"
	_ "github.com/v2fly/v2ray-core/v4/proxy/shadowsocks/plugin/self"
	_ "github.com/v2fly/v2ray-core/v4/proxy/socks"
	_ "github.com/v2fly/v2ray-core/v4/proxy/trojan"
	_ "github.com/v2fly/v2ray-core/v4/proxy/vless/inbound"
	_ "github.com/v2fly/v2ray-core/v4/proxy/vless/outbound"
	_ "github.com/v2fly/v2ray-core/v4/proxy/vmess/inbound"
	_ "github.com/v2fly/v2ray-core/v4/proxy/vmess/outbound"

	// Transports
	_ "github.com/v2fly/v2ray-core/v4/transport/internet/domainsocket"
	_ "github.com/v2fly/v2ray-core/v4/transport/internet/grpc"
	_ "github.com/v2fly/v2ray-core/v4/transport/internet/http"
	_ "github.com/v2fly/v2ray-core/v4/transport/internet/kcp"
	_ "github.com/v2fly/v2ray-core/v4/transport/internet/quic"
	_ "github.com/v2fly/v2ray-core/v4/transport/internet/tcp"
	_ "github.com/v2fly/v2ray-core/v4/transport/internet/tls"
	_ "github.com/v2fly/v2ray-core/v4/transport/internet/udp"
	_ "github.com/v2fly/v2ray-core/v4/transport/internet/websocket"

	// Transport headers
	_ "github.com/v2fly/v2ray-core/v4/transport/internet/headers/http"
	_ "github.com/v2fly/v2ray-core/v4/transport/internet/headers/noop"
	_ "github.com/v2fly/v2ray-core/v4/transport/internet/headers/srtp"
	_ "github.com/v2fly/v2ray-core/v4/transport/internet/headers/tls"
	_ "github.com/v2fly/v2ray-core/v4/transport/internet/headers/utp"
	_ "github.com/v2fly/v2ray-core/v4/transport/internet/headers/wechat"
	_ "github.com/v2fly/v2ray-core/v4/transport/internet/headers/wireguard"

	// Geo loaders
	_ "github.com/v2fly/v2ray-core/v4/infra/conf/geodata/memconservative"
	_ "github.com/v2fly/v2ray-core/v4/infra/conf/geodata/standard"

	// JSON, TOML, YAML config support. (jsonv4) This disable selective compile
	_ "github.com/v2fly/v2ray-core/v4/main/formats"

	// commands
	_ "github.com/v2fly/v2ray-core/v4/main/commands/all"

	//engineering commands
	_ "github.com/v2fly/v2ray-core/v4/main/commands/all/engineering"

	// Commands that rely on jsonv4 format This disable selective compile
	_ "github.com/v2fly/v2ray-core/v4/main/commands/all/api/jsonv4"
	_ "github.com/v2fly/v2ray-core/v4/main/commands/all/jsonv4"

	// V5 version of json configure file parser
	_ "github.com/v2fly/v2ray-core/v4/infra/conf/v5cfg"

	_ "github.com/v2fly/v2ray-core/v4/proxy/http/simplified"
	_ "github.com/v2fly/v2ray-core/v4/proxy/shadowsocks/simplified"
	_ "github.com/v2fly/v2ray-core/v4/proxy/socks/simplified"
	_ "github.com/v2fly/v2ray-core/v4/proxy/trojan/simplified"
)
