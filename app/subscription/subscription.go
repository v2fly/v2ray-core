package subscription

import "github.com/ghxhy/v2ray-core/v5/features"

//go:generate go run github.com/ghxhy/v2ray-core/v5/common/errors/errorgen

type SubscriptionManager interface {
	features.Feature
	AddTrackedSubscriptionFromImportSource(importSource *ImportSource) error
	RemoveTrackedSubscription(name string) error
	ListTrackedSubscriptions() []string
	GetTrackedSubscriptionStatus(name string) (*TrackedSubscriptionStatus, error)
	UpdateTrackedSubscription(name string) error
}

func SubscriptionManagerType() interface{} {
	return (*SubscriptionManager)(nil)
}
