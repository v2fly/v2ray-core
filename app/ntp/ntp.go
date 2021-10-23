//go:build !confonly
// +build !confonly

package ntp

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

import (
	"context"
	"time"

	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/app/ntp/ntptime"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/retry"
	"github.com/v2fly/v2ray-core/v4/common/session"
	"github.com/v2fly/v2ray-core/v4/common/task"
	"github.com/v2fly/v2ray-core/v4/features/extension"
	"github.com/v2fly/v2ray-core/v4/features/routing"
)

var _ extension.NTPClient = (*NTP)(nil)

type NTP struct {
	ctx      context.Context
	server   Server
	periodic *task.Periodic
	offset   time.Duration
}

func (s *NTP) Type() interface{} {
	return extension.NTPType()
}

func (s *NTP) Start() error {
	ntptime.Instance = s
	go s.periodic.Start()
	return nil
}

func (s *NTP) Close() error {
	if ntptime.Instance == s {
		ntptime.Instance = nil
	}
	return s.periodic.Close()
}

func (s *NTP) run() error {
	return retry.ExponentialBackoff(10, 1000).On(func() error {
		offset, err := s.server.QueryClockOffset()
		if err == nil {
			s.offset = offset
			newError("system clock offset: ", offset.String()).AtInfo().WriteToLog()
		} else {
			newError("failed to lookup time").Base(err).AtWarning().WriteToLog()
		}
		return err
	})
}

func (s *NTP) FixedNow() time.Time {
	return time.Now().Add(s.offset)
}

func New(ctx context.Context, config *Config) (*NTP, error) {
	s := &NTP{}
	destination := config.Address.AsDestination()
	/*if address := destination.Address; address.Family().IsDomain() {
		// place holder
	}*/
	err := core.RequireFeatures(ctx, func(dispatcher routing.Dispatcher) error {
		ctx = session.ContextWithInbound(ctx, &session.Inbound{
			Tag: config.InboundTag,
		})
		s.ctx = ctx
		s.server = NewClassicNTPClient(ctx, destination, dispatcher)
		s.periodic = &task.Periodic{
			Execute:  s.run,
			Interval: time.Second * time.Duration(config.SyncInterval),
		}
		return nil
	})
	return s, err
}

type Server interface {
	QueryClockOffset() (time.Duration, error)
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*Config))
	}))
}
