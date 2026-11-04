package reverse

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/mux"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/pipe"
)

// silentDropWriter reports successful writes while discarding data after the fault.
type silentDropWriter struct {
	writer buf.Writer
	drop   atomic.Bool
}

func (w *silentDropWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	if w.drop.Load() {
		buf.ReleaseMulti(mb)
		return nil
	}
	return w.writer.WriteMultiBuffer(mb)
}

// muxWorkerPair connects one Portal and one Bridge over in-memory Mux transport.
type muxWorkerPair struct {
	client         *mux.ClientWorker
	portal         *PortalWorker
	server         *mux.ServerWorker
	bridgeInput    *pipe.Reader
	portalToBridge *silentDropWriter
	bridgeToPortal *silentDropWriter
}

func newMuxWorkerPair(t *testing.T, ctx context.Context) *muxWorkerPair {
	t.Helper()
	portalReader, bridgeWriter := pipe.New(pipe.WithoutSizeLimit()) // Bridge → Portal
	bridgeReader, portalWriter := pipe.New(pipe.WithoutSizeLimit()) // Portal → Bridge
	portalToBridge := &silentDropWriter{writer: portalWriter}
	bridgeToPortal := &silentDropWriter{writer: bridgeWriter}
	client, err := mux.NewClientWorker(transport.Link{Reader: portalReader, Writer: portalToBridge}, mux.ClientStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	bridge := &BridgeWorker{}
	server, err := mux.NewServerWorker(ctx, bridge, &transport.Link{Reader: bridgeReader, Writer: bridgeToPortal})
	if err != nil {
		t.Fatal(err)
	}
	portal, err := NewPortalWorker(ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		portal.control.Close()
		portalReader.Interrupt()
		bridgeReader.Interrupt()
		bridgeWriter.Interrupt()
		portalWriter.Interrupt()
	})
	return &muxWorkerPair{client: client, portal: portal, server: server, bridgeInput: bridgeReader, portalToBridge: portalToBridge, bridgeToPortal: bridgeToPortal}
}

// TestPortalStopsSelectingHalfOpenWorker checks that a healthy replacement is
// selected when the old Bridge side has closed but the Portal side remains open.
func TestPortalStopsSelectingHalfOpenWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	old := newMuxWorkerPair(t, ctx)
	picker := &StaticMuxPicker{}
	picker.AddWorker(old.portal)

	// Establish the original control stream before breaking the tunnel.
	time.Sleep(time.Second)
	old.portalToBridge.drop.Store(true)
	old.bridgeToPortal.drop.Store(true)
	old.bridgeInput.Interrupt() // Bridge sees EOF; Portal's receive pipe stays open.
	deadline := time.Now().Add(3 * time.Second)
	for !old.server.Closed() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if !old.server.Closed() || old.client.Closed() {
		t.Fatalf("expected Bridge closed and Portal half-open: bridgeClosed=%t portalClosed=%t", old.server.Closed(), old.client.Closed())
	}

	// The old worker remains half-open while its replacement becomes available.
	fresh := newMuxWorkerPair(t, ctx)
	picker.AddWorker(fresh.portal)

	// Portal should eventually stop assigning requests to the old worker.
	time.Sleep(13 * time.Second)
	worker, err := picker.PickAvailable()
	if err != nil || worker != fresh.client {
		t.Fatalf("selected stale worker after Bridge disconnect: selected=%p, old=%p, replacement=%p, err=%v", worker, old.client, fresh.client, err)
	}
}
