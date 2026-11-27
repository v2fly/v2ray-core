package inbound

import (
	"context"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

type udpSession struct {
	conn    *udpConn
	release chan struct{}
}

// heldUDPInbound keeps every UDP session open until the test releases it.
type heldUDPInbound struct {
	sessions chan udpSession
}

func (*heldUDPInbound) Network() []net.Network {
	return []net.Network{net.Network_UDP}
}

func (p *heldUDPInbound) Process(_ context.Context, _ net.Network, conn internet.Connection, _ routing.Dispatcher) error {
	s := udpSession{conn: conn.(*udpConn), release: make(chan struct{})}
	p.sessions <- s
	<-s.release
	return nil
}

func TestUDPWorkerEndedSessionKeepsReplacementConn(t *testing.T) {
	inbound := &heldUDPInbound{sessions: make(chan udpSession, 2)}
	w := &udpWorker{
		proxy:      inbound,
		address:    net.LocalHostIP,
		port:       net.Port(1080),
		activeConn: make(map[connID]*udpConn),
		ctx:        context.Background(),
	}
	w.checker = &task.Periodic{Interval: time.Hour, Execute: w.clean}
	defer w.checker.Close()

	source := net.UDPDestination(net.LocalHostIP, net.Port(53000))
	id := connID{src: source}

	w.callback(buf.New(), source, net.Destination{})
	first := <-inbound.sessions
	// The first session ends, but its goroutine has not removed it yet.
	first.conn.Close()

	w.callback(buf.New(), source, net.Destination{})
	second := <-inbound.sessions
	defer close(second.release)
	if second.conn == first.conn {
		t.Fatal("a packet after the session ended reused the closed connection")
	}

	close(first.release)
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		w.RLock()
		current := w.activeConn[id]
		w.RUnlock()
		if current != second.conn {
			t.Fatalf("the ended session removed its replacement: activeConn[%v] = %p, want %p", source, current, second.conn)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
