//go:build !confonly
// +build !confonly

package rrpitTransport

import "time"

const defaultReconnectRetryInterval = time.Second

type connectionPersistencePolicy struct {
	DisconnectedSessionRetention       time.Duration
	ReconnectRetryInterval             time.Duration
	KeepTransportSessionWithoutStreams bool
	IdleTimeout                        time.Duration
	RemoteControlInactivityTimeout     time.Duration
}

func buildConnectionPersistencePolicy(config *Config) connectionPersistencePolicy {
	var persistence *ConnectionPersistenceSetting
	if config != nil {
		persistence = config.GetPersistence()
	}
	policy := connectionPersistencePolicy{
		DisconnectedSessionRetention:       time.Duration(persistence.GetDisconnectedSessionRetention()),
		ReconnectRetryInterval:             time.Duration(persistence.GetReconnectRetryInterval()),
		KeepTransportSessionWithoutStreams: persistence.GetKeepTransportSessionWithoutStreams(),
		IdleTimeout:                        time.Duration(persistence.GetIdleTimeout()),
		RemoteControlInactivityTimeout:     time.Duration(persistence.GetRemoteControlInactivityTimeout()),
	}
	if policy.DisconnectedSessionRetention < 0 {
		policy.DisconnectedSessionRetention = 0
	}
	if policy.ReconnectRetryInterval < 0 {
		policy.ReconnectRetryInterval = 0
	}
	if policy.DisconnectedSessionRetention > 0 && policy.ReconnectRetryInterval == 0 {
		policy.ReconnectRetryInterval = defaultReconnectRetryInterval
	}
	if policy.IdleTimeout <= 0 {
		policy.IdleTimeout = rrpitClientSessionIdleTimeout
	}
	if policy.RemoteControlInactivityTimeout < 0 {
		policy.RemoteControlInactivityTimeout = 0
	}
	return policy
}
