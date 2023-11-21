package subscriptionmanager

import "time"

func (s *SubscriptionManagerImpl) checkupSubscription(subscriptionName string) error {
	var trackedSub *trackedSubscription
	if trackedSubFound, found := s.trackedSubscriptions[subscriptionName]; !found {
		return newError("not found")
	} else {
		trackedSub = trackedSubFound
	}

	shouldUpdate := false

	if trackedSub.currentDocumentExpireTime.Before(time.Now()) {
		shouldUpdate = true
	}

	if shouldUpdate {
		if err := s.updateSubscription(subscriptionName); err != nil {
			return newError("failed to update subscription: ", err)
		}
	}

	return nil
}
