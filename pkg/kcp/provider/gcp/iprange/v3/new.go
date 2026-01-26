package v3

import (
	"context"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// New returns the action for GCP IpRange v3 provisioning.
// This is the refactored implementation following GcpSubnet pattern with clean action composition.
func New(stateFactory StateFactory) composed.Action {
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
					composed.If(
						needsPsaConnection,
						composed.IfElse(
							composed.Not(psaConnectionExists),
							createPsaConnection,
							updatePsaConnection,
						),
					),

					// Final status update
					updateStatus,
				),
				composed.ComposeActions(
					"delete",
					// Delete PSA connection first (if exists)
					deletePsaConnection,

					// Then delete address
					deleteAddress,

					// Remove finalizer
					actions.RemoveCommonFinalizer(),
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}

// NewAllocateIpRangeAction returns an action suitable for allocation flow.
// This populates ExistingCidrRanges with occupied CIDR ranges so the allocation
// can pick a free slot. This is called before the main provisioning flow.
func NewAllocateIpRangeAction(_ StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		// Similar to v2/AWS/Azure implementations: just prepare state for CIDR allocation
		// The actual address creation happens in the main New() flow later

		state := st.(iprangetypes.State)

		if len(state.Scope().Spec.Scope.Gcp.Network.Nodes) == 0 {
			state.ObjAsIpRange().Status.State = v1beta1.StateError
			return composed.PatchStatus(state.ObjAsIpRange()).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonCidrAllocationFailed,
					Message: "Error due to unknown SKR nodes range",
				}).
				ErrorLogMessage("Failed patching KCP IpRange status with error due to unknown SKR nodes range").
				SuccessLogMsg("Forgetting KCP IpRange in error state due to unknown SKR nodes range").
				Run(ctx, st)
		}

		state.SetExistingCidrRanges([]string{
			state.Scope().Spec.Scope.Gcp.Network.Nodes,
			state.Scope().Spec.Scope.Gcp.Network.Pods,
			state.Scope().Spec.Scope.Gcp.Network.Services,
		})

		return nil, ctx
	}
}
