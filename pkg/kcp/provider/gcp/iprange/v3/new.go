package v3

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {

		state, err := stateFactory.NewState(ctx, st.(types.State))
		if err != nil {
			composed.LoggerFromCtx(ctx).Error(err, "Failed to bootstrap GCP IpRange state")
			privateIpRange := st.Obj().(*cloudcontrolv1beta1.IpRange)
			privateIpRange.Status.State = cloudcontrolv1beta1.StateError
			return composed.UpdateStatus(privateIpRange).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
					Message: "Failed to create IpRange state",
				}).
				SuccessError(composed.StopAndForget).
				SuccessLogMsg(fmt.Sprintf("Error creating new GCP IpRange state: %s", err)).
				Run(ctx, st)
		}

		return composed.ComposeActions(
			"privateIpRange",
			actions.AddCommonFinalizer(),
			loadSubnet,
			loadConnectionPolicy,
			composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"privateIpRange-create",
					createSubnet,
					createConnectionPolicy,
					updateStatusId,
					updateStatus,
				),
				composed.ComposeActions(
					"privateIpRange-delete",
					removeReadyCondition,
					deleteConnectionPolicy,
					deleteSubnet,
					actions.RemoveCommonFinalizer(),
					composed.StopAndForgetAction,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}
