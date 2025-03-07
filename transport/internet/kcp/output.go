package kcp

import (
	"io"
	"sync"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/retry"
)

type SegmentWriter interface {
	Write(seg Segment) error
	Release()
}

type SimpleSegmentWriter struct {
	sync.Mutex
	buffer *buf.Buffer
	writer io.Writer
}

func NewSegmentWriter(writer io.Writer) SegmentWriter {
	return &SimpleSegmentWriter{
		writer: writer,
		buffer: buf.New(),
	}
}

func (w *SimpleSegmentWriter) Write(seg Segment) error {
	w.Lock()
	defer w.Unlock()

	w.buffer.Clear()
	rawBytes := w.buffer.Extend(seg.ByteSize())
	seg.Serialize(rawBytes)
	_, err := w.writer.Write(w.buffer.Bytes())
	return err
}

func (w *SimpleSegmentWriter) Release() {
	w.buffer.Release()
}

type RetryableWriter struct {
	writer SegmentWriter
}

func NewRetryableWriter(writer SegmentWriter) SegmentWriter {
	return &RetryableWriter{
		writer: writer,
	}
}

func (w *RetryableWriter) Write(seg Segment) error {
	return retry.Timed(5, 100).On(func() error {
		return w.writer.Write(seg)
	})
}

func (w *RetryableWriter) Release() {
	w.writer.Release()
}
