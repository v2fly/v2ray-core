package observatory

import (
	"context"
	"github.com/v2fly/v2ray-core/v4/common/signal/done"
	"github.com/v2fly/v2ray-core/v4/features/extension"
	"github.com/v2fly/v2ray-core/v4/features/outbound"
	"sync"
)

type Observer struct {
	config *Config
	ctx    context.Context

	statusLock sync.Mutex
	status     []OutboundStatus

	finished *done.Instance

	ohm outbound.Manager
}

func (o *Observer) Type() interface{} {
	return extension.ObservatoryType()
}

func (o *Observer) Start() error {
	o.finished = done.New()
	go o.background()
	return nil
}

func (o *Observer) Close() error {
	return o.finished.Close()
}

func (o *Observer) background() {
	for !o.finished.Done() {
		hs, ok := o.ohm.(outbound.HandlerSelector)
		if !ok {
			newError("outbound.Manager is not a HandlerSelector").WriteToLog()
			return
		}
		outbounds := hs.Select(o.config.SubjectSelector)

	}
}
func (o *Observer) updateStatus(outbounds []string) {
	o.statusLock.Lock()
	defer o.statusLock.Unlock()

	o.status
}
