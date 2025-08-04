package e2e

import (
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
)

type SubscriptionInfo struct {
	Name      string
	Provider  cloudcontrolv1beta1.ProviderType
	IsDefault bool
}

var SubscriptionRegistry = &SubscriptionRegistryType{
	subscriptions: make(map[string]*SubscriptionInfo),
}

type SubscriptionRegistryType struct {
	subscriptions map[string]*SubscriptionInfo
}

func (sr *SubscriptionRegistryType) AddSubscription(name string, subscription *SubscriptionInfo) error {
	_, alreadyAdded := sr.subscriptions[name]
	if alreadyAdded {
		return fmt.Errorf("subscription %s already added", name)
	}
	sr.subscriptions[name] = subscription
	return nil
}

func (sr *SubscriptionRegistryType) FindFirst(cb func(s *SubscriptionInfo) bool) *SubscriptionInfo {
	for _, sub := range sr.subscriptions {
		if cb(sub) {
			return sub
		}
	}
	return nil
}

func (sr *SubscriptionRegistryType) GetDefaultForProvider(provider cloudcontrolv1beta1.ProviderType) *SubscriptionInfo {
	return sr.FindFirst(func(s *SubscriptionInfo) bool {
		return s.Provider == provider && s.IsDefault
	})
}
