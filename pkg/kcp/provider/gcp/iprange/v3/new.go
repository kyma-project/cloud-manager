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
			PreventCidrEdit,
			CopyCidrToStatus,
			ValidateCidr,

			// Load remote resources
			LoadAddress,
			LoadPsaConnection,

			// Wait for any pending operations
			WaitOperationDone,

			// Branch based on deletion
			composed.IfElse(
				composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"create-update",
					actions.AddCommonFinalizer(),
					UpdateStatusId,

					// Create address if needed
					CreateAddress,

					// Identify peering IP ranges
					IdentifyPeeringIpRanges,

					// PSA connection management
					CreateOrUpdatePsaConnection,

					// Final status update
					UpdateStatus,
				),
				composed.ComposeActions(
					"delete",
					// Delete PSA connection first (if exists)
					DeletePsaConnection,

					// Then delete address
					DeleteAddress,

					// Remove finalizer and stop
					actions.RemoveCommonFinalizer(),
					composed.StopAndForgetAction,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}

// NewAllocateIpRangeAction returns an action suitable for allocation flow.
// This only provisions the GCP Address without PSA connection.
func NewAllocateIpRangeAction(stateFactory StateFactory) composed.Action {
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
			PreventCidrEdit,
			CopyCidrToStatus,
			ValidateCidr,

			// Prepare allocation-specific state
			PrepareAllocateIpRange,

			// Load remote resources
			LoadAddress,

			// Wait for any pending operations
			WaitOperationDone,

			// Branch based on deletion
			composed.IfElse(
				composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"allocate",
					actions.AddCommonFinalizer(),
					UpdateStatusId,

					// Identify peering IP ranges for allocation
					IdentifyPeeringIpRanges,

					// Create address if needed (no PSA connection in allocation flow)
					CreateAddress,

					// Final status update
					UpdateStatus,
				),
				composed.ComposeActions(
					"deallocate",
					// Delete address
					DeleteAddress,

					// Remove finalizer and stop
					actions.RemoveCommonFinalizer(),
					composed.StopAndForgetAction,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}
