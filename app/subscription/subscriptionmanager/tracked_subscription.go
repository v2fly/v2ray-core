package subscriptionmanager

import (
	"time"

	"github.com/v2fly/v2ray-core/v5/app/subscription"
	"github.com/v2fly/v2ray-core/v5/app/subscription/containers"
	"github.com/v2fly/v2ray-core/v5/app/subscription/specs"
)

func newTrackedSubscription(importSource *subscription.ImportSource) (*trackedSubscription, error) { //nolint: unparam
	return &trackedSubscription{importSource: importSource, materialized: map[string]*materializedServer{}}, nil
}

type trackedSubscription struct {
	importSource *subscription.ImportSource

	currentDocumentExpireTime time.Time
	currentDocument           *specs.SubscriptionDocument

	materialized map[string]*materializedServer

	originalDocument     []byte
	originalContainer    *containers.Container
	originalServerConfig map[string]*originalServerConfig

	addedByAPI bool
}

type originalServerConfig struct {
	data []byte
}

func (s *trackedSubscription) diff(newDocument *specs.SubscriptionDocument) (changedDocument, error) { //nolint: unparam
	delta := changedDocument{}
	seen := make(map[string]bool)

	for _, server := range newDocument.Server {
		if currentMaterialized, found := s.materialized[server.Id]; found {
			if currentMaterialized.serverConfig.Metadata[ServerMetadataFullyQualifiedName] == server.Metadata[ServerMetadataFullyQualifiedName] {
				delta.unchanged = append(delta.unchanged, server.Id)
			} else {
				delta.modified = append(delta.modified, server.Id)
			}
			seen[server.Id] = true
		} else {
			delta.added = append(delta.added, server.Id)
		}
	}

	for name := range s.materialized {
		if _, ok := seen[name]; !ok {
			delta.removed = append(delta.removed, name)
		}
	}

	return delta, nil
}

func (s *trackedSubscription) recordRemovedServer(name string) {
	delete(s.materialized, name)
}

func (s *trackedSubscription) recordUpdatedServer(name, tagPostfix string, serverConfig *specs.SubscriptionServerConfig) {
	s.materialized[name] = &materializedServer{tagPostfix: tagPostfix, serverConfig: serverConfig}
}

func (s *trackedSubscription) getCurrentServer(name string) (*materializedServer, error) {
	if materialized, found := s.materialized[name]; found {
		return materialized, nil
	} else {
		return nil, newError("not found")
	}
}

type materializedServer struct {
	tagPostfix string

	serverConfig *specs.SubscriptionServerConfig
}

func (s *trackedSubscription) fillStatus(status *subscription.TrackedSubscriptionStatus) error { //nolint: unparam
	status.ImportSource = s.importSource
	if s.currentDocument == nil {
		return nil
	}
	status.DocumentMetadata = s.currentDocument.Metadata
	status.Servers = make(map[string]*subscription.SubscriptionServer)
	for _, v := range s.currentDocument.Server {
		status.Servers[v.Id] = &subscription.SubscriptionServer{
			ServerMetadata: v.Metadata,
		}
		if materializedInstance, ok := s.materialized[v.Id]; ok {
			status.Servers[v.Id].Tag = materializedInstance.tagPostfix
		}
	}
	status.AddedByApi = s.addedByAPI
	return nil
}
