package all

import (
	// The following are necessary as they register handlers in their init functions.

	// Required features. Can't remove unless there is replacements.
	_ "v2ray.com/core/v4/app/dispatcher"
	_ "v2ray.com/core/v4/app/proxyman/inbound"
	_ "v2ray.com/core/v4/app/proxyman/outbound"

	// Default commander and all its services. This is an optional feature.
	_ "v2ray.com/core/v4/app/commander"
	_ "v2ray.com/core/v4/app/log/command"
	_ "v2ray.com/core/v4/app/proxyman/command"
	_ "v2ray.com/core/v4/app/stats/command"

	// Other optional features.
	_ "v2ray.com/core/v4/app/dns"
	_ "v2ray.com/core/v4/app/log"
	_ "v2ray.com/core/v4/app/policy"
	_ "v2ray.com/core/v4/app/reverse"
	_ "v2ray.com/core/v4/app/router"
	_ "v2ray.com/core/v4/app/stats"

	// Inbound and outbound proxies.
	_ "v2ray.com/core/v4/proxy/blackhole"
	_ "v2ray.com/core/v4/proxy/dns"
	_ "v2ray.com/core/v4/proxy/dokodemo"
	_ "v2ray.com/core/v4/proxy/freedom"
	_ "v2ray.com/core/v4/proxy/http"
	_ "v2ray.com/core/v4/proxy/mtproto"
	_ "v2ray.com/core/v4/proxy/shadowsocks"
	_ "v2ray.com/core/v4/proxy/socks"
	_ "v2ray.com/core/v4/proxy/vmess/inbound"
	_ "v2ray.com/core/v4/proxy/vmess/outbound"

	// Transports
	_ "v2ray.com/core/v4/transport/internet/domainsocket"
	_ "v2ray.com/core/v4/transport/internet/http"
	_ "v2ray.com/core/v4/transport/internet/kcp"
	_ "v2ray.com/core/v4/transport/internet/quic"
	_ "v2ray.com/core/v4/transport/internet/tcp"
	_ "v2ray.com/core/v4/transport/internet/tls"
	_ "v2ray.com/core/v4/transport/internet/udp"
	_ "v2ray.com/core/v4/transport/internet/websocket"

	// Transport headers
	_ "v2ray.com/core/v4/transport/internet/headers/http"
	_ "v2ray.com/core/v4/transport/internet/headers/noop"
	_ "v2ray.com/core/v4/transport/internet/headers/srtp"
	_ "v2ray.com/core/v4/transport/internet/headers/tls"
	_ "v2ray.com/core/v4/transport/internet/headers/utp"
	_ "v2ray.com/core/v4/transport/internet/headers/wechat"
	_ "v2ray.com/core/v4/transport/internet/headers/wireguard"

	// JSON config support. Choose only one from the two below.
	// The following line loads JSON from v2ctl
	_ "v2ray.com/core/v4/main/json"
	// The following line loads JSON internally
	// _ "v2ray.com/core/main/jsonem"

	// Load config from file or http(s)
	_ "v2ray.com/core/v4/main/confloader/external"
)
