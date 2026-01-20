package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	gcpiprangev2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v2"
	gcpiprangev3 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3"
)

// New returns the action for GCP IpRange provisioning.
// It routes to either the v3 refactored implementation or the v2 legacy implementation
// based on the gcpIpRangeV3 feature flag.
// Both state factories are passed from main.go to ensure proper provider wiring.
func New(v3StateFactory gcpiprangev3.StateFactory, v2StateFactory gcpiprangev2.StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)

		// Check feature flag to determine which implementation to use
		if feature.GcpIpRangeV3.Value(ctx) {
			logger.Info("Using v3 refactored IpRange implementation")
			return gcpiprangev3.New(v3StateFactory)(ctx, st)
		}

		logger.Info("Using v2 legacy IpRange implementation")
		return gcpiprangev2.New(v2StateFactory)(ctx, st)
	}
}

// NewAllocateIpRangeAction returns an action suitable for allocation flow.
// This populates ExistingCidrRanges with occupied CIDR ranges so the allocation
// can pick a free slot. Routes to either v3 refactored or v2 legacy implementation
// based on feature flag. Note: State factories not needed for allocation action.
func NewAllocateIpRangeAction(v3StateFactory gcpiprangev3.StateFactory, v2StateFactory gcpiprangev2.StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)

		// Check feature flag to determine which implementation to use
		if feature.GcpIpRangeV3.Value(ctx) {
			logger.Info("Using v3 refactored IpRange allocation")
			return gcpiprangev3.NewAllocateIpRangeAction(v3StateFactory)(ctx, st)
		}

		logger.Info("Using v2 legacy IpRange allocation")
		return gcpiprangev2.NewAllocateIpRangeAction(v2StateFactory)(ctx, st)
	}
}
