package wgcommon

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"golang.zx2c4.com/wireguard/tun"

	"github.com/v2fly/v2ray-core/v5/common/packetswitch"
)

// fakeNetDevice implements packetswitch.NetworkLayerDevice and optionally exposes Events().
type fakeNetDevice struct {
	mu     sync.Mutex
	writer packetswitch.NetworkLayerPacketWriter
	writes [][]byte
	closed bool
	events chan tun.Event
}

func (f *fakeNetDevice) OnAttach(w packetswitch.NetworkLayerPacketWriter) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.writer = w
	return nil
}

func (f *fakeNetDevice) Write(packet []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, errClosed
	}
	cp := make([]byte, len(packet))
	copy(cp, packet)
	f.writes = append(f.writes, cp)
	return len(packet), nil
}

func (f *fakeNetDevice) Close() error {
	f.mu.Lock()
	f.closed = true
	f.mu.Unlock()
	return nil
}

func (f *fakeNetDevice) getWriter() packetswitch.NetworkLayerPacketWriter {
	f.mu.Lock()
	w := f.writer
	f.mu.Unlock()
	return w
}

func (f *fakeNetDevice) lastWrite() []byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.writes) == 0 {
		return nil
	}
	return f.writes[len(f.writes)-1]
}

// Provide Events() so adaptor can forward events when present.
func (f *fakeNetDevice) Events() <-chan tun.Event {
	return f.events
}

var errClosed = &fakeError{"closed"}

type fakeError struct{ s string }

func (e *fakeError) Error() string { return e.s }

func TestNewAdaptor_ReadWrite_Basic(t *testing.T) {
	fd := &fakeNetDevice{events: make(chan tun.Event, 4)}
	// batchSize 2, inboundChannelSize 4
	a, err := NewNetworkLayerDeviceToWireguardTunDeviceAdaptor(1500, fd, 2, 4)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}

	w := fd.getWriter()
	if w == nil {
		t.Fatal("expected writer to be attached to fake device")
	}

	p1 := []byte{0x01, 0x02, 0x03}
	if _, err := w.Write(p1); err != nil {
		t.Fatalf("writer.Write failed: %v", err)
	}

	bufs := make([][]byte, 2)
	bufs[0] = make([]byte, 64)
	bufs[1] = make([]byte, 64)
	sizes := make([]int, 2)

	ret, err := a.Read(bufs, sizes, 0)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if ret != 1 {
		t.Fatalf("expected 1 packet read, got %d", ret)
	}
	if sizes[0] != len(p1) {
		t.Fatalf("expected sizes[0]=%d, got %d", len(p1), sizes[0])
	}
	if !reflect.DeepEqual(bufs[0][:sizes[0]], p1) {
		t.Fatalf("payload mismatch: got %v want %v", bufs[0][:sizes[0]], p1)
	}

	// Now write two packets and read both
	p2 := []byte{0x0a, 0x0b}
	p3 := []byte{0x0c}
	if _, err := w.Write(p2); err != nil {
		t.Fatalf("writer.Write p2 failed: %v", err)
	}
	if _, err := w.Write(p3); err != nil {
		t.Fatalf("writer.Write p3 failed: %v", err)
	}

	bufs2 := make([][]byte, 2)
	bufs2[0] = make([]byte, 16)
	bufs2[1] = make([]byte, 16)
	sizes2 := make([]int, 2)

	ret2, err := a.Read(bufs2, sizes2, 0)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if ret2 != 2 {
		t.Fatalf("expected 2 packets read, got %d", ret2)
	}
	if sizes2[0] != len(p2) || sizes2[1] != len(p3) {
		t.Fatalf("unexpected sizes: %v", sizes2)
	}
}

func TestNewAdaptor_WriteToNetwork(t *testing.T) {
	fd := &fakeNetDevice{events: make(chan tun.Event, 4)}
	a, err := NewNetworkLayerDeviceToWireguardTunDeviceAdaptor(1500, fd, 1, 4)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}

	payload := []byte{0xaa, 0xbb, 0xcc}
	bufs := make([][]byte, 1)
	bufs[0] = payload

	written, err := a.Write(bufs, 0)
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if written != 1 {
		t.Fatalf("expected 1 written, got %d", written)
	}
	lw := fd.lastWrite()
	if !reflect.DeepEqual(lw, payload) {
		t.Fatalf("device write mismatch: got %v want %v", lw, payload)
	}
}

func TestNewAdaptor_InboundBufferFullDrops(t *testing.T) {
	fd := &fakeNetDevice{events: make(chan tun.Event, 4)}
	// inboundChannelSize 1 so second write should fail
	a, err := NewNetworkLayerDeviceToWireguardTunDeviceAdaptor(1500, fd, 2, 1)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}
	w := fd.getWriter()
	if w == nil {
		t.Fatal("expected writer to be attached to fake device")
	}

	p1 := []byte{0x01}
	p2 := []byte{0x02}
	if _, err := w.Write(p1); err != nil {
		t.Fatalf("first write failed: %v", err)
	}
	// second write should return error due to buffer full
	if _, err := w.Write(p2); err == nil {
		t.Fatalf("expected second write to fail due to full buffer")
	}

	bufs := make([][]byte, 1)
	bufs[0] = make([]byte, 8)
	sizes := make([]int, 1)
	ret, err := a.Read(bufs, sizes, 0)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if ret != 1 {
		t.Fatalf("expected 1 packet read, got %d", ret)
	}
}

func TestNewAdaptor_ReadWithNonZeroOffset(t *testing.T) {
	// This test reproduces the 100% CPU bug where offset was misinterpreted
	// as a buffer index rather than a byte offset within each buffer.
	fd := &fakeNetDevice{events: make(chan tun.Event, 4)}
	a, err := NewNetworkLayerDeviceToWireguardTunDeviceAdaptor(1500, fd, 1, 4)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}
	w := fd.getWriter()
	if w == nil {
		t.Fatal("expected writer to be attached")
	}

	p1 := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	if _, err := w.Write(p1); err != nil {
		t.Fatalf("writer.Write failed: %v", err)
	}

	// Use offset=16 to mimic WireGuard's MessageTransportHeaderSize
	const offset = 16
	bufs := make([][]byte, 1)
	bufs[0] = make([]byte, 64)
	sizes := make([]int, 1)

	ret, err := a.Read(bufs, sizes, offset)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if ret != 1 {
		t.Fatalf("expected 1 packet read, got %d", ret)
	}
	if sizes[0] != len(p1) {
		t.Fatalf("expected sizes[0]=%d, got %d", len(p1), sizes[0])
	}
	// Data should be at bufs[0][offset:offset+sizes[0]], not bufs[0][0:sizes[0]]
	got := bufs[0][offset : offset+sizes[0]]
	if !reflect.DeepEqual(got, p1) {
		t.Fatalf("payload mismatch: got %v want %v", got, p1)
	}
	// The header area before offset should be untouched (all zeros)
	for i := 0; i < offset; i++ {
		if bufs[0][i] != 0 {
			t.Fatalf("byte at position %d was modified: %x", i, bufs[0][i])
		}
	}
}

func TestNewAdaptor_EventForwardingAndClose(t *testing.T) {
	fd := &fakeNetDevice{events: make(chan tun.Event, 4)}
	a, err := NewNetworkLayerDeviceToWireguardTunDeviceAdaptor(1500, fd, 1, 4)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}

	// send event from underlying device
	fd.events <- tun.EventUp

	select {
	case ev := <-a.Events():
		if ev != tun.EventUp {
			t.Fatalf("expected EventUp, got %v", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for forwarded event")
	}

	// Close adaptor and ensure events channel is closed
	if err := a.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	// reading from closed events channel should return immediately with ok==false
	select {
	case _, ok := <-a.Events():
		if ok {
			t.Fatal("expected events channel to be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for events channel close")
	}
}
