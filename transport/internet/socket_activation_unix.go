//go:build unix
// +build unix

package internet

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"syscall"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

func activateSocket(address string) (net.Listener, error) {
	fd, err := strconv.Atoi(path.Base(address))
	if err != nil {
		return nil, err
	}

	err = syscall.SetNonblock(fd, true)
	if err != nil {
		return nil, err
	}

	acceptConn, err := syscall.GetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_ACCEPTCONN)
	if err != nil {
		return nil, err
	}
	if acceptConn == 0 {
		return nil, fmt.Errorf("socket '%s' has not been marked to accept connections", address)
	}

	sockType, err := syscall.GetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_TYPE)
	if err != nil {
		return nil, err
	}
	if sockType != syscall.SOCK_STREAM {
		// XXX: currently only stream socks are supported
		return nil, fmt.Errorf("socket '%s' is not a stream socket", address)
	}

	file := os.NewFile(uintptr(fd), address)
	defer file.Close()

	return net.FileListener(file)
}
