//go:build linux && !confonly
// +build linux,!confonly

package socks5ify

const (
	childConfigEnv = "V2RAY_SOCKS5IFY_CHILD_CONFIG"
	childSockFDEnv = "V2RAY_SOCKS5IFY_CHILD_SOCK_FD"

	defaultTunName = "socks5ify0"
	defaultMTU     = 1500

	defaultTunIPv4Host   = "198.18.0.1"
	defaultTunIPv4Guest  = "198.18.0.2"
	defaultTunIPv4Prefix = 30

	defaultTunIPv6Host   = "fd00:736f:636b:35::1"
	defaultTunIPv6Guest  = "fd00:736f:636b:35::2"
	defaultTunIPv6Prefix = 126
)

type bindFile struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type tunProtocolConfig struct {
	Host   string `json:"host"`
	Guest  string `json:"guest"`
	Prefix int    `json:"prefix"`
}

type childConfig struct {
	TunName    string            `json:"tun_name"`
	MTU        int               `json:"mtu"`
	KeepUID    bool              `json:"keep_uid"`
	CallerUID  int               `json:"caller_uid"`
	CallerGID  int               `json:"caller_gid"`
	IPv4       tunProtocolConfig `json:"ipv4"`
	IPv6       bool              `json:"ipv6"`
	IPv6Config tunProtocolConfig `json:"ipv6_config"`
	DNS        []string          `json:"dns"`
	ResolvConf string            `json:"resolv_conf"`
	BindFiles  []bindFile        `json:"bind_files"`
	Command    []string          `json:"command"`
}

type parentOptions struct {
	SOCKS socksServer
	Quiet bool
}

type socksServer struct {
	Host     string
	Port     uint32
	Username string
	Password string
}
