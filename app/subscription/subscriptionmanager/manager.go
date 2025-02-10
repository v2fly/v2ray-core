package subscriptionmanager

import (
	"archive/zip"
	"bytes"
	"context"
	"sync"
	"time"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/persistentstorage"
	"github.com/v2fly/v2ray-core/v5/app/persistentstorage/protostorage"
	"github.com/v2fly/v2ray-core/v5/app/subscription"
	"github.com/v2fly/v2ray-core/v5/app/subscription/entries"
	"github.com/v2fly/v2ray-core/v5/app/subscription/entries/nonnative/nonnativeifce"
	"github.com/v2fly/v2ray-core/v5/app/subscription/specs"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/task"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type SubscriptionManagerImpl struct {
	sync.Mutex
	config *subscription.Config
	ctx    context.Context

	s         *core.Instance
	converter *entries.ConverterRegistry

	trackedSubscriptions map[string]*trackedSubscription

	refreshTask *task.Periodic

	persistentStorage               persistentstorage.ScopedPersistentStorage
	persistImportStorage            persistentstorage.ScopedPersistentStorage
	persistImportSourceProtoStorage protostorage.ProtoPersistentStorage
}

func (s *SubscriptionManagerImpl) Type() interface{} {
	return subscription.SubscriptionManagerType()
}

func (s *SubscriptionManagerImpl) housekeeping() error {
	for subscriptionName := range s.trackedSubscriptions {
		s.Lock()
		if err := s.checkupSubscription(subscriptionName); err != nil {
			newError("failed to checkup subscription: ", err).AtWarning().WriteToLog()
		}
		s.Unlock()
	}
	return nil
}

func (s *SubscriptionManagerImpl) Start() error {
	if s.config.Persistence {
		appEnvironment := envctx.EnvironmentFromContext(s.ctx).(environment.AppEnvironment)
		s.persistentStorage = appEnvironment.PersistentStorage()

		importsStorage, err := s.persistentStorage.NarrowScope(s.ctx, []byte("imports"))
		if err != nil {
			return newError("failed to get persistent storage for imports").Base(err)
		}
		s.persistImportStorage = importsStorage
		s.persistImportSourceProtoStorage = importsStorage.(protostorage.ProtoPersistentStorage)
		if err = s.loadAllFromPersistentStorage(); err != nil {
			newError("failed to load all from persistent storage: ", err).WriteToLog()
		}
	}
	go func() {
		if err := s.refreshTask.Start(); err != nil {
			return
		}
	}()
	return nil
}

func (s *SubscriptionManagerImpl) Close() error {
	if err := s.refreshTask.Close(); err != nil {
		return err
	}
	return nil
}

func (s *SubscriptionManagerImpl) addTrackedSubscriptionFromImportSource(importSource *subscription.ImportSource,
	addedByAPI bool,
) error {
	if s.config.Persistence && addedByAPI {
		err := s.persistImportSourceProtoStorage.PutProto(s.ctx, importSource.Name, importSource)
		if err != nil {
			return newError("failed to persist import source: ", err)
		}
	}

	tracked, err := newTrackedSubscription(importSource)
	if err != nil {
		return newError("failed to init subscription ", importSource.Name, ": ", err)
	}
	tracked.addedByAPI = addedByAPI
	s.trackedSubscriptions[importSource.Name] = tracked
	return nil
}

func (s *SubscriptionManagerImpl) removeTrackedSubscription(subscriptionName string) error {
	if s.config.Persistence {
		err := s.persistImportStorage.Put(s.ctx, []byte(subscriptionName), nil)
		if err != nil {
			return newError("failed to delete import source: ", err)
		}
	}
	if _, ok := s.trackedSubscriptions[subscriptionName]; ok {
		err := s.applySubscriptionTo(subscriptionName, &specs.SubscriptionDocument{Server: make([]*specs.SubscriptionServerConfig, 0)})
		if err != nil {
			return newError("failed to apply empty subscription: ", err)
		}
		delete(s.trackedSubscriptions, subscriptionName)
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
		if err := s.addTrackedSubscriptionFromImportSource(v, false); err != nil {
			return newError("failed to add tracked subscription: ", err)
		}
	}
	return nil
}

func (s *SubscriptionManagerImpl) loadAllFromPersistentStorage() error {
	if !s.config.Persistence {
		return nil
	}
	protoImportSources, err := s.persistImportStorage.List(s.ctx, []byte(""))
	if err != nil {
		return newError("failed to list import sources: ", err)
	}
	for _, protoImportSource := range protoImportSources {
		var importSource subscription.ImportSource
		err := s.persistImportSourceProtoStorage.GetProto(s.ctx, string(protoImportSource), &importSource)
		if err != nil {
			return newError("failed to get import source: ", err)
		}
		if err := s.addTrackedSubscriptionFromImportSource(&importSource, false); err != nil {
			return newError("failed to add tracked subscription: ", err)
		}
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
