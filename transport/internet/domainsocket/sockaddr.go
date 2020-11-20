// +build !confonly

package domainsocket

import (
	"golang.org/x/sys/unix"
)

func UnixSockaddrCap() int {
	addr := &unix.RawSockaddrUnix{}
	path := addr.Path
	return cap(path)
}
