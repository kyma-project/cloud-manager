package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	v2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v2"
	v3 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3"
)

// New returns the action for GCP IpRange provisioning.
// It routes to either the v3 refactored implementation or the v2 legacy implementation
// based on the ipRangeRefactored feature flag.
// Both state factories are passed from main.go to ensure proper provider wiring.
func New(v3StateFactory v3.StateFactory, v2StateFactory v2.StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)

		// Check feature flag to determine which implementation to use
		if feature.IpRangeRefactored.Value(ctx) {
			logger.Info("Using v3 refactored IpRange implementation")
			return v3.New(v3StateFactory)(ctx, st)
		}

		logger.Info("Using v2 legacy IpRange implementation")
		return v2.New(v2StateFactory)(ctx, st)
	}
}

// NewAllocateIpRangeAction returns an action suitable for allocation flow.
// This only provisions the GCP Address without PSA connection.
// Routes to either v3 refactored or v2 legacy implementation based on feature flag.
// Both state factories are passed from main.go to ensure proper provider wiring.
func NewAllocateIpRangeAction(v3StateFactory v3.StateFactory, v2StateFactory v2.StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)

		// Check feature flag to determine which implementation to use
		if feature.IpRangeRefactored.Value(ctx) {
			logger.Info("Using v3 refactored IpRange allocation")
			return v3.NewAllocateIpRangeAction(v3StateFactory)(ctx, st)
		}

		logger.Info("Using v2 legacy IpRange allocation")
		return v2.NewAllocateIpRangeAction(v2StateFactory)(ctx, st)
	}
}
