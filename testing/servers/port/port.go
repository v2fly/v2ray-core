// Package port hands out port numbers for tests.
//
// Picking a port by binding to port 0 and immediately releasing the socket is
// racy: the number returned that way belongs to the ephemeral port range, so
// the operating system may hand it to an outgoing connection of the test
// process, or to a later pick, before the server under test manages to bind
// it. Allocating from a window below the ephemeral range and remembering every
// number already handed out removes both races.
package port

import (
	"math/rand"
	"sync"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

const (
	// rangeStart and rangeEnd delimit the window ports are allocated from. It
	// sits above the fixed ports used by the JSON configurations in
	// testing/scenarios/config (17773-18104) and below the ephemeral port
	// range of every supported platform (32768 on Linux, 49152 on macOS and
	// Windows), so an allocated port cannot be claimed by an outgoing
	// connection before its server binds it.
	rangeStart = 20000
	rangeEnd   = 32000
)

var (
	access sync.Mutex
	// reserved holds every port already handed out by this process, so that no
	// port is ever returned twice even after the server using it went away.
	reserved = make(map[uint32]struct{})
	// cursor is where the next scan starts. It is randomized because
	// `go test ./...` runs the test binaries of several packages in parallel,
	// and each of them allocates from the same window.
	cursor = uint32(rangeStart + rand.Intn(rangeEnd-rangeStart)) // nolint: gosec
)

// Pick returns a port that is free and that no other Pick or PickRange call of
// this process has returned before.
func Pick() net.Port {
	return PickRange(1)
}

// PickRange returns the first port of a block of count consecutive free ports.
// Every port of the block is reserved, so tests that let a server listen on a
// port range can rely on the whole block staying available.
func PickRange(count uint32) net.Port {
	if count == 0 {
		panic("testing/servers/port: count must be positive")
	}

	access.Lock()
	defer access.Unlock()

	// Every start position of the window is tried once before giving up. Each
	// attempt moves the cursor by exactly one, because the additional advance
	// past an accepted block is followed by an immediate return.
	for attempt := uint32(0); attempt < rangeEnd-rangeStart; attempt++ {
		first := cursor
		advance(1)
		if first+count > rangeEnd {
			continue
		}
		if !available(first, count) {
			continue
		}
		for port := first; port < first+count; port++ {
			reserved[port] = struct{}{}
		}
		advance(count - 1)
		return net.Port(first)
	}

	panic("testing/servers/port: no free port block left")
}

// advance moves the scan cursor forward, wrapping around at the end of the
// window. It must be called with access held.
func advance(count uint32) {
	cursor += count
	if cursor >= rangeEnd {
		cursor = rangeStart + (cursor-rangeEnd)%(rangeEnd-rangeStart)
	}
}

// available reports whether count consecutive ports starting at first are
// unreserved and can be bound for both TCP and UDP. It must be called with
// access held.
func available(first, count uint32) bool {
	for port := first; port < first+count; port++ {
		if _, taken := reserved[port]; taken {
			return false
		}
		if !bindable(port) {
			return false
		}
	}
	return true
}

// bindable reports whether the port is currently free for both TCP and UDP on
// the loopback interface. Both protocols are checked because a port handed out
// here may be used for either of them.
func bindable(port uint32) bool {
	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.LocalHostIP.IP(), Port: int(port)})
	if err != nil {
		return false
	}
	if err := listener.Close(); err != nil {
		return false
	}

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.LocalHostIP.IP(), Port: int(port)})
	if err != nil {
		return false
	}
	return conn.Close() == nil
}
