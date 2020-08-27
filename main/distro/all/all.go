package all

import (
	// The following are necessary as they register handlers in their init functions.

	// Required features. Can't remove unless there is replacements.
	_ "github.com/v2fly/v2ray-core/app/dispatcher"
	_ "github.com/v2fly/v2ray-core/app/proxyman/inbound"
	_ "github.com/v2fly/v2ray-core/app/proxyman/outbound"

	// Default commander and all its services. This is an optional feature.
	_ "github.com/v2fly/v2ray-core/app/commander"
	_ "github.com/v2fly/v2ray-core/app/log/command"
	_ "github.com/v2fly/v2ray-core/app/proxyman/command"
	_ "github.com/v2fly/v2ray-core/app/stats/command"

	// Other optional features.
	_ "github.com/v2fly/v2ray-core/app/dns"
	_ "github.com/v2fly/v2ray-core/app/log"
	_ "github.com/v2fly/v2ray-core/app/policy"
	_ "github.com/v2fly/v2ray-core/app/reverse"
	_ "github.com/v2fly/v2ray-core/app/router"
	_ "github.com/v2fly/v2ray-core/app/stats"

	// Inbound and outbound proxies.
	_ "github.com/v2fly/v2ray-core/proxy/blackhole"
	_ "github.com/v2fly/v2ray-core/proxy/dns"
	_ "github.com/v2fly/v2ray-core/proxy/dokodemo"
	_ "github.com/v2fly/v2ray-core/proxy/freedom"
	_ "github.com/v2fly/v2ray-core/proxy/http"
	_ "github.com/v2fly/v2ray-core/proxy/mtproto"
	_ "github.com/v2fly/v2ray-core/proxy/shadowsocks"
	_ "github.com/v2fly/v2ray-core/proxy/socks"
	_ "github.com/v2fly/v2ray-core/proxy/vless/inbound"
	_ "github.com/v2fly/v2ray-core/proxy/vless/outbound"
	_ "github.com/v2fly/v2ray-core/proxy/vmess/inbound"
	_ "github.com/v2fly/v2ray-core/proxy/vmess/outbound"

	// Transports
	_ "github.com/v2fly/v2ray-core/transport/internet/domainsocket"
	_ "github.com/v2fly/v2ray-core/transport/internet/http"
	_ "github.com/v2fly/v2ray-core/transport/internet/kcp"
	_ "github.com/v2fly/v2ray-core/transport/internet/quic"
	_ "github.com/v2fly/v2ray-core/transport/internet/tcp"
	_ "github.com/v2fly/v2ray-core/transport/internet/tls"
	_ "github.com/v2fly/v2ray-core/transport/internet/udp"
	_ "github.com/v2fly/v2ray-core/transport/internet/websocket"

	// Transport headers
	_ "github.com/v2fly/v2ray-core/transport/internet/headers/http"
	_ "github.com/v2fly/v2ray-core/transport/internet/headers/noop"
	_ "github.com/v2fly/v2ray-core/transport/internet/headers/srtp"
	_ "github.com/v2fly/v2ray-core/transport/internet/headers/tls"
	_ "github.com/v2fly/v2ray-core/transport/internet/headers/utp"
	_ "github.com/v2fly/v2ray-core/transport/internet/headers/wechat"
	_ "github.com/v2fly/v2ray-core/transport/internet/headers/wireguard"

	// JSON config support. Choose only one from the two below.
	// The following line loads JSON from v2ctl
	_ "github.com/v2fly/v2ray-core/main/json"
	// The following line loads JSON internally
	// _ "github.com/v2fly/v2ray-core/main/jsonem"

	// Load config from file or http(s)
	_ "github.com/v2fly/v2ray-core/main/confloader/external"
)
