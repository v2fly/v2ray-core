// +build !confonly

package dispatcher

import (
	"v2ray.com/core/common"
	"v2ray.com/core/common/buf"
	"v2ray.com/core/features/stats"
)

type SizeStatWriter struct {
	Counter stats.Counter
	Writer  buf.Writer
	Record *int64
}

func (w *SizeStatWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	Len := int64(mb.Len())
	*w.Record += Len
	w.Counter.Add(Len)
	return w.Writer.WriteMultiBuffer(mb)
}

func (w *SizeStatWriter) Close() error {
	return common.Close(w.Writer)
}

func (w *SizeStatWriter) Interrupt() {
	common.Interrupt(w.Writer)
}
