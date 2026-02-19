package interconnect

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v5/common/packetswitch"
)

type testWriter struct {
	mu       sync.Mutex
	received [][]byte
	ch       chan []byte
}

func newTestWriter(bufSize int) *testWriter {
	w := &testWriter{ch: make(chan []byte, bufSize)}
	return w
}

func (w *testWriter) Write(p []byte) (int, error) {
	// copy payload
	cp := make([]byte, len(p))
	copy(cp, p)
	w.mu.Lock()
	w.received = append(w.received, cp)
	w.mu.Unlock()
	select {
	case w.ch <- cp:
	default:
	}
	return len(p), nil
}

func (w *testWriter) ReceivedAll() [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([][]byte, len(w.received))
	copy(out, w.received)
	return out
}

func TestCable_HappyPath(t *testing.T) {
	c, err := NewNetworkLayerCable(context.Background())
	if err != nil {
		t.Fatalf("failed to create cable: %v", err)
	}
	l := c.GetLSideDevice()
	r := c.GetRSideDevice()

	wL := newTestWriter(4)
	wR := newTestWriter(4)

	if err := l.OnAttach(wL); err != nil {
		t.Fatalf("attach left failed: %v", err)
	}
	if err := r.OnAttach(wR); err != nil {
		t.Fatalf("attach right failed: %v", err)
	}

	payloadL := []byte("from-left")
	n, err := l.Write(payloadL)
	if err != nil {
		t.Fatalf("write left failed: %v", err)
	}
	if n != len(payloadL) {
		t.Fatalf("write returned wrong length: %d", n)
	}

	select {
	case got := <-wR.ch:
		if string(got) != string(payloadL) {
			t.Fatalf("unexpected payload at right: %s", string(got))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timeout waiting for payload on right")
	}

	payloadR := []byte("from-right")
	n, err = r.Write(payloadR)
	if err != nil {
		t.Fatalf("write right failed: %v", err)
	}
	if n != len(payloadR) {
		t.Fatalf("write returned wrong length: %d", n)
	}
	select {
	case got := <-wL.ch:
		if string(got) != string(payloadR) {
			t.Fatalf("unexpected payload at left: %s", string(got))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timeout waiting for payload on left")
	}
}

func TestCable_NoPeer(t *testing.T) {
	c, _ := NewNetworkLayerCable(context.Background())
	l := c.GetLSideDevice()
	wL := newTestWriter(1)
	if err := l.OnAttach(wL); err != nil {
		t.Fatalf("attach left failed: %v", err)
	}
	if n, err := l.Write([]byte("x")); err == nil || n != 0 {
		t.Fatalf("expected write to fail when no peer attached got n=%d err=%v", n, err)
	}
}

func TestCable_DoubleAttachAndClose(t *testing.T) {
	c, _ := NewNetworkLayerCable(context.Background())
	l := c.GetLSideDevice()
	w1 := newTestWriter(1)
	w2 := newTestWriter(1)
	if err := l.OnAttach(w1); err != nil {
		t.Fatalf("attach left failed: %v", err)
	}
	if err := l.OnAttach(w2); err == nil {
		t.Fatalf("expected second attach to fail")
	}

	r := c.GetRSideDevice()
	wr := newTestWriter(2)
	if err := r.OnAttach(wr); err != nil {
		t.Fatalf("attach right failed: %v", err)
	}

	// close left and ensure right cannot write
	if err := l.Close(); err != nil {
		t.Fatalf("close left failed: %v", err)
	}
	if n, err := r.Write([]byte("hello")); err == nil || n != 0 {
		t.Fatalf("expected write from right to fail after left closed got n=%d err=%v", n, err)
	}
}

func TestCable_ConcurrentWrites(t *testing.T) {
	c, _ := NewNetworkLayerCable(context.Background())
	l := c.GetLSideDevice()
	r := c.GetRSideDevice()
	wL := newTestWriter(100)
	wR := newTestWriter(100)
	if err := l.OnAttach(wL); err != nil {
		t.Fatalf("attach left failed: %v", err)
	}
	if err := r.OnAttach(wR); err != nil {
		t.Fatalf("attach right failed: %v", err)
	}

	var wg sync.WaitGroup
	count := 200
	wg.Add(count * 2)
	for i := 0; i < count; i++ {
		payloadL := []byte(fmt.Sprintf("L-%d", i%10))
		payloadR := []byte(fmt.Sprintf("R-%d", i%10))
		go func(p []byte) {
			defer wg.Done()
			_, _ = l.Write(p)
		}(payloadL)
		go func(p []byte) {
			defer wg.Done()
			_, _ = r.Write(p)
		}(payloadR)
	}
	wg.Wait()

	// drain channels (best-effort)
	timed := time.After(500 * time.Millisecond)
	for {
		select {
		case <-wL.ch:
		case <-wR.ch:
		case <-timed:
			return
		}
	}
}

// ensure testWriter implements the interface
var _ packetswitch.NetworkLayerPacketWriter = (*testWriter)(nil)
