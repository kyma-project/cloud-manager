package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// New returns an Action that will provision and deprovision resource in the cloud.
// This follows the GcpSubnet pattern with clean action composition.
// Common post actions are executed after it in the common iprange flow
// so in the case of success it must return nil error as a signal of success.
// If it returns non-nil error then it will break the common iprange flow
// immediately so it must as well set the error conditions properly.
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
			preventCidrEdit,
			copyCidrToStatus,
			validateCidr,
			actions.AddCommonFinalizer(),

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
					// Create address if needed
					createAddress,
					waitOperationDone,
					updateStatusId,

					// PSA connection management
					identifyPeeringIpRanges,
					composed.IfElse(
						needsPsaConnection,
						composed.ComposeActions(
							"psa-connection",
							createOrUpdatePsaConnection,
							waitOperationDone,
						),
						nil,
					),

					// Final status update
					updateStatus,
				),
				composed.ComposeActions(
					"delete",
					// Delete PSA connection first (if exists)
					identifyPeeringIpRanges,
					deletePsaConnection,
					waitOperationDone,

					// Then delete address
					deleteAddress,
					waitOperationDone,

					// Remove finalizer and stop
					actions.RemoveCommonFinalizer(),
					composed.StopAndForgetAction,
				),
			),
		)(ctx, state)
	}
}

// NewAllocateIpRangeAction returns an Action that will populate state.ExistingCidrRanges
// with occupied cidr ranges so the allocation can pick a free slot.
func NewAllocateIpRangeAction(_ StateFactory) composed.Action {
	return prepareAllocateIpRange
}
