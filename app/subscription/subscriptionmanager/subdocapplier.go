package subscriptionmanager

import (
	"fmt"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/subscription/specs"
)

func (s *SubscriptionManagerImpl) applySubscriptionTo(name string, document *specs.SubscriptionDocument) error {
	var trackedSub *trackedSubscription
	if trackedSubFound, found := s.trackedSubscriptions[name]; !found {
		return newError("not found")
	} else {
		trackedSub = trackedSubFound
	}

	delta, err := trackedSub.diff(document)
	if err != nil {
		return err
	}

	nameToServerConfig := make(map[string]*specs.SubscriptionServerConfig)
	for _, server := range document.Server {
		nameToServerConfig[server.Id] = server
	}

	for _, serverName := range delta.removed {
		if err := s.removeManagedServer(name, serverName); err != nil {
			newError("failed to remove managed server: ", err).AtWarning().WriteToLog()
			continue
		}
		trackedSub.recordRemovedServer(serverName)
	}

	for _, serverName := range delta.modified {
		serverConfig := nameToServerConfig[serverName]
		if err := s.updateManagedServer(name, serverName, serverConfig); err != nil {
			newError("failed to update managed server: ", err).AtWarning().WriteToLog()
			continue
		}
		trackedSub.recordUpdatedServer(serverName, serverConfig.Metadata[ServerMetadataTagName], serverConfig)
	}

	for _, serverName := range delta.added {
		serverConfig := nameToServerConfig[serverName]
		if err := s.addManagedServer(name, serverName, serverConfig); err != nil {
			newError("failed to add managed server: ", err).AtWarning().WriteToLog()
			continue
		}
		trackedSub.recordUpdatedServer(serverName, serverConfig.Metadata[ServerMetadataTagName], serverConfig)
	}

	newError("finished applying subscription, ", name, "; ", fmt.Sprintf(
		"%v updated, %v added, %v removed, %v unchanged",
		len(delta.modified), len(delta.added), len(delta.removed), len(delta.unchanged))).AtInfo().WriteToLog()

	return nil
}

func (s *SubscriptionManagerImpl) removeManagedServer(subscriptionName, serverName string) error {
	var trackedSub *trackedSubscription
	if trackedSubFound, found := s.trackedSubscriptions[subscriptionName]; !found {
		return newError("not found")
	} else {
		trackedSub = trackedSubFound
	}

	var trackedServer *materializedServer
	if trackedServerFound, err := trackedSub.getCurrentServer(serverName); err != nil {
		return err
	} else {
		trackedServer = trackedServerFound
	}

	tagName := fmt.Sprintf("%s_%s", trackedSub.importSource.TagPrefix, trackedServer.tagPostfix)

	if err := core.RemoveOutboundHandler(s.s, tagName); err != nil {
		return newError("failed to remove handler: ", err)
	}
	trackedSub.recordRemovedServer(serverName)
	return nil
}

func (s *SubscriptionManagerImpl) addManagedServer(subscriptionName, serverName string,
	serverSpec *specs.SubscriptionServerConfig,
) error {
	var trackedSub *trackedSubscription
	if trackedSubFound, found := s.trackedSubscriptions[subscriptionName]; !found {
		return newError("not found")
	} else {
		trackedSub = trackedSubFound
	}
	tagPostfix := serverSpec.Metadata[ServerMetadataTagName]
	tagName := fmt.Sprintf("%s_%s", trackedSub.importSource.TagPrefix, tagPostfix)

	materialized, err := s.materialize(subscriptionName, tagName, serverSpec)
	if err != nil {
		return newError("failed to materialize server: ", err)
	}

	if err := core.AddOutboundHandler(s.s, materialized); err != nil {
		return newError("failed to add handler: ", err)
	}

	trackedSub.recordUpdatedServer(serverName, tagPostfix, serverSpec)

	return nil
}

func (s *SubscriptionManagerImpl) updateManagedServer(subscriptionName, serverName string,
	serverSpec *specs.SubscriptionServerConfig,
) error {
	if err := s.removeManagedServer(subscriptionName, serverName); err != nil {
		return newError("failed to update managed server: ", err).AtWarning()
	}
	if err := s.addManagedServer(subscriptionName, serverName, serverSpec); err != nil {
		return newError("failed to update managed server : ", err).AtWarning()
	}
	return nil
}
