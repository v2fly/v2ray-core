package packetaddr

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport"
)

type blockingInterruptibleReader struct {
	started       chan struct{}
	interrupted   chan struct{}
	startOnce     sync.Once
	interruptOnce sync.Once
}

func newBlockingInterruptibleReader() *blockingInterruptibleReader {
	return &blockingInterruptibleReader{
		started:     make(chan struct{}),
		interrupted: make(chan struct{}),
	}
}

func (r *blockingInterruptibleReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	r.startOnce.Do(func() {
		close(r.started)
	})
	<-r.interrupted
	return nil, io.ErrClosedPipe
}

func (r *blockingInterruptibleReader) Interrupt() {
	r.interruptOnce.Do(func() {
		close(r.interrupted)
	})
}

type recordingInterruptibleWriter struct {
	interrupted   chan struct{}
	interruptOnce sync.Once
}

func newRecordingInterruptibleWriter() *recordingInterruptibleWriter {
	return &recordingInterruptibleWriter{interrupted: make(chan struct{})}
}

func (w *recordingInterruptibleWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	buf.ReleaseMulti(mb)
	return nil
}

func (w *recordingInterruptibleWriter) Interrupt() {
	w.interruptOnce.Do(func() {
		close(w.interrupted)
	})
}

func TestPacketConnectionAdaptorCloseInterruptsBlockedRead(t *testing.T) {
	reader := newBlockingInterruptibleReader()
	writer := newRecordingInterruptibleWriter()
	conn, err := ToPacketAddrConn(&transport.Link{
		Reader: reader,
		Writer: writer,
	}, net.Destination{
		Network: net.Network_TCP,
		Address: net.DomainAddress(streamPacketMagicAddress),
	})
	if err != nil {
		t.Fatalf("failed to create packet connection adaptor: %v", err)
	}

	readDone := make(chan error, 1)
	go func() {
		_, _, err := conn.ReadFrom(make([]byte, buf.Size))
		readDone <- err
	}()

	select {
	case <-reader.started:
	case <-time.After(time.Second):
		t.Fatal("ReadFrom did not block on the link reader")
	}

	closeDone := make(chan error, 1)
	go func() {
		closeDone <- conn.Close()
	}()

	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatalf("Close returned an error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close blocked while ReadFrom was waiting on the link reader")
	}

	select {
	case <-writer.interrupted:
	default:
		t.Fatal("Close did not interrupt the link writer")
	}

	select {
	case err := <-readDone:
		if err == nil {
			t.Fatal("ReadFrom returned without an error after interruption")
		}
	case <-time.After(time.Second):
		t.Fatal("ReadFrom did not return after Close interrupted the link reader")
	}
}

type singlePacketReader struct {
	payload []byte
	read    bool
}

func (r *singlePacketReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	if r.read {
		return nil, io.EOF
	}
	r.read = true
	b := buf.New()
	if _, err := b.Write(r.payload); err != nil {
		b.Release()
		return nil, err
	}
	return buf.MultiBuffer{b}, nil
}

func (r *singlePacketReader) Interrupt() {}

type discardingInterruptibleWriter struct{}

func (discardingInterruptibleWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	buf.ReleaseMulti(mb)
	return nil
}

func (discardingInterruptibleWriter) Interrupt() {}

func TestPacketConnectionAdaptorReadFromRejectsMalformedPacket(t *testing.T) {
	// 0xff is not a known packetaddr address type, so address extraction fails.
	conn, err := ToPacketAddrConn(&transport.Link{
		Reader: &singlePacketReader{payload: []byte{0xff, 0x01, 0x02}},
		Writer: discardingInterruptibleWriter{},
	}, net.Destination{
		Network: net.Network_UDP,
		Address: net.DomainAddress(seqPacketMagicAddress),
	})
	if err != nil {
		t.Fatalf("failed to create packet connection adaptor: %v", err)
	}

	n, addr, err := conn.ReadFrom(make([]byte, buf.Size))
	if err == nil {
		t.Fatal("ReadFrom accepted a malformed packetaddr packet")
	}
	if n != 0 {
		t.Fatalf("ReadFrom reported %v bytes for a malformed packetaddr packet", n)
	}
	if addr != nil {
		t.Fatalf("ReadFrom reported address %v for a malformed packetaddr packet", addr)
	}
}
