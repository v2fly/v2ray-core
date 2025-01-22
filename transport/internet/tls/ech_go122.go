//go:build !go1.23
// +build !go1.23

package tls

import (
	"crypto/tls"
)

func ApplyECH(c *Config, config *tls.Config) error { //nolint: staticcheck
	return newError("using ECH require go 1.23 or higher")
}
