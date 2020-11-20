// +build !confonly

package domainsocket

import (
	"golang.org/x/sys/unix"
)

// UnixSockaddrCap represents maximum length limit of a path of unix domain socket.
func UnixSockaddrCap() int {
	addr := &unix.RawSockaddrUnix{}
	path := addr.Path
	return cap(path)
}
