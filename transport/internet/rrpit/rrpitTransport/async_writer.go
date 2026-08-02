//go:build !confonly

package rrpitTransport

import (
	"io"
	"sync"
)

// asyncPacketWriter owns one FIFO worker for one physical packet channel. A
// transport session with multiple channels therefore encrypts and writes on
// multiple goroutines while preserving the materialized sequence order within
// each channel.
type asyncPacketWriter struct {
	writer io.WriteCloser
	queue  chan []byte
	done   chan struct{}
	exited chan struct{}

	terminalOnce sync.Once
	closeOnce    sync.Once
	errMu        sync.Mutex
	err          error
}

func newAsyncPacketWriter(writer io.WriteCloser, capacity int) io.WriteCloser {
	if writer == nil || capacity <= 0 {
		return writer
	}
	w := &asyncPacketWriter{
		writer: writer,
		queue:  make(chan []byte, capacity),
		done:   make(chan struct{}),
		exited: make(chan struct{}),
	}
	go w.run()
	return w
}

func (w *asyncPacketWriter) Write(payload []byte) (int, error) {
	owned := append([]byte(nil), payload...)
	return w.WriteOwned(owned)
}

// WriteOwned transfers ownership of payload to the writer on success. It is
// used by the materialized RRIPT channel, which has just allocated the packet
// and no longer needs it after Write returns.
func (w *asyncPacketWriter) WriteOwned(payload []byte) (int, error) {
	if w == nil || w.writer == nil {
		return 0, io.ErrClosedPipe
	}
	select {
	case <-w.done:
		return 0, w.terminalError()
	default:
	}
	select {
	case w.queue <- payload:
		return len(payload), nil
	case <-w.done:
		return 0, w.terminalError()
	}
}

func (w *asyncPacketWriter) run() {
	defer close(w.exited)
	for {
		select {
		case payload := <-w.queue:
			written, err := w.writer.Write(payload)
			if err == nil && written != len(payload) {
				err = io.ErrShortWrite
			}
			if err != nil {
				w.stop(err)
				_ = w.writer.Close()
				return
			}
		case <-w.done:
			return
		}
	}
}

func (w *asyncPacketWriter) stop(err error) {
	if err == nil {
		err = io.ErrClosedPipe
	}
	w.terminalOnce.Do(func() {
		w.errMu.Lock()
		w.err = err
		w.errMu.Unlock()
		close(w.done)
	})
}

func (w *asyncPacketWriter) terminalError() error {
	w.errMu.Lock()
	defer w.errMu.Unlock()
	if w.err == nil {
		return io.ErrClosedPipe
	}
	return w.err
}

func (w *asyncPacketWriter) Close() error {
	if w == nil || w.writer == nil {
		return nil
	}
	var closeErr error
	w.closeOnce.Do(func() {
		w.stop(io.ErrClosedPipe)
		closeErr = w.writer.Close()
		<-w.exited
	})
	return closeErr
}
