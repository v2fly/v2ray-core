// +build android

package internet

import (
	"context"
	"testing"
)

func TestDNSResolver(t *testing.T) {
	resolver := NewDNSResolver()
	if ips, err := resolver.LookupIP(context.Background(), "ip", "www.google.com"); err != nil {
		t.Errorf("failed to lookupIP, %v, %v", ips, err)
	}
}
