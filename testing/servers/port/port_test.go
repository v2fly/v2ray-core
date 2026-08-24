package port_test

import (
	"testing"

	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/testing/servers/port"
)

func TestPickIsUnique(t *testing.T) {
	seen := make(map[net.Port]bool)
	for i := 0; i < 512; i++ {
		p := port.Pick()
		if seen[p] {
			t.Fatal("port handed out twice: ", p)
		}
		seen[p] = true
	}
}

func TestPickRangeIsContiguousAndFree(t *testing.T) {
	const count = 16
	first := port.PickRange(count)

	listeners := make([]net.Listener, 0, count)
	defer func() {
		for _, listener := range listeners {
			listener.Close()
		}
	}()

	for p := first; p < first+count; p++ {
		listener, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.LocalHostIP.IP(), Port: int(p)})
		if err != nil {
			t.Fatal("port ", p, " of the reserved range is not free: ", err)
		}
		listeners = append(listeners, listener)
	}
}

func TestPickDoesNotOverlapPickRange(t *testing.T) {
	const count = 8
	first := port.PickRange(count)

	for i := 0; i < 64; i++ {
		p := port.Pick()
		if p >= first && p < first+count {
			t.Fatal("Pick returned ", p, " which belongs to the range reserved at ", first)
		}
	}
}
