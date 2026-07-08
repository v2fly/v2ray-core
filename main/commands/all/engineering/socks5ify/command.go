//go:build linux && !confonly
// +build linux,!confonly

/*
This feature is machine generated.
*/

package socks5ify

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/v2fly/v2ray-core/v5/main/commands/all/engineering"
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
)

var (
	socksFlag      *string
	socksUserFlag  *string
	socksPassFlag  *string
	quietFlag      *bool
	keepUIDFlag    *bool
	tunNameFlag    *string
	mtuFlag        *int
	ipv4HostFlag   *string
	ipv4GuestFlag  *string
	ipv4PrefixFlag *int
	ipv6Flag       *bool
	ipv6HostFlag   *string
	ipv6GuestFlag  *string
	ipv6PrefixFlag *int
	dnsFlag        *string
	resolvConfFlag *string
	bindFilesFlag  bindFileFlags
)

var cmdSocks5ify = &base.Command{
	UsageLine: "{{.Exec}} engineering socks5ify",
	Short:     "run a shell whose traffic is sent through a SOCKS5 proxy",
	Long: `
Create an unprivileged user, mount, and network namespace, configure a TUN
interface inside it, run V2Ray TUN outside the namespace, and start a shell or
command whose TCP and UDP traffic is proxied through SOCKS5.

Arguments:

	-socks <host:port|socks5://[user[:pass]@]host:port>
		Upstream SOCKS5 server. Required.

	-socks-user, -socks-pass
		Optional SOCKS5 username and password. These override URL credentials.

	-quiet
		Suppress non-error messages from the embedded V2Ray instance.

	-keep-uid
		Run the command as the original non-root caller UID and primary GID
		after namespace setup is complete.

	-tun-name <name>
		TUN interface name inside the network namespace. Default socks5ify0.

	-mtu <bytes>
		TUN MTU. Default 1500.

	-ipv6
		Also configure IPv6 address and default route.

	-tun-ipv4-host <ip>
		V2Ray-side IPv4 address. Default 198.18.0.1.

	-tun-ipv4-guest <ip>
		Namespace-side IPv4 address. Default 198.18.0.2.

	-tun-ipv4-prefix, -tun-ipv4-mask <bits>
		IPv4 address prefix length. Default 30.

	-tun-ipv6-host <ip>
		V2Ray-side IPv6 address. Default fd00:736f:636b:35::1.

	-tun-ipv6-guest <ip>
		Namespace-side IPv6 address. Default fd00:736f:636b:35::2.

	-tun-ipv6-prefix, -tun-ipv6-mask <bits>
		IPv6 address prefix length. Default 126.

	-dns <ip[,ip...]>
		Opt-in generated /etc/resolv.conf override inside the mount namespace.

	-resolv-conf <path>
		Opt-in bind mount of an existing resolver file onto /etc/resolv.conf.

	-bind-file <source:target>
		Read-only bind mount for a single file. May be repeated.

Examples:

	{{.Exec}} engineering socks5ify -socks 127.0.0.1:1080
	{{.Exec}} engineering socks5ify -socks socks5://user:pass@127.0.0.1:1080 -- curl https://example.com
`,
	Flag: func() flag.FlagSet {
		fs := flag.NewFlagSet("", flag.ExitOnError)
		socksFlag = fs.String("socks", "", "")
		socksUserFlag = fs.String("socks-user", "", "")
		socksPassFlag = fs.String("socks-pass", "", "")
		quietFlag = fs.Bool("quiet", false, "")
		keepUIDFlag = fs.Bool("keep-uid", false, "")
		tunNameFlag = fs.String("tun-name", defaultTunName, "")
		mtuFlag = fs.Int("mtu", defaultMTU, "")
		ipv4HostFlag = fs.String("tun-ipv4-host", defaultTunIPv4Host, "")
		ipv4GuestFlag = fs.String("tun-ipv4-guest", defaultTunIPv4Guest, "")
		ipv4PrefixFlag = fs.Int("tun-ipv4-prefix", defaultTunIPv4Prefix, "")
		fs.IntVar(ipv4PrefixFlag, "tun-ipv4-mask", defaultTunIPv4Prefix, "")
		ipv6Flag = fs.Bool("ipv6", false, "")
		ipv6HostFlag = fs.String("tun-ipv6-host", defaultTunIPv6Host, "")
		ipv6GuestFlag = fs.String("tun-ipv6-guest", defaultTunIPv6Guest, "")
		ipv6PrefixFlag = fs.Int("tun-ipv6-prefix", defaultTunIPv6Prefix, "")
		fs.IntVar(ipv6PrefixFlag, "tun-ipv6-mask", defaultTunIPv6Prefix, "")
		dnsFlag = fs.String("dns", "", "")
		resolvConfFlag = fs.String("resolv-conf", "", "")
		fs.Var(&bindFilesFlag, "bind-file", "")
		return *fs
	}(),
	Run: executeSocks5ify,
}

func init() {
	engineering.AddCommand(cmdSocks5ify)
}

func executeSocks5ify(cmd *base.Command, args []string) {
	if os.Getenv(childConfigEnv) != "" {
		if err := runChildFromEnv(); err != nil {
			base.Fatalf("socks5ify child failed: %v", err)
		}
		return
	}

	if err := cmd.Flag.Parse(args); err != nil {
		base.Fatalf("failed to parse flags: %v", err)
	}

	opts, child, err := buildOptions(cmd.Flag.Args())
	if err != nil {
		base.Fatalf("%v", err)
	}

	if err := runParent(opts, child); err != nil {
		base.Fatalf("socks5ify failed: %v", err)
	}
}

func buildOptions(command []string) (parentOptions, childConfig, error) {
	if *socksFlag == "" {
		return parentOptions{}, childConfig{}, fmt.Errorf("-socks is required")
	}
	if *mtuFlag <= 0 {
		return parentOptions{}, childConfig{}, fmt.Errorf("-mtu must be positive, got %d", *mtuFlag)
	}
	if *tunNameFlag == "" {
		return parentOptions{}, childConfig{}, fmt.Errorf("-tun-name must not be empty")
	}
	if *dnsFlag != "" && *resolvConfFlag != "" {
		return parentOptions{}, childConfig{}, fmt.Errorf("-dns and -resolv-conf are mutually exclusive")
	}
	if !*ipv6Flag && ipv6AddressOptionsChanged() {
		return parentOptions{}, childConfig{}, fmt.Errorf("IPv6 TUN address options require -ipv6")
	}

	socksServer, err := parseSocksServer(*socksFlag, *socksUserFlag, *socksPassFlag)
	if err != nil {
		return parentOptions{}, childConfig{}, err
	}
	ipv4, err := parseTunProtocolConfig("tun-ipv4", *ipv4HostFlag, *ipv4GuestFlag, *ipv4PrefixFlag, false)
	if err != nil {
		return parentOptions{}, childConfig{}, err
	}
	ipv6, err := parseTunProtocolConfig("tun-ipv6", *ipv6HostFlag, *ipv6GuestFlag, *ipv6PrefixFlag, true)
	if err != nil {
		return parentOptions{}, childConfig{}, err
	}

	child := childConfig{
		TunName:    *tunNameFlag,
		MTU:        *mtuFlag,
		KeepUID:    *keepUIDFlag && os.Getuid() != 0,
		CallerUID:  os.Getuid(),
		CallerGID:  os.Getgid(),
		IPv4:       ipv4,
		IPv6:       *ipv6Flag,
		IPv6Config: ipv6,
		DNS:        splitCommaList(*dnsFlag),
		ResolvConf: *resolvConfFlag,
		BindFiles:  append([]bindFile(nil), bindFilesFlag...),
		Command:    append([]string(nil), command...),
	}
	return parentOptions{SOCKS: socksServer, Quiet: *quietFlag}, child, nil
}

func ipv6AddressOptionsChanged() bool {
	return *ipv6HostFlag != defaultTunIPv6Host ||
		*ipv6GuestFlag != defaultTunIPv6Guest ||
		*ipv6PrefixFlag != defaultTunIPv6Prefix
}

func splitCommaList(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func encodeChildConfig(cfg childConfig) (string, error) {
	raw, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}

func decodeChildConfig(raw string) (childConfig, error) {
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return childConfig{}, err
	}
	var cfg childConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return childConfig{}, err
	}
	return cfg, nil
}

func childSocketFD() (int, error) {
	raw := os.Getenv(childSockFDEnv)
	if raw == "" {
		return -1, fmt.Errorf("%s is not set", childSockFDEnv)
	}
	fd, err := strconv.Atoi(raw)
	if err != nil {
		return -1, err
	}
	if fd < 0 {
		return -1, fmt.Errorf("invalid child socket fd %d", fd)
	}
	return fd, nil
}
