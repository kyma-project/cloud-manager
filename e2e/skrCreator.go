package e2e

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
)

type SkrInfo struct {
}

type SkrCreator interface {
	CreateSkr(ctx context.Context, provider cloudcontrolv1beta1.ProviderType) (*SkrInfo, error)
}

func NewSkrCreator(world *World, subscription *SubscriptionInfo) SkrCreator {
	return &skrCreatorGardenerNetwork{
		world:        world,
		subscription: subscription,
	}
}
