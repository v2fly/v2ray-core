// +build android

package conf

import (
	"context"
	"net"
	"testing"
)

func TestBootstrapDNS(t *testing.T) {
	if ips, err := net.LookupIP("www.google.com"); len(ips) == 0 {
		t.Errorf("failed to lookupIP with BootstrapDNS, %v", err)
	}
}
