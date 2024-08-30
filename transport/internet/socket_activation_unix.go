//go:build unix
// +build unix

package internet

import (
	"os"
	"strconv"
	"syscall"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

func activate_socket(address string) (net.Listener, error) {
	fd, err := strconv.Atoi(address[8:])
	if err != nil {
		return nil, err
	}
	// Ignore the fail of SetNonblock: it's merely an optimization so that Go can poll this fd.
	_ = syscall.SetNonblock(fd, true)
	return net.FileListener(os.NewFile(uintptr(fd), address))
}
