package vmess3p

import (
	"encoding/json"
	"fmt"
	"strings"

	"v2ray.com/core/infra/conf"
)

// TPLink represents a third party vmess link
type TPLink struct {
	Ver      string      `json:"v,omitempty"`
	Add      string      `json:"add,omitempty"`
	Aid      interface{} `json:"aid,omitempty"`
	Host     string      `json:"host,omitempty"`
	ID       string      `json:"id,omitempty"`
	Net      string      `json:"net,omitempty"`
	Path     string      `json:"path,omitempty"`
	Port     interface{} `json:"port,omitempty"`
	Ps       string      `json:"ps,omitempty"`
	TLS      string      `json:"tls,omitempty"`
	Type     string      `json:"type,omitempty"`
	OrigLink string      `json:"-,omitempty"`
}

// ToString implements Link.ToString()
func (v TPLink) ToString() string {
	return v.OrigLink
}

// Tag implements Link.Tag()
func (v *TPLink) Tag() string {
	return v.Ps
}

// Detail implements Link.Detail()
func (v TPLink) Detail() string {
	return fmt.Sprintf("Net: %s\nAddr: %s\nPort: %v\nUUID: %s\nType: %s\nTLS: %s\nPS: %s\n", v.Net, v.Add, v.Port, v.ID, v.Type, v.TLS, v.Ps)
}

// ToOutbound implements Link.ToOutbound()
func (v *TPLink) ToOutbound() *conf.OutboundDetourConfig {
	out := &conf.OutboundDetourConfig{}
	out.Protocol = "vmess"

	p := conf.TransportProtocol(v.Net)
	s := &conf.StreamConfig{
		Network:  &p,
		Security: v.TLS,
	}

	switch v.Net {
	case "tcp":
		s.TCPSettings = &conf.TCPConfig{}
		if v.Type == "" || v.Type == "none" {
			s.TCPSettings.HeaderConfig = json.RawMessage([]byte(`{ "type": "none" }`))
		} else {
			pathb, _ := json.Marshal(strings.Split(v.Path, ","))
			hostb, _ := json.Marshal(strings.Split(v.Host, ","))
			s.TCPSettings.HeaderConfig = json.RawMessage([]byte(fmt.Sprintf(`
			{
				"type": "http",
				"request": {
					"path": %s,
					"headers": {
						"Host": %s
					}
				}
			}
			`, string(pathb), string(hostb))))
		}
	case "kcp":
		s.KCPSettings = &conf.KCPConfig{}
		s.KCPSettings.HeaderConfig = json.RawMessage([]byte(fmt.Sprintf(`{ "type": "%s" }`, v.Type)))
	case "ws":
		s.WSSettings = &conf.WebSocketConfig{}
		s.WSSettings.Path = v.Path
		s.WSSettings.Headers = map[string]string{
			"Host": v.Host,
		}
	case "h2", "http":
		s.HTTPSettings = &conf.HTTPConfig{
			Path: v.Path,
		}
		if v.Host != "" {
			h := conf.StringList(strings.Split(v.Host, ","))
			s.HTTPSettings.Host = &h
		}
	}

	if v.TLS == "tls" {
		s.TLSSettings = &conf.TLSConfig{
			Insecure: true,
		}
		if v.Host != "" {
			s.TLSSettings.ServerName = v.Host
		}
	}

	out.StreamSetting = s
	oset := json.RawMessage([]byte(fmt.Sprintf(`{
  "vnext": [
    {
      "address": "%s",
      "port": %v,
      "users": [
        {
          "id": "%s",
          "alterId": %v,
          "security": "auto"
        }
      ]
    }
  ]
}`, v.Add, v.Port, v.ID, v.Aid)))
	out.Settings = &oset
	return out
}

// IsEqual tests if this vmess link is equal to another
func (v *TPLink) IsEqual(c *TPLink) bool {
	realNet := func(n string) string {
		if n == "" {
			return "tcp"
		}
		return n
	}
	if realNet(v.Net) != realNet(c.Net) {
		return false
	}
	if fmt.Sprintf("%v", c.Port) != fmt.Sprintf("%v", v.Port) {
		return false
	}

	return v.Add == c.Add && v.Aid == c.Aid && v.Host == c.Host && v.ID == c.ID &&
		v.Path == c.Path && v.TLS == c.TLS && v.Type == c.Type
}
