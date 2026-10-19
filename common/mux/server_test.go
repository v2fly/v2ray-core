package mux_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/mux"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/pipe"
)

func TestServerWorkerCloseInterruptsLink(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	incoming, incomingWriter := pipe.New(pipe.WithoutSizeLimit())
	outgoing, outgoingWriter := pipe.New(pipe.WithoutSizeLimit())
	// No frames are dispatched, so this worker needs no routing dispatcher.
	worker, err := mux.NewServerWorker(ctx, nil, &transport.Link{Reader: incoming, Writer: outgoingWriter})
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		worker.Close()
		cancel()
		incoming.Interrupt()
		outgoing.Interrupt()
	})

	if err := worker.Close(); err != nil {
		t.Fatal(err)
	}
	if !worker.Closed() {
		t.Fatal("server worker still reports active sessions")
	}
	if _, err := outgoing.ReadMultiBufferTimeout(3 * time.Second); err != io.ErrClosedPipe {
		t.Fatalf("outgoing link was not interrupted: %v", err)
	}
	if err := incomingWriter.WriteMultiBuffer(buf.MergeBytes(nil, []byte("x"))); err != io.ErrClosedPipe {
		t.Fatalf("incoming link was not interrupted: %v", err)
	}
}
