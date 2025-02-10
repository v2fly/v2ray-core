package subscriptionmanager

import "github.com/v2fly/v2ray-core/v5/app/subscription"

func (s *SubscriptionManagerImpl) AddTrackedSubscriptionFromImportSource(importSource *subscription.ImportSource) error {
	s.Lock()
	defer s.Unlock()
	return s.addTrackedSubscriptionFromImportSource(importSource, true)
}

func (s *SubscriptionManagerImpl) RemoveTrackedSubscription(name string) error {
	s.Lock()
	defer s.Unlock()
	return s.removeTrackedSubscription(name)
}

func (s *SubscriptionManagerImpl) UpdateTrackedSubscription(name string) error {
	s.Lock()
	defer s.Unlock()
	return s.updateSubscription(name)
}

func (s *SubscriptionManagerImpl) ListTrackedSubscriptions() []string {
	s.Lock()
	defer s.Unlock()

	var names []string
	for name := range s.trackedSubscriptions {
		names = append(names, name)
	}
	return names
}

func (s *SubscriptionManagerImpl) GetTrackedSubscriptionStatus(name string) (*subscription.TrackedSubscriptionStatus, error) {
	s.Lock()
	defer s.Unlock()
	if trackedSubscriptionItem, ok := s.trackedSubscriptions[name]; ok {
		result := &subscription.TrackedSubscriptionStatus{}
		if err := trackedSubscriptionItem.fillStatus(result); err != nil {
			return nil, newError("failed to fill status").Base(err)
		}
		result.ImportSource = trackedSubscriptionItem.importSource
		return result, nil
	} else {
		return nil, newError("unable to locate")
	}
}
