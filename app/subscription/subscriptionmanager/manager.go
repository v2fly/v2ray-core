package subscriptionmanager

import (
	"archive/zip"
	"bytes"
	"context"
	"time"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/subscription"
	"github.com/v2fly/v2ray-core/v5/app/subscription/entries"
	"github.com/v2fly/v2ray-core/v5/app/subscription/entries/nonnative/nonnativeifce"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/features/extension"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type SubscriptionManagerImpl struct {
	config *subscription.Config
	ctx    context.Context

	s         *core.Instance
	converter *entries.ConverterRegistry

	trackedSubscriptions map[string]*trackedSubscription

	refreshTask *task.Periodic
}

func (s *SubscriptionManagerImpl) Type() interface{} {
	return extension.SubscriptionManagerType()
}

func (s *SubscriptionManagerImpl) housekeeping() error {
	for subscriptionName := range s.trackedSubscriptions {
		if err := s.checkupSubscription(subscriptionName); err != nil {
			newError("failed to checkup subscription: ", err).AtWarning().WriteToLog()
		}
	}
	return nil
}

func (s *SubscriptionManagerImpl) Start() error {
	if err := s.refreshTask.Start(); err != nil {
		return err
	}
	return nil
}

func (s *SubscriptionManagerImpl) Close() error {
	if err := s.refreshTask.Close(); err != nil {
		return err
	}
	return nil
}

func (s *SubscriptionManagerImpl) init() error {
	s.refreshTask = &task.Periodic{
		Interval: time.Duration(60) * time.Second,
		Execute:  s.housekeeping,
	}
	s.trackedSubscriptions = make(map[string]*trackedSubscription)
	s.converter = entries.GetOverlayConverterRegistry()
	if s.config.NonnativeConverterOverlay != nil {
		zipReader, err := zip.NewReader(bytes.NewReader(s.config.NonnativeConverterOverlay), int64(len(s.config.NonnativeConverterOverlay)))
		if err != nil {
			return newError("failed to read nonnative converter overlay: ", err)
		}
		converter, err := nonnativeifce.NewNonNativeConverterConstructor(zipReader)
		if err != nil {
			return newError("failed to construct nonnative converter: ", err)
		}
		if err := s.converter.RegisterConverter("user_nonnative", converter); err != nil {
			return newError("failed to register user nonnative converter: ", err)
		}
	}

	for _, v := range s.config.Imports {
		tracked, err := newTrackedSubscription(v)
		if err != nil {
			return newError("failed to init subscription ", v.Name, ": ", err)
		}
		s.trackedSubscriptions[v.Name] = tracked
	}
	return nil
}

func NewSubscriptionManager(ctx context.Context, config *subscription.Config) (*SubscriptionManagerImpl, error) {
	instance := core.MustFromContext(ctx)
	impl := &SubscriptionManagerImpl{ctx: ctx, s: instance, config: config}
	if err := impl.init(); err != nil {
		return nil, newError("failed to init subscription manager: ", err)
	}
	return impl, nil
}

func init() {
	common.Must(common.RegisterConfig((*subscription.Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewSubscriptionManager(ctx, config.(*subscription.Config))
	}))
}
