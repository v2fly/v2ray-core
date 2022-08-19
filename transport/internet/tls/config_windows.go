//go:build windows
// +build windows

package tls

import "crypto/x509"

func (c *Config) getCertPool() (*x509.CertPool, error) {
	if c.DisableSystemRoot {
		return c.loadSelfCertPool(Certificate_AUTHORITY_VERIFY)
	}

	return nil, nil
}
