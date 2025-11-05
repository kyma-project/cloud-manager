package vpcpeering

import (
	"context"
	"k8s.io/apimachinery/pkg/api/meta"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func kcpNetworkRemoteWait(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.IsMarkedForDeletion(state.Obj()) {
		return nil, ctx
	}

	if meta.IsStatusConditionTrue(state.remoteNetwork.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady) {
		return nil, ctx
	}

	if meta.IsStatusConditionTrue(state.remoteNetwork.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError) {
		changed := false

		if meta.RemoveStatusCondition(state.ObjAsVpcPeering().Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
			changed = true
		}

		if meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonWaitingDependency,
			Message: "Remote network not ready",
		}) {
			changed = true
		}

		if state.ObjAsVpcPeering().Status.State != string(cloudcontrolv1beta1.StateError) {
			state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateError)
			changed = true
		}

		if !changed {
			return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
		}

		return composed.PatchStatus(state.ObjAsVpcPeering()).
			ErrorLogMessage("Error patching KCP VpcPeering status with local network not ready").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			SuccessLogMsg("KCP VpcPeering local KCP Network not ready").
			Run(ctx, state)
	}

	logger.Info("Waiting KCP remote network ready state")

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
}
