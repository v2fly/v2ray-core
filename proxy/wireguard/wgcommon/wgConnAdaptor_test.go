package wgcommon

import (
	"net"
	"testing"
	"time"

	"golang.zx2c4.com/wireguard/conn"
)

func TestNetPacketConnToWg_OpenReceive_Send(t *testing.T) {
	// setup a UDP listener (server)
	svAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	svConn, err := net.ListenUDP("udp", svAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = svConn.Close() }()

	// client
	clConn, err := net.DialUDP("udp", nil, svConn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = clConn.Close() }()

	// Wrap server conn as Bind
	bind := NewNetPacketConnToWg(svConn)
	fns, port, err := bind.Open(0)
	if err != nil {
		t.Fatal(err)
	}
	if port == 0 {
		// LocalAddr should have set actualPort, otherwise use svConn
		if la := svConn.LocalAddr(); la != nil {
			if ua, ok := la.(*net.UDPAddr); ok && ua.Port != 0 {
				port = uint16(ua.Port)
				_ = port
			}
		}
	}
	if len(fns) == 0 {
		t.Fatal("no receive functions returned")
	}

	recvFn := fns[0]

	// run receiver in a goroutine
	recvBuf := make([]byte, 1500)
	sizes := make([]int, 1)
	eps := make([]conn.Endpoint, 1)
	ch := make(chan error, 1)
	go func() {
		_, err := recvFn([][]byte{recvBuf}, sizes, eps)
		ch <- err
	}()

	// send a message from client to server
	msg := []byte("hello-wg")
	if _, err := clConn.Write(msg); err != nil {
		t.Fatal(err)
	}

	// wait for receive
	select {
	case err := <-ch:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for receive")
	}

	// verify sizes and endpoint
	if sizes[0] != len(msg) {
		t.Fatalf("unexpected size: got %d want %d", sizes[0], len(msg))
	}
	if eps[0] == nil {
		t.Fatal("nil endpoint returned")
	}
	// endpoint DstToString should be the client's address
	if eps[0].DstToString() != clConn.LocalAddr().String() {
		t.Fatalf("unexpected endpoint dst: %s", eps[0].DstToString())
	}

	// Test Send: use a separate client conn to receive
	rcv2, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rcv2.Close() }()

	// create a bind adapter from a separate unconnected "sender" socket
	senderConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = senderConn.Close() }()
	senderBind := NewNetPacketConnToWg(senderConn)
	// parse endpoint for rcv2
	ep, err := senderBind.ParseEndpoint(rcv2.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}

	// send from senderBind to rcv2 via Send
	p := [][]byte{[]byte("ping")}
	if err := senderBind.Send(p, ep); err != nil {
		t.Fatal(err)
	}

	// read on rcv2
	buf := make([]byte, 1500)
	if err := rcv2.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	n, _, err := rcv2.ReadFromUDP(buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(p[0]) {
		t.Fatalf("unexpected recv len: %d", n)
	}
	_ = buf
}

func TestParseEndpointAndBatchSizeAndNilConstructor(t *testing.T) {
	// Parse valid IPv4 endpoint
	ep, err := NewNetPacketConnToWg(nil).ParseEndpoint("127.0.0.1:12345")
	if err != nil {
		t.Fatalf("ParseEndpoint failed: %v", err)
	}
	if ep == nil {
		t.Fatal("expected endpoint, got nil")
	}
	if ep.DstToString() != "127.0.0.1:12345" {
		t.Fatalf("unexpected DstToString: %s", ep.DstToString())
	}

	// Parse IPv6 endpoint
	ipv6ep, err := NewNetPacketConnToWg(nil).ParseEndpoint("[::1]:54321")
	if err != nil {
		t.Fatalf("ParseEndpoint v6 failed: %v", err)
	}
	if ipv6ep == nil {
		t.Fatal("expected ipv6 endpoint, got nil")
	}

	// BatchSize and Close/SetMark for nil-constructed adapter
	bind := NewNetPacketConnToWg(nil)
	if bind.BatchSize() != 1 {
		t.Fatalf("unexpected batch size: %d", bind.BatchSize())
	}
	// Close and SetMark should be no-ops and not panic
	if err := bind.Close(); err != nil {
		t.Fatalf("Close on nil adapter returned error: %v", err)
	}
	if err := bind.SetMark(123); err != nil {
		t.Fatalf("SetMark on nil adapter returned error: %v", err)
	}
}

func TestSendWithConnectedSocketProducesError(t *testing.T) {
	// create a server to target
	target, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = target.Close() }()

	// create a connected client socket (DialUDP)
	cl, err := net.DialUDP("udp", nil, target.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cl.Close() }()

	bind := NewNetPacketConnToWg(cl)
	ep, err := bind.ParseEndpoint("127.0.0.1:1")
	if err != nil {
		t.Fatal(err)
	}
	// Attempting to Send with a connected socket uses WriteTo and is expected
	// to return an error about using WriteTo on a pre-connected connection.
	p := [][]byte{[]byte("x")}
	err = bind.Send(p, ep)
	if err == nil {
		t.Fatalf("expected error when calling Send on adapter wrapping connected socket, got nil")
	}
}

func TestReceiveMultiplePackets(t *testing.T) {
	// server
	sv, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sv.Close() }()

	// client
	c, err := net.DialUDP("udp", nil, sv.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }()

	bind := NewNetPacketConnToWg(sv)
	fns, _, err := bind.Open(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(fns) == 0 {
		t.Fatal("no receive functions")
	}
	recv := fns[0]

	// prepare two buffers
	b1 := make([]byte, 64)
	b2 := make([]byte, 64)
	sizes := make([]int, 2)
	eps := make([]conn.Endpoint, 2)
	ch := make(chan error, 1)
	go func() {
		_, err := recv([][]byte{b1, b2}, sizes, eps)
		ch <- err
	}()

	// send two packets quickly
	if _, err := c.Write([]byte("one")); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Write([]byte("two")); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-ch:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for batched receive")
	}

	if sizes[0] == 0 && sizes[1] == 0 {
		t.Fatalf("expected at least one packet size to be non-zero")
	}
}
