package commonAction

import "sigs.k8s.io/controller-runtime/pkg/client"

type ObjReferringSubscription interface {
	client.Object

	SubscriptionName() string
}
