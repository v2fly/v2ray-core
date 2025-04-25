package pipe

import (
	"errors"
	"io"
	"runtime"
	"sync"
	"time"

	"github.com/ghxhy/v2ray-core/v5/common"
	"github.com/ghxhy/v2ray-core/v5/common/buf"
	"github.com/ghxhy/v2ray-core/v5/common/signal"
	"github.com/ghxhy/v2ray-core/v5/common/signal/done"
)

type state byte

const (
	open state = iota
	closed
	errord
)

type pipeOption struct {
	limit           int32 // maximum buffer size in bytes
	discardOverflow bool
}

func (o *pipeOption) isFull(curSize int32) bool {
	return o.limit >= 0 && curSize > o.limit
}

type pipe struct {
	sync.Mutex
	data        buf.MultiBuffer
	readSignal  *signal.Notifier
	writeSignal *signal.Notifier
	done        *done.Instance
	option      pipeOption
	state       state
}

var (
	errBufferFull = errors.New("buffer full")
	errSlowDown   = errors.New("slow down")
)

func (p *pipe) getState(forRead bool) error {
	switch p.state {
	case open:
		if !forRead && p.option.isFull(p.data.Len()) {
			return errBufferFull
		}
		return nil
	case closed:
		if !forRead {
			return io.ErrClosedPipe
		}
		if !p.data.IsEmpty() {
			return nil
		}
		return io.EOF
	case errord:
		return io.ErrClosedPipe
	default:
		panic("impossible case")
	}
}

func (p *pipe) readMultiBufferInternal() (buf.MultiBuffer, error) {
	p.Lock()
	defer p.Unlock()

	if err := p.getState(true); err != nil {
		return nil, err
	}

	data := p.data
	p.data = nil
	return data, nil
}

func (p *pipe) ReadMultiBuffer() (buf.MultiBuffer, error) {
	for {
		data, err := p.readMultiBufferInternal()
		if data != nil || err != nil {
			p.writeSignal.Signal()
			return data, err
		}

		select {
		case <-p.readSignal.Wait():
		case <-p.done.Wait():
		}
	}
}

func (p *pipe) ReadMultiBufferTimeout(d time.Duration) (buf.MultiBuffer, error) {
	timer := time.NewTimer(d)
	defer timer.Stop()

	for {
		data, err := p.readMultiBufferInternal()
		if data != nil || err != nil {
			p.writeSignal.Signal()
			return data, err
		}

		select {
		case <-p.readSignal.Wait():
		case <-p.done.Wait():
		case <-timer.C:
			return nil, buf.ErrReadTimeout
		}
	}
}

func (p *pipe) writeMultiBufferInternal(mb buf.MultiBuffer) error {
	p.Lock()
	defer p.Unlock()

	if err := p.getState(false); err != nil {
		return err
	}

	if p.data == nil {
		p.data = mb
		return nil
	}

	p.data, _ = buf.MergeMulti(p.data, mb)
	return errSlowDown
}

func (p *pipe) WriteMultiBuffer(mb buf.MultiBuffer) error {
	if mb.IsEmpty() {
		return nil
	}

	for {
		err := p.writeMultiBufferInternal(mb)
		if err == nil {
			p.readSignal.Signal()
			return nil
		}

		if err == errSlowDown {
			p.readSignal.Signal()

			// Yield current goroutine. Hopefully the reading counterpart can pick up the payload.
			runtime.Gosched()
			return nil
		}

		if err == errBufferFull && p.option.discardOverflow {
			buf.ReleaseMulti(mb)
			return nil
		}

		if err != errBufferFull {
			buf.ReleaseMulti(mb)
			p.readSignal.Signal()
			return err
		}

		select {
		case <-p.writeSignal.Wait():
		case <-p.done.Wait():
			return io.ErrClosedPipe
		}
	}
}

func (p *pipe) Close() error {
	p.Lock()
	defer p.Unlock()

	if p.state == closed || p.state == errord {
		return nil
	}

	p.state = closed
	common.Must(p.done.Close())
	return nil
}

// Interrupt implements common.Interruptible.
func (p *pipe) Interrupt() {
	p.Lock()
	defer p.Unlock()

	if p.state == closed || p.state == errord {
		return
	}

	p.state = errord

	if !p.data.IsEmpty() {
		buf.ReleaseMulti(p.data)
		p.data = nil
	}

	common.Must(p.done.Close())
}
