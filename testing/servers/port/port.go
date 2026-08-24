// Package port provides an allocator of free ports for testing purposes.
package port

import (
	"os"
	"strconv"
	"sync"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

const (
	rangeStart = 11000
	rangeEnd   = 60000
	// blockSize is the distance between the first ports tried by two test
	// processes running on the same machine.
	blockSize = 512
)

var (
	access   sync.Mutex
	used     = make(map[int]struct{})
	nextPort = firstPort()
)

// firstPort returns the port the current process starts to search from. It is
// derived from the process id, so that test binaries running in parallel (as
// `go test ./...` does) are unlikely to hand out the same port.
func firstPort() int {
	blocks := (rangeEnd - rangeStart) / blockSize
	return rangeStart + (os.Getpid()%blocks)*blockSize
}

// Pick returns a port of the given network ("tcp4" or "udp4") that is free on
// the system and that has never been returned before by the current process.
func Pick(network string) net.Port {
	access.Lock()
	defer access.Unlock()

	for i := 0; i < rangeEnd-rangeStart; i++ {
		port := nextPort
		nextPort++
		if nextPort >= rangeEnd {
			nextPort = rangeStart
		}
		if _, found := used[port]; found {
			continue
		}
		if !isAvailable(network, port) {
			continue
		}
		used[port] = struct{}{}
		return net.Port(port)
	}

	panic("failed to pick an available " + network + " port")
}

func isAvailable(network string, port int) bool {
	switch network {
	case "udp", "udp4", "udp6":
		conn, err := net.ListenUDP(network, &net.UDPAddr{
			IP:   net.LocalHostIP.IP(),
			Port: port,
		})
		if err != nil {
			return false
		}
		conn.Close()
		return true
	default:
		listener, err := net.Listen(network, net.LocalHostIP.IP().String()+":"+strconv.Itoa(port))
		if err != nil {
			return false
		}
		listener.Close()
		return true
	}
}
