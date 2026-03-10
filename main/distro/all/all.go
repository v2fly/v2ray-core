package all

import (
	// The following are necessary as they register handlers in their init functions.

	// Mandatory features. Can't remove unless there are replacements.
	_ "github.com/v2fly/v2ray-core/v5/app/dispatcher"
	_ "github.com/v2fly/v2ray-core/v5/app/proxyman/inbound"
	_ "github.com/v2fly/v2ray-core/v5/app/proxyman/outbound"

	// Default commander and all its services. This is an optional feature.
	_ "github.com/v2fly/v2ray-core/v5/app/commander"
	_ "github.com/v2fly/v2ray-core/v5/app/log/command"
	_ "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	_ "github.com/v2fly/v2ray-core/v5/app/stats/command"

	// Developer preview services
	_ "github.com/v2fly/v2ray-core/v5/app/instman/command"
	_ "github.com/v2fly/v2ray-core/v5/app/observatory/command"

	// Other optional features.
	_ "github.com/v2fly/v2ray-core/v5/app/dns"
	_ "github.com/v2fly/v2ray-core/v5/app/dns/fakedns"
	_ "github.com/v2fly/v2ray-core/v5/app/log"
	_ "github.com/v2fly/v2ray-core/v5/app/policy"
	_ "github.com/v2fly/v2ray-core/v5/app/reverse"
	_ "github.com/v2fly/v2ray-core/v5/app/router"
	_ "github.com/v2fly/v2ray-core/v5/app/stats"

	// Fix dependency cycle caused by core import in internet package
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/tagged/taggedimpl"

	// Developer preview features
	_ "github.com/v2fly/v2ray-core/v5/app/commander/webcommander"
	_ "github.com/v2fly/v2ray-core/v5/app/instman"
	_ "github.com/v2fly/v2ray-core/v5/app/observatory"
	_ "github.com/v2fly/v2ray-core/v5/app/persistentstorage/filesystemstorage"
	_ "github.com/v2fly/v2ray-core/v5/app/tun"

	// Inbound and outbound proxies.
	_ "github.com/v2fly/v2ray-core/v5/proxy/blackhole"
	_ "github.com/v2fly/v2ray-core/v5/proxy/dns"
	_ "github.com/v2fly/v2ray-core/v5/proxy/dokodemo"
	_ "github.com/v2fly/v2ray-core/v5/proxy/freedom"
	_ "github.com/v2fly/v2ray-core/v5/proxy/http"
	_ "github.com/v2fly/v2ray-core/v5/proxy/shadowsocks"
	_ "github.com/v2fly/v2ray-core/v5/proxy/socks"
	_ "github.com/v2fly/v2ray-core/v5/proxy/trojan"
	_ "github.com/v2fly/v2ray-core/v5/proxy/vless/inbound"
	_ "github.com/v2fly/v2ray-core/v5/proxy/vless/outbound"
	_ "github.com/v2fly/v2ray-core/v5/proxy/vmess/inbound"
	_ "github.com/v2fly/v2ray-core/v5/proxy/vmess/outbound"

	// Developer preview proxies
	_ "github.com/v2fly/v2ray-core/v5/proxy/vlite/inbound"
	_ "github.com/v2fly/v2ray-core/v5/proxy/vlite/outbound"

	_ "github.com/v2fly/v2ray-core/v5/proxy/hysteria2"
	_ "github.com/v2fly/v2ray-core/v5/proxy/shadowsocks2022"

	// Transports
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/domainsocket"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/grpc"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/http"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/kcp"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/quic"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/tcp"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/tls"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/tls/utls"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/udp"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/websocket"

	// Developer preview transports
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/request/assembly"

	_ "github.com/v2fly/v2ray-core/v5/transport/internet/request/assembler/simple"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/request/roundtripper/httprt"

	_ "github.com/v2fly/v2ray-core/v5/transport/internet/request/assembler/packetconn"

	_ "github.com/v2fly/v2ray-core/v5/transport/internet/request/stereotype/meek"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/request/stereotype/mekya"

	_ "github.com/v2fly/v2ray-core/v5/transport/internet/request/roundtripperreverserserver/clicommand"

	_ "github.com/v2fly/v2ray-core/v5/transport/internet/dtls"

	_ "github.com/v2fly/v2ray-core/v5/transport/internet/httpupgrade"

	_ "github.com/v2fly/v2ray-core/v5/transport/internet/hysteria2"

	_ "github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorenrollment/roundtripperenrollmentconfirmation"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/server"

	_ "github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/mirrorenrollment/clicommand"

	// Transport headers
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/headers/http"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/headers/noop"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/headers/srtp"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/headers/tls"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/headers/utp"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/headers/wechat"
	_ "github.com/v2fly/v2ray-core/v5/transport/internet/headers/wireguard"

	// Geo loaders
	_ "github.com/v2fly/v2ray-core/v5/infra/conf/geodata/memconservative"
	_ "github.com/v2fly/v2ray-core/v5/infra/conf/geodata/standard"

	// JSON, TOML, YAML config support. (jsonv4) This disable selective compile
	_ "github.com/v2fly/v2ray-core/v5/main/formats"

	// commands
	_ "github.com/v2fly/v2ray-core/v5/main/commands/all"

	// engineering commands
	_ "github.com/v2fly/v2ray-core/v5/main/commands/all/engineering"
	_ "github.com/v2fly/v2ray-core/v5/main/commands/all/engineering/generateRandomData"

	// Commands that rely on jsonv4 format This disable selective compile
	_ "github.com/v2fly/v2ray-core/v5/main/commands/all/api/jsonv4"
	_ "github.com/v2fly/v2ray-core/v5/main/commands/all/jsonv4"

	// V5 version of json configure file parser
	_ "github.com/v2fly/v2ray-core/v5/infra/conf/v5cfg"

	// Simplified config
	_ "github.com/v2fly/v2ray-core/v5/proxy/http/simplified"
	_ "github.com/v2fly/v2ray-core/v5/proxy/shadowsocks/simplified"
	_ "github.com/v2fly/v2ray-core/v5/proxy/socks/simplified"
	_ "github.com/v2fly/v2ray-core/v5/proxy/trojan/simplified"

	// Subscription Supports
	_ "github.com/v2fly/v2ray-core/v5/app/subscription/subscriptionmanager"

	_ "github.com/v2fly/v2ray-core/v5/app/subscription/subscriptionmanager/command"

	// Subscription Containers: general purpose
	_ "github.com/v2fly/v2ray-core/v5/app/subscription/containers/base64urlline"
	_ "github.com/v2fly/v2ray-core/v5/app/subscription/containers/dataurlsingle"
	_ "github.com/v2fly/v2ray-core/v5/app/subscription/containers/jsonfieldarray"
	_ "github.com/v2fly/v2ray-core/v5/app/subscription/containers/jsonfieldarray/jsonified"
	_ "github.com/v2fly/v2ray-core/v5/app/subscription/containers/urlline"

	// Subscription Fetchers
	_ "github.com/v2fly/v2ray-core/v5/app/subscription/documentfetcher/dataurlfetcher"
	_ "github.com/v2fly/v2ray-core/v5/app/subscription/documentfetcher/httpfetcher"

	// Subscription Entries Converters
	_ "github.com/v2fly/v2ray-core/v5/app/subscription/entries/nonnative"
	_ "github.com/v2fly/v2ray-core/v5/app/subscription/entries/outbound" // Natively Supported Outbound Format
)
