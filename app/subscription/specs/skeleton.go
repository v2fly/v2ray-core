package specs

import (
	"encoding/json"
)

type OutboundConfig struct {
	Protocol      string            `json:"protocol"`
	Settings      json.RawMessage   `json:"settings"`
	StreamSetting *StreamConfig     `json:"streamSettings"`
	Metadata      map[string]string `json:"metadata"`
}

type StreamConfig struct {
	Transport         string          `json:"transport"`
	TransportSettings json.RawMessage `json:"transportSettings"`
	Security          string          `json:"security"`
	SecuritySettings  json.RawMessage `json:"securitySettings"`
}
