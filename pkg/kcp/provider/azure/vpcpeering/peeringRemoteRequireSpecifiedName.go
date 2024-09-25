package vpcpeering

import (
	"context"
	"errors"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func peeringRemoteRequireSpecifiedName(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.ObjAsVpcPeering().Spec.Details.PeeringName != "" {
		return nil, ctx
	}

	logger.Error(errors.New("peering name not specified"), "Invalid KCP VpcPeering")

	changed := false
	if state.ObjAsVpcPeering().Status.State != string(cloudcontrolv1beta1.ErrorState) {
		state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.ErrorState)
		changed = true
	}

	if len(state.ObjAsVpcPeering().Status.Conditions) != 1 {
		changed = true
	}

	conditionMessage := "Peering name not specified"
	cond := meta.FindStatusCondition(state.ObjAsVpcPeering().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	if cond == nil {
		changed = true
	} else if cond.Status != metav1.ConditionTrue || cond.Reason != cloudcontrolv1beta1.ConditionTypeError || cond.Message != conditionMessage {
		changed = true
	}

	if !changed {
		return composed.StopAndForget, ctx
	}

	return composed.PatchStatus(state.ObjAsVpcPeering()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ConditionTypeError,
			Message: conditionMessage,
		}).
		ErrorLogMessage("Error patching KCP VpcPeering with error state after missing peering name").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
