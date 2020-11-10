package speed

import (
	"context"
	"golang.org/x/time/rate"
	"time"
	"v2ray.com/core/common"
	"v2ray.com/core/common/buf"
)

type Writer struct {
	writer  buf.Writer
	limiter *rate.Limiter
}

func RateWriter(writer buf.Writer, limiter *rate.Limiter) buf.Writer {
	return &Writer{
		writer:  writer,
		limiter: limiter,
	}
}

func (w *Writer) WriteMultiBuffer(mb buf.MultiBuffer) error {
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(100 * time.Millisecond))
	err := w.limiter.WaitN(ctx, int(mb.Len()) / 4)
	if err != nil {
		_ = newError("waiting to get a new ticket").AtDebug()
	}
	common.Must(err)

	return w.writer.WriteMultiBuffer(mb)
}

func (w *Writer) Close() error {
	return common.Close(w.writer)
}