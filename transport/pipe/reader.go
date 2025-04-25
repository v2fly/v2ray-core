package pipe

import (
	"time"

	"github.com/ghxhy/v2ray-core/v5/common/buf"
)

// Reader is a buf.Reader that reads content from a pipe.
type Reader struct {
	pipe *pipe
}

// ReadMultiBuffer implements buf.Reader.
func (r *Reader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	return r.pipe.ReadMultiBuffer()
}

// ReadMultiBufferTimeout reads content from a pipe within the given duration, or returns buf.ErrTimeout otherwise.
func (r *Reader) ReadMultiBufferTimeout(d time.Duration) (buf.MultiBuffer, error) {
	return r.pipe.ReadMultiBufferTimeout(d)
}

// Interrupt implements common.Interruptible.
func (r *Reader) Interrupt() {
	r.pipe.Interrupt()
}
