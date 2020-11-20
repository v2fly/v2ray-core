// +build !confonly

package domainsocket

import (
	"testing"
)

const sizeofSunPath = 108

func TestUnixSockaddrCap(t *testing.T) {
	addrCap := UnixSockaddrCap()
	if addrCap == 0 {
		t.Error("error addrCap")
	}
}
