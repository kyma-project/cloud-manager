package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	v2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// V2StateFactory is an alias for v2.StateFactory to be used by the reconciler.
type V2StateFactory = v2.StateFactory

// NewV2StateFactory is a wrapper for v2.NewStateFactory to be called from controller setup.
func NewV2StateFactory(
	serviceNetworkingClientProvider gcpclient.ClientProvider[gcpiprangeclient.ServiceNetworkingClient],
	oldComputeClientProvider gcpclient.ClientProvider[gcpiprangeclient.OldComputeClient],
	env abstractions.Environment,
) V2StateFactory {
	return v2.NewStateFactory(serviceNetworkingClientProvider, oldComputeClientProvider, env)
}

// New returns the action for GCP IpRange provisioning.
// It routes to either the refactored implementation or the legacy v2 implementation
// based on the ipRangeRefactored feature flag.
// Both state factories are passed from main.go to ensure proper provider wiring.
func New(refactoredStateFactory StateFactory, v2StateFactory v2.StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)

		// Check feature flag to determine which implementation to use
		if feature.IpRangeRefactored.Value(ctx) {
			logger.Info("Using refactored IpRange implementation (new)")
			return newRefactored(refactoredStateFactory)(ctx, st)
		}

		logger.Info("Using legacy IpRange implementation (v2)")
		return newLegacy(v2StateFactory)(ctx, st)
	}
}

// newRefactored is the new implementation following GcpSubnet pattern with clean action composition.
func newRefactored(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)

		// Convert shared iprange state to GCP-specific state
		state, err := stateFactory.NewState(ctx, st.(iprangetypes.State))
		if err != nil {
			logger.Error(err, "Failed to bootstrap GCP IpRange state")
			ipRange := st.Obj().(*v1beta1.IpRange)
			return composed.PatchStatus(ipRange).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonGcpError,
					Message: "Failed to create GCP IpRange state",
				}).
				SuccessError(composed.StopAndForget).
				SuccessLogMsg("Error creating new GCP IpRange state").
				Run(ctx, st)
		}

		return composed.ComposeActions(
			"gcpIpRange",
			// Validation and setup
			preventCidrEdit,
			copyCidrToStatus,
			validateCidr,

			// Load remote resources
			loadAddress,
			loadPsaConnection,

			// Wait for any pending operations
			waitOperationDone,

			// Branch based on deletion
			composed.IfElse(
				composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"create-update",
					actions.AddCommonFinalizer(),
					updateStatusId,

					// Create address if needed
					createAddress,

					// Identify peering IP ranges
					identifyPeeringIpRanges,

					// PSA connection management
					createOrUpdatePsaConnection,

					// Final status update
					updateStatus,
				),
				composed.ComposeActions(
					"delete",
					// Delete PSA connection first (if exists)
					deletePsaConnection,

					// Then delete address
					deleteAddress,

					// Remove finalizer and stop
					actions.RemoveCommonFinalizer(),
					composed.StopAndForgetAction,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}

// newLegacy wraps the v2 implementation for backward compatibility.
// This is used when the ipRangeRefactored feature flag is disabled.
// Note: v2 StateFactory should be passed from main.go with proper provider wiring.
func newLegacy(v2Factory v2.StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		return v2.New(v2Factory)(ctx, st)
	}
}

// NewAllocateIpRangeAction returns an action suitable for allocation flow.
// This only provisions the GCP Address without PSA connection.
// Routes to either refactored or legacy implementation based on feature flag.
// Both state factories are passed from main.go to ensure proper provider wiring.
func NewAllocateIpRangeAction(refactoredStateFactory StateFactory, v2StateFactory v2.StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)

		// Check feature flag to determine which implementation to use
		if feature.IpRangeRefactored.Value(ctx) {
			logger.Info("Using refactored IpRange allocation (new)")
			return newAllocateRefactored(refactoredStateFactory)(ctx, st)
		}

		logger.Info("Using legacy IpRange allocation (v2)")
		return newAllocateLegacy(v2StateFactory)(ctx, st)
	}
}

// newAllocateRefactored is the new allocation implementation.
func newAllocateRefactored(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)

		// Convert shared iprange state to GCP-specific state
		state, err := stateFactory.NewState(ctx, st.(iprangetypes.State))
		if err != nil {
			logger.Error(err, "Failed to bootstrap GCP IpRange state for allocation")
			ipRange := st.Obj().(*v1beta1.IpRange)
			return composed.PatchStatus(ipRange).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonGcpError,
					Message: "Failed to create GCP IpRange state",
				}).
				SuccessError(composed.StopAndForget).
				SuccessLogMsg("Error creating new GCP IpRange state for allocation").
				Run(ctx, st)
		}

		return composed.ComposeActions(
			"gcpIpRangeAllocation",
			// Validation and setup
			preventCidrEdit,
			copyCidrToStatus,
			validateCidr,

			// Prepare allocation-specific state
			prepareAllocateIpRange,

			// Load remote resources
			loadAddress,

			// Wait for any pending operations
			waitOperationDone,

			// Branch based on deletion
			composed.IfElse(
				composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"allocate",
					actions.AddCommonFinalizer(),
					updateStatusId,

					// Identify peering IP ranges for allocation
					identifyPeeringIpRanges,

					// Create address if needed (no PSA connection in allocation flow)
					createAddress,

					// Final status update
					updateStatus,
				),
				composed.ComposeActions(
					"deallocate",
					// Delete address
					deleteAddress,

					// Remove finalizer and stop
					actions.RemoveCommonFinalizer(),
					composed.StopAndForgetAction,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}

// newAllocateLegacy wraps the v2 allocation implementation for backward compatibility.
// Note: v2 StateFactory should be passed from main.go with proper provider wiring.
func newAllocateLegacy(v2Factory v2.StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		return v2.NewAllocateIpRangeAction(v2Factory)(ctx, st)
	}
}
