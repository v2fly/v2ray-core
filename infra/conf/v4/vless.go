package v4

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/golang/protobuf/proto"

	"github.com/ghxhy/v2ray-core/v5/common/net"
	"github.com/ghxhy/v2ray-core/v5/common/protocol"
	"github.com/ghxhy/v2ray-core/v5/common/serial"
	"github.com/ghxhy/v2ray-core/v5/infra/conf/cfgcommon"
	"github.com/ghxhy/v2ray-core/v5/proxy/vless"
	"github.com/ghxhy/v2ray-core/v5/proxy/vless/inbound"
	"github.com/ghxhy/v2ray-core/v5/proxy/vless/outbound"
)

type VLessInboundFallback struct {
	Alpn string          `json:"alpn"`
	Path string          `json:"path"`
	Type string          `json:"type"`
	Dest json.RawMessage `json:"dest"`
	Xver uint64          `json:"xver"`
}

type VLessInboundConfig struct {
	Clients    []json.RawMessage       `json:"clients"`
	Decryption string                  `json:"decryption"`
	Fallback   json.RawMessage         `json:"fallback"`
	Fallbacks  []*VLessInboundFallback `json:"fallbacks"`
}

// Build implements Buildable
func (c *VLessInboundConfig) Build() (proto.Message, error) {
	config := new(inbound.Config)
	config.Clients = make([]*protocol.User, len(c.Clients))
	for idx, rawUser := range c.Clients {
		user := new(protocol.User)
		if err := json.Unmarshal(rawUser, user); err != nil {
			return nil, newError(`VLESS clients: invalid user`).Base(err)
		}
		account := new(vless.Account)
		if err := json.Unmarshal(rawUser, account); err != nil {
			return nil, newError(`VLESS clients: invalid user`).Base(err)
		}

		if account.Encryption != "" {
			return nil, newError(`VLESS clients: "encryption" should not in inbound settings`)
		}

		user.Account = serial.ToTypedMessage(account)
		config.Clients[idx] = user
	}

	if c.Decryption != "none" {
		return nil, newError(`VLESS settings: please add/set "decryption":"none" to every settings`)
	}
	config.Decryption = c.Decryption

	if c.Fallback != nil {
		return nil, newError(`VLESS settings: please use "fallbacks":[{}] instead of "fallback":{}`)
	}
	for _, fb := range c.Fallbacks {
		var i uint16
		var s string
		if err := json.Unmarshal(fb.Dest, &i); err == nil {
			s = strconv.Itoa(int(i))
		} else {
			_ = json.Unmarshal(fb.Dest, &s)
		}
		config.Fallbacks = append(config.Fallbacks, &inbound.Fallback{
			Alpn: fb.Alpn,
			Path: fb.Path,
			Type: fb.Type,
			Dest: s,
			Xver: fb.Xver,
		})
	}
	for _, fb := range config.Fallbacks {
		/*
			if fb.Alpn == "h2" && fb.Path != "" {
				return nil, newError(`VLESS fallbacks: "alpn":"h2" doesn't support "path"`)
			}
		*/
		if fb.Path != "" && fb.Path[0] != '/' {
			return nil, newError(`VLESS fallbacks: "path" must be empty or start with "/"`)
		}
		if fb.Type == "" && fb.Dest != "" {
			if fb.Dest == "serve-ws-none" { // nolint:gocritic
				fb.Type = "serve"
			} else if filepath.IsAbs(fb.Dest) || fb.Dest[0] == '@' {
				fb.Type = "unix"
				if strings.HasPrefix(fb.Dest, "@@") && (runtime.GOOS == "linux" || runtime.GOOS == "android") {
					fullAddr := make([]byte, len(syscall.RawSockaddrUnix{}.Path)) // may need padding to work with haproxy
					copy(fullAddr, fb.Dest[1:])
					fb.Dest = string(fullAddr)
				}
			} else {
				if _, err := strconv.Atoi(fb.Dest); err == nil {
					fb.Dest = "127.0.0.1:" + fb.Dest
				}
				if _, _, err := net.SplitHostPort(fb.Dest); err == nil {
					fb.Type = "tcp"
				}
			}
		}
		if fb.Type == "" {
			return nil, newError(`VLESS fallbacks: please fill in a valid value for every "dest"`)
		}
		if fb.Xver > 2 {
			return nil, newError(`VLESS fallbacks: invalid PROXY protocol version, "xver" only accepts 0, 1, 2`)
		}
	}

	return config, nil
}

type VLessOutboundVnext struct {
	Address *cfgcommon.Address `json:"address"`
	Port    uint16             `json:"port"`
	Users   []json.RawMessage  `json:"users"`
}

type VLessOutboundConfig struct {
	Vnext []*VLessOutboundVnext `json:"vnext"`
}

// Build implements Buildable
func (c *VLessOutboundConfig) Build() (proto.Message, error) {
	config := new(outbound.Config)

	if len(c.Vnext) == 0 {
		return nil, newError(`VLESS settings: "vnext" is empty`)
	}
	config.Vnext = make([]*protocol.ServerEndpoint, len(c.Vnext))
	for idx, rec := range c.Vnext {
		if rec.Address == nil {
			return nil, newError(`VLESS vnext: "address" is not set`)
		}
		if len(rec.Users) == 0 {
			return nil, newError(`VLESS vnext: "users" is empty`)
		}
		spec := &protocol.ServerEndpoint{
			Address: rec.Address.Build(),
			Port:    uint32(rec.Port),
			User:    make([]*protocol.User, len(rec.Users)),
		}
		for idx, rawUser := range rec.Users {
			user := new(protocol.User)
			if err := json.Unmarshal(rawUser, user); err != nil {
				return nil, newError(`VLESS users: invalid user`).Base(err)
			}
			account := new(vless.Account)
			if err := json.Unmarshal(rawUser, account); err != nil {
				return nil, newError(`VLESS users: invalid user`).Base(err)
			}

			if account.Encryption != "none" {
				return nil, newError(`VLESS users: please add/set "encryption":"none" for every user`)
			}

			user.Account = serial.ToTypedMessage(account)
			spec.User[idx] = user
		}
		config.Vnext[idx] = spec
	}

	return config, nil
}
