package gvisorstack

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/v2fly/v2ray-core/v5/common/packetswitch"
)

// fakeDevice implements packetswitch.NetworkLayerDevice for testing.
type fakeDevice struct {
	mu       sync.Mutex
	writer   packetswitch.NetworkLayerPacketWriter
	writes   [][]byte
	closed   bool
	onAttach func(packetswitch.NetworkLayerPacketWriter) error
}

func (f *fakeDevice) OnAttach(w packetswitch.NetworkLayerPacketWriter) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.writer = w
	if f.onAttach != nil {
		return f.onAttach(w)
	}
	return nil
}

func (f *fakeDevice) Write(packet []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, errors.New("closed")
	}
	// make a copy to avoid aliasing test buffer.
	cp := make([]byte, len(packet))
	copy(cp, packet)
	f.writes = append(f.writes, cp)
	return len(packet), nil
}

func (f *fakeDevice) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

func (f *fakeDevice) getWriter() packetswitch.NetworkLayerPacketWriter {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.writer
}

func (f *fakeDevice) lastWrite() []byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.writes) == 0 {
		return nil
	}
	return f.writes[len(f.writes)-1]
}

// fakeDispatcher implements stack.NetworkDispatcher, capturing delivered packets.
type fakeDispatcher struct {
	mu        sync.Mutex
	protocols []tcpip.NetworkProtocolNumber
	pkts      [][]byte
}

func (d *fakeDispatcher) DeliverNetworkPacket(protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) {
	// capture packet payload safely by copying from AsSlices
	slices := pkt.AsSlices()
	total := 0
	for _, s := range slices {
		total += len(s)
	}
	cp := make([]byte, total)
	off := 0
	for _, s := range slices {
		copy(cp[off:], s)
		off += len(s)
	}
	// record
	d.mu.Lock()
	d.protocols = append(d.protocols, protocol)
	d.pkts = append(d.pkts, cp)
	d.mu.Unlock()
	// release the packet
	pkt.DecRef()
}

func (d *fakeDispatcher) DeliverLinkPacket(protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) {
	// not used in these tests
	pkt.DecRef()
}

func (d *fakeDispatcher) last() (tcpip.NetworkProtocolNumber, []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.protocols) == 0 {
		return 0, nil
	}
	return d.protocols[len(d.protocols)-1], d.pkts[len(d.pkts)-1]
}

func TestAttachAndInboundIPv4IPv6(t *testing.T) {
	dev := &fakeDevice{}
	a := NewNetworkLayerDeviceToGvisorLinkEndpointAdaptor(context.Background(), 1500, dev)
	d := &fakeDispatcher{}
	// Initially not attached
	if a.IsAttached() {
		t.Fatal("expected not attached")
	}

	// Attach should call device.OnAttach and store writer
	a.Attach(d)
	w := dev.getWriter()
	if w == nil {
		t.Fatal("device did not receive writer on attach")
	}
	if !a.IsAttached() {
		t.Fatal("expected attached after successful OnAttach")
	}

	// Send an IPv4 packet (first byte 0x45 = version 4, IHL 5)
	ipv4pkt := []byte{0x45, 0x00, 0x00, 0x04}
	if _, err := w.Write(ipv4pkt); err != nil {
		t.Fatalf("writer.Write failed: %v", err)
	}
	// Check dispatcher received it
	proto, payload := d.last()
	if proto != ipv4.ProtocolNumber {
		t.Fatalf("expected ipv4 protocol, got %v", proto)
	}
	if !reflect.DeepEqual(payload, ipv4pkt) {
		t.Fatalf("unexpected payload: %v", payload)
	}

	// IPv6 packet (first byte 0x60)
	ipv6pkt := []byte{0x60, 0x00, 0x00, 0x00}
	if _, err := w.Write(ipv6pkt); err != nil {
		t.Fatalf("writer.Write failed: %v", err)
	}
	proto2, payload2 := d.last()
	if proto2 != ipv6.ProtocolNumber {
		t.Fatalf("expected ipv6 protocol, got %v", proto2)
	}
	if !reflect.DeepEqual(payload2, ipv6pkt) {
		t.Fatalf("unexpected payload: %v", payload2)
	}
}

func TestInboundNonIPIsDropped(t *testing.T) {
	dev := &fakeDevice{}
	a := NewNetworkLayerDeviceToGvisorLinkEndpointAdaptor(context.Background(), 1500, dev)
	d := &fakeDispatcher{}
	a.Attach(d)
	w := dev.getWriter()
	if w == nil {
		t.Fatal("device did not receive writer on attach")
	}

	// Non-IP packet: first nibble 0
	pkt := []byte{0x00, 0x01, 0x02}
	if _, err := w.Write(pkt); err != nil {
		t.Fatalf("writer.Write failed: %v", err)
	}
	// Dispatcher should not have any packets
	proto, payload := d.last()
	if payload != nil || proto != 0 {
		t.Fatalf("expected no delivery for non-ip packet, got proto=%v payload=%v", proto, payload)
	}
}

func makePacketBufferPayload(b []byte) *stack.PacketBuffer {
	buf := buffer.MakeWithData(b)
	pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
		Payload: buf,
	})
	return pkt
}

func TestOutboundWritePacketsOk(t *testing.T) {
	dev := &fakeDevice{}
	a := NewNetworkLayerDeviceToGvisorLinkEndpointAdaptor(context.Background(), 1500, dev)

	// Prepare a PacketBufferList with one packet
	list := stack.PacketBufferList{}
	payload := []byte{0x45, 0x01, 0x02, 0x03}
	pkt := makePacketBufferPayload(payload)
	list.PushBack(pkt)

	written, err := a.WritePackets(list)
	if err != nil {
		t.Fatalf("WritePackets returned error: %v", err)
	}
	if written != 1 {
		t.Fatalf("expected 1 written, got %d", written)
	}
	// Device should have received the payload
	lw := dev.lastWrite()
	if !reflect.DeepEqual(lw, payload) {
		t.Fatalf("device write mismatch: got %v, want %v", lw, payload)
	}
}

func TestWritePacketsMTUExceeded(t *testing.T) {
	dev := &fakeDevice{}
	// mtu set to 2
	a := NewNetworkLayerDeviceToGvisorLinkEndpointAdaptor(context.Background(), 2, dev)

	list := stack.PacketBufferList{}
	payload := []byte{0x45, 0x01, 0x02} // len 3 > mtu 2
	pkt := makePacketBufferPayload(payload)
	list.PushBack(pkt)

	written, err := a.WritePackets(list)
	if err == nil {
		t.Fatalf("expected error due to message too long, got nil")
	}
	if _, ok := err.(*tcpip.ErrMessageTooLong); !ok {
		t.Fatalf("expected ErrMessageTooLong, got %T", err)
	}
	if written != 0 {
		t.Fatalf("expected 0 written, got %d", written)
	}
}

func TestCloseAndOnCloseActionAndWait(t *testing.T) {
	dev := &fakeDevice{}
	a := NewNetworkLayerDeviceToGvisorLinkEndpointAdaptor(context.Background(), 1500, dev)
	called := make(chan struct{})
	a.SetOnCloseAction(func() { close(called) })

	// Wait should block until Close is called. Run Wait in goroutine.
	done := make(chan struct{})
	go func() {
		a.Wait()
		close(done)
	}()

	// Give goroutine a moment to start
	time.Sleep(5 * time.Millisecond)
	// Now close
	a.Close()

	// onClose action should be called
	select {
	case <-called:
		// ok
	case <-time.After(100 * time.Millisecond):
		t.Fatal("onClose was not called")
	}

	// Wait should return
	select {
	case <-done:
		// ok
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Wait did not return after Close")
	}
}

func TestAttachOnAttachFailLeavesNotAttached(t *testing.T) {
	dev := &fakeDevice{}
	// make OnAttach return error
	dev.onAttach = func(w packetswitch.NetworkLayerPacketWriter) error {
		return errors.New("attach fail")
	}
	a := NewNetworkLayerDeviceToGvisorLinkEndpointAdaptor(context.Background(), 1500, dev)
	d := &fakeDispatcher{}
	a.Attach(d)
	if a.IsAttached() {
		t.Fatal("expected not attached when OnAttach fails")
	}
}

func TestWritePacketsWhenNoDevice(t *testing.T) {
	// Create adaptor with nil device
	a := NewNetworkLayerDeviceToGvisorLinkEndpointAdaptor(context.Background(), 1500, nil)
	list := stack.PacketBufferList{}
	payload := []byte{0x45}
	pkt := makePacketBufferPayload(payload)
	list.PushBack(pkt)
	written, err := a.WritePackets(list)
	if err == nil {
		t.Fatalf("expected ErrClosedForSend, got nil")
	}
	if _, ok := err.(*tcpip.ErrClosedForSend); !ok {
		t.Fatalf("expected ErrClosedForSend, got %T", err)
	}
	if written != 0 {
		t.Fatalf("expected 0 written, got %d", written)
	}
}

func TestSetMTUAndCapsAndHeaders(t *testing.T) {
	// Create adaptor and test MTU setter
	a := NewNetworkLayerDeviceToGvisorLinkEndpointAdaptor(context.Background(), 1500, nil)
	if a.MTU() != 1500 {
		t.Fatalf("initial MTU mismatch: %d", a.MTU())
	}
	a.SetMTU(9000)
	if a.MTU() != 9000 {
		t.Fatalf("MTU not updated: %d", a.MTU())
	}

	// Caps and headers
	if a.MaxHeaderLength() != 0 {
		t.Fatalf("MaxHeaderLength expected 0, got %d", a.MaxHeaderLength())
	}
	if a.Capabilities() != stack.CapabilityNone {
		t.Fatalf("Capabilities expected CapabilityNone, got %v", a.Capabilities())
	}
	if a.LinkAddress() != "" {
		t.Fatalf("LinkAddress expected empty, got %v", a.LinkAddress())
	}
	// AddHeader/ParseHeader should not panic and parse returns true
	pkt := makePacketBufferPayload([]byte{0x45})
	// Should be no-op
	a.AddHeader(pkt)
	if !a.ParseHeader(pkt) {
		t.Fatalf("ParseHeader expected true")
	}
	pkt.DecRef()
}

// errDevice fails after a certain number of writes to simulate partial failures.
type errDevice struct {
	mu        sync.Mutex
	writer    packetswitch.NetworkLayerPacketWriter
	writes    [][]byte
	failAfter int
	calls     int
	closed    bool
}

func (e *errDevice) OnAttach(w packetswitch.NetworkLayerPacketWriter) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.writer = w
	return nil
}

func (e *errDevice) Write(packet []byte) (int, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return 0, errors.New("closed")
	}
	e.calls++
	if e.failAfter > 0 && e.calls > e.failAfter {
		return 0, errors.New("injected write error")
	}
	cp := make([]byte, len(packet))
	copy(cp, packet)
	e.writes = append(e.writes, cp)
	return len(packet), nil
}

func (e *errDevice) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.closed = true
	return nil
}

func TestWritePacketsPartialOnError(t *testing.T) {
	e := &errDevice{failAfter: 1}
	a := NewNetworkLayerDeviceToGvisorLinkEndpointAdaptor(context.Background(), 1500, e)

	list := stack.PacketBufferList{}
	p1 := makePacketBufferPayload([]byte{0x45, 0x01})
	p2 := makePacketBufferPayload([]byte{0x45, 0x02})
	list.PushBack(p1)
	list.PushBack(p2)

	written, err := a.WritePackets(list)
	if err == nil {
		t.Fatalf("expected error due to injected write error")
	}
	// We mapped device write errors to ErrNoBufferSpace
	if _, ok := err.(*tcpip.ErrNoBufferSpace); !ok {
		t.Fatalf("expected ErrNoBufferSpace, got %T", err)
	}
	if written != 1 {
		t.Fatalf("expected 1 written, got %d", written)
	}
}

func TestMultiplePacketsWrite(t *testing.T) {
	dev := &fakeDevice{}
	a := NewNetworkLayerDeviceToGvisorLinkEndpointAdaptor(context.Background(), 1500, dev)

	list := stack.PacketBufferList{}
	p1 := makePacketBufferPayload([]byte{0x45, 0x01})
	p2 := makePacketBufferPayload([]byte{0x45, 0x02})
	p3 := makePacketBufferPayload([]byte{0x45, 0x03})
	list.PushBack(p1)
	list.PushBack(p2)
	list.PushBack(p3)

	written, err := a.WritePackets(list)
	if err != nil {
		t.Fatalf("WritePackets returned error: %v", err)
	}
	if written != 3 {
		t.Fatalf("expected 3 written, got %d", written)
	}
	// last write should be p3
	lw := dev.lastWrite()
	if !reflect.DeepEqual(lw, []byte{0x45, 0x03}) {
		t.Fatalf("unexpected last write: %v", lw)
	}
}

func TestConcurrentWritePacketsAndClose(t *testing.T) {
	dev := &fakeDevice{}
	a := NewNetworkLayerDeviceToGvisorLinkEndpointAdaptor(context.Background(), 1500, dev)

	wg := sync.WaitGroup{}
	nWorkers := 5
	nPerWorker := 50
	wg.Add(nWorkers)

	for i := 0; i < nWorkers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < nPerWorker; j++ {
				list := stack.PacketBufferList{}
				payload := []byte{0x45, byte(id), byte(j)}
				pkt := makePacketBufferPayload(payload)
				list.PushBack(pkt)
				_, _ = a.WritePackets(list)
			}
		}(i)
	}

	// Close after a short delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		a.Close()
	}()

	wg.Wait()
	// Wait should return quickly after Close
	a.Wait()
}
