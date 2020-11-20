// +build !confonly

package domainsocket

import (
	"golang.org/x/sys/unix"
)

func UnixSockaddrCap() int {
	addr := &unix.RawSockaddrUnix{}
	return cap(addr.Path)
}
