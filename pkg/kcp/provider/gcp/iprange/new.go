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
	v3 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// V3StateFactory is an alias for v3.StateFactory to be used by the reconciler.
type V3StateFactory = v3.StateFactory

// NewV3StateFactory is a wrapper for v3.NewStateFactory to be called from controller setup.
func NewV3StateFactory(
	serviceNetworkingClientProvider gcpclient.GcpClientProvider[gcpiprangeclient.ServiceNetworkingClient],
	computeClientProvider gcpclient.GcpClientProvider[gcpiprangeclient.ComputeClient],
	env abstractions.Environment,
) V3StateFactory {
	return v3.NewStateFactory(serviceNetworkingClientProvider, computeClientProvider, env)
}

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
func New(v3StateFactory v3.StateFactory, v2StateFactory v2.StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)

		// Check feature flag to determine which implementation to use
		if feature.IpRangeRefactored.Value(ctx) {
			logger.Info("Using v3 refactored IpRange implementation")
			return newRefactored(v3StateFactory)(ctx, st)
		}

		logger.Info("Using v2 legacy IpRange implementation")
		return newLegacy(v2StateFactory)(ctx, st)
	}
}

// newRefactored is the v3 refactored implementation following GcpSubnet pattern with clean action composition.
func newRefactored(stateFactory v3.StateFactory) composed.Action {
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
			v3.PreventCidrEdit,
			v3.CopyCidrToStatus,
			v3.ValidateCidr,

			// Load remote resources
			v3.LoadAddress,
			v3.LoadPsaConnection,

			// Wait for any pending operations
			v3.WaitOperationDone,

			// Branch based on deletion
			composed.IfElse(
				composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"create-update",
					actions.AddCommonFinalizer(),
					v3.UpdateStatusId,

					// Create address if needed
					v3.CreateAddress,

					// Identify peering IP ranges
					v3.IdentifyPeeringIpRanges,

					// PSA connection management
					v3.CreateOrUpdatePsaConnection,

					// Final status update
					v3.UpdateStatus,
				),
				composed.ComposeActions(
					"delete",
					// Delete PSA connection first (if exists)
					v3.DeletePsaConnection,

					// Then delete address
					v3.DeleteAddress,

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
func NewAllocateIpRangeAction(v3StateFactory v3.StateFactory, v2StateFactory v2.StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)

		// Check feature flag to determine which implementation to use
		if feature.IpRangeRefactored.Value(ctx) {
			logger.Info("Using v3 refactored IpRange allocation")
			return newAllocateRefactored(v3StateFactory)(ctx, st)
		}

		logger.Info("Using v2 legacy IpRange allocation")
		return newAllocateLegacy(v2StateFactory)(ctx, st)
	}
}

// newAllocateRefactored is the v3 refactored allocation implementation.
func newAllocateRefactored(stateFactory v3.StateFactory) composed.Action {
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
			v3.PreventCidrEdit,
			v3.CopyCidrToStatus,
			v3.ValidateCidr,

			// Prepare allocation-specific state
			v3.PrepareAllocateIpRange,

			// Load remote resources
			v3.LoadAddress,

			// Wait for any pending operations
			v3.WaitOperationDone,

			// Branch based on deletion
			composed.IfElse(
				composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"allocate",
					actions.AddCommonFinalizer(),
					v3.UpdateStatusId,

					// Identify peering IP ranges for allocation
					v3.IdentifyPeeringIpRanges,

					// Create address if needed (no PSA connection in allocation flow)
					v3.CreateAddress,

					// Final status update
					v3.UpdateStatus,
				),
				composed.ComposeActions(
					"deallocate",
					// Delete address
					v3.DeleteAddress,

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
