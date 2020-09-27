package conf

import "v2ray.com/core/app/stats"

// StatsConfig is the JSON config for app/stats.Manager.
type StatsConfig struct {
	Routing ChannelConfig `json:"routing"`
}

// ChannelConfig is the JSON config for app/stats.Channel.
type ChannelConfig struct {
	Enabled         bool   `json:"enabled"`
	SubscriberLimit *int32 `json:"subscriberLimit"`
}

// Build converts JSON config to Protobuf config used by app module.
func (c *StatsConfig) Build() (*stats.Config, error) {
	routing, err := c.Routing.Build()
	if err != nil {
		return nil, err
	}
	return &stats.Config{
		Routing: routing,
	}, nil
}

// Build converts JSON config to Protobuf config used by app module.
func (c *ChannelConfig) Build() (*stats.ChannelConfig, error) {
	if !c.Enabled {
		return nil, nil
	}
	cfg := &stats.ChannelConfig{
		SubscriberLimit:  1,
		BufferSize:       16,
		BroadcastTimeout: 100,
	}
	if c.SubscriberLimit != nil {
		cfg.SubscriberLimit = *c.SubscriberLimit
	}
	return cfg, nil
}
