package network

import (
	"context"
	"errors"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// preventCreationOfNonCmManagedNetwork passes if obj is managed CM KCP Network, otherwise
// it will set error status and forget the KCP Network object. Currently, CM supports only
// creation of the cm managed networks
func preventCreationOfNonCmManagedNetwork(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*state)
	logger := composed.LoggerFromCtx(ctx)

	if common.IsKcpNetworkCM(state.ObjAsNetwork().Name, state.Scope().Name) {
		return nil, ctx
	}

	logger.Error(errors.New("not supported"), "Non-cm managed network can not be created")

	changed := false

	if state.ObjAsNetwork().Status.State != string(cloudcontrolv1beta1.ErrorState) {
		state.ObjAsNetwork().Status.State = string(cloudcontrolv1beta1.ErrorState)
		changed = true
	}

	if len(state.ObjAsNetwork().Status.Conditions) != 1 {
		changed = true
	}

	message := "Not supported"

	cond := meta.FindStatusCondition(state.ObjAsNetwork().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	if cond == nil {
		changed = true
	} else if cond.Status != metav1.ConditionTrue || cond.Reason != cloudcontrolv1beta1.ConditionTypeError || cond.Message != message {
		changed = true
	}

	if !changed {
		return composed.StopAndForget, ctx
	}

	return composed.PatchStatus(state.ObjAsNetwork()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ConditionTypeError,
			Message: message,
		}).
		ErrorLogMessage("Error patching KCP Network status with error on unsupported non-cm managed network creation").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
