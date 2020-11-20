// +build !confonly

package domainsocket

import (
	"testing"
)

func TestUnixSockaddrCap(t *testing.T) {
	addrCap := UnixSockaddrCap()
	if addrCap == 0 {
		t.Error("error addrCap")
	}
}
