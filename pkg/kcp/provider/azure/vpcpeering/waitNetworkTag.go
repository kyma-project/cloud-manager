package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	peeringconfig "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func waitNetworkTag(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	if state.remotePeering != nil {
		return nil, nil
	}

	// If VpcNetwork is found but tags don't match user can recover by adding tag to remote VPC network so, we are
	// adding stop with requeue delay of one minute.

	_, hasShootTag := state.remoteVpc.Tags[peeringconfig.VpcPeeringConfig.NetworkTag]
	if !hasShootTag {
		_, hasShootTag = state.remoteVpc.Tags[state.Scope().Spec.ShootName]
	}

	if hasShootTag {
		return nil, ctx
	}

	var kv []any

	for k, v := range state.remoteVpc.Tags {
		kv = append(kv, k, v)
	}

	logger.Info("Loaded remote VPC network have no matching tags", kv...)

	changed := false

	if state.ObjAsVpcPeering().Status.State != string(cloudcontrolv1beta1.StateWarning) {
		state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateWarning)
		changed = true
	}

	if meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
		Type:    cloudcontrolv1beta1.ConditionTypeError,
		Status:  "True",
		Reason:  cloudcontrolv1beta1.ReasonFailedLoadingRemoteVpcNetwork,
		Message: "Loaded remote VPC network has no matching tags",
	}) {
		changed = true
	}

	successError := composed.StopWithRequeueDelay(util.Timing.T60000ms())

	if !changed {
		return successError, ctx
	}

	return composed.PatchStatus(obj).
		ErrorLogMessage("Error updating VpcPeering status due to remote VPC network tag mismatch").
		SuccessError(successError).
		Run(ctx, state)

}
