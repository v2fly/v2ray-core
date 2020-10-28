package all

import (
	// The following are necessary as they register handlers in their init functions.

	// Required features. Can't remove unless there is replacements.
	_ "github.com/v2fly/v2ray-core/v5/app/dispatcher"
	_ "github.com/v2fly/v2ray-core/v5/app/proxyman/inbound"
	_ "github.com/v2fly/v2ray-core/v5/app/proxyman/outbound"

	// Default commander and all its services. This is an optional feature.
	_ "github.com/v2fly/v2ray-core/v5/app/commander"
	_ "github.com/v2fly/v2ray-core/v5/app/log/command"
	_ "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	_ "github.com/v2fly/v2ray-core/v5/app/stats/command"

	// Other optional features.
	_ "github.com/v2fly/v2ray-core/v5/app/dns"
	_ "github.com/v2fly/v2ray-core/v5/app/log"
	_ "github.com/v2fly/v2ray-core/v5/app/policy"
	_ "github.com/v2fly/v2ray-core/v5/app/reverse"
	_ "github.com/v2fly/v2ray-core/v5/app/router"
	_ "github.com/v2fly/v2ray-core/v5/app/stats"

	// Inbound and outbound proxies.
	_ "github.com/v2fly/v2ray-core/v5/proxy/blackhole"
	_ "github.com/v2fly/v2ray-core/v5/proxy/dns"
	_ "github.com/v2fly/v2ray-core/v5/proxy/dokodemo"
	_ "github.com/v2fly/v2ray-core/v5/proxy/freedom"
	_ "github.com/v2fly/v2ray-core/v5/proxy/http"
	_ "github.com/v2fly/v2ray-core/v5/proxy/mtproto"
	_ "github.com/v2fly/v2ray-core/v5/proxy/shadowsocks"
	_ "github.com/v2fly/v2ray-core/v5/proxy/socks"
	_ "github.com/v2fly/v2ray-core/v5/proxy/trojan"
	_ "github.com/v2fly/v2ray-core/v5/proxy/vless/inbound"
	_ "github.com/v2fly/v2ray-core/v5/proxy/vless/outbound"
	_ "github.com/v2fly/v2ray-core/v5/proxy/vmess/inbound"
	_ "github.com/v2fly/v2ray-core/v5/proxy/vmess/outbound"

	// Transports
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/domainsocket"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/http"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/kcp"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/quic"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/tcp"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/tls"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/udp"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/websocket"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/xtls"

	// Transport headers
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/headers/http"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/headers/noop"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/headers/srtp"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/headers/tls"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/headers/utp"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/headers/wechat"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/headers/wireguard"

	// JSON config support. Choose only one from the two below.
	// The following line loads JSON from v2ctl
	// _ "github.com/v2fly/v2ray-core/v5/main/json"
	// The following line loads JSON internally
	_ "github.com/v2fly/v2ray-core/v5/main/jsonem"

	// Load config from file or http(s)
	_ "github.com/v2fly/v2ray-core/v5/main/confloader/external"
)
