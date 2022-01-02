package muxcfg

import "github.com/v2fly/v2ray-core/v5/app/proxyman"

type MuxConfig struct {
	Enabled     bool  `json:"enabled"`
	Concurrency int16 `json:"concurrency"`
}

// Build creates MultiplexingConfig, Concurrency < 0 completely disables mux.
func (m *MuxConfig) Build() *proxyman.MultiplexingConfig {
	if m.Concurrency < 0 {
		return nil
	}

	var con uint32 = 8
	if m.Concurrency > 0 {
		con = uint32(m.Concurrency)
	}

	return &proxyman.MultiplexingConfig{
		Enabled:     m.Enabled,
		Concurrency: con,
	}
}
