//go:build linux && !confonly
// +build linux,!confonly

package socks5ify

const (
	childConfigEnv = "V2RAY_SOCKS5IFY_CHILD_CONFIG"
	childSockFDEnv = "V2RAY_SOCKS5IFY_CHILD_SOCK_FD"

	defaultTunName = "socks5ify0"
	defaultMTU     = 1500

	tunIPv4Host   = "198.18.0.1"
	tunIPv4Guest  = "198.18.0.2"
	tunIPv4Prefix = 30

	tunIPv6Host   = "fd00:736f:636b:35::1"
	tunIPv6Guest  = "fd00:736f:636b:35::2"
	tunIPv6Prefix = 126
)

type bindFile struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type childConfig struct {
	TunName    string     `json:"tun_name"`
	MTU        int        `json:"mtu"`
	IPv6       bool       `json:"ipv6"`
	DNS        []string   `json:"dns"`
	ResolvConf string     `json:"resolv_conf"`
	BindFiles  []bindFile `json:"bind_files"`
	Command    []string   `json:"command"`
}

type parentOptions struct {
	SOCKS socksServer
}

type socksServer struct {
	Host     string
	Port     uint32
	Username string
	Password string
}
