package vpcpeering

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createVpcPeerConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vpcPeering != nil {
		return nil, ctx
	}

	obj := state.ObjAsVpcPeering()
	remoteNet := state.RemoteNetwork()

	if remoteNet.Status.Network == nil || remoteNet.Status.Network.Alicloud == nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("remote network Alicloud status not populated yet"),
			"Remote network Alicloud status is nil, requeueing",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()),
			ctx,
		)
	}

	if state.localVpcId == "" {
		return composed.LogErrorAndReturn(
			fmt.Errorf("local VPC ID not yet available from VpcNetwork"),
			"Local VPC ID is empty, requeueing",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()),
			ctx,
		)
	}

	remoteVpcId := remoteNet.Status.Network.Alicloud.VpcId
	remoteAccountId := remoteNet.Status.Network.Alicloud.AccountId

	logger.Info("Creating AliCloud VpcPeerConnection",
		"localVpcId", state.localVpcId,
		"remoteVpcId", remoteVpcId,
		"remoteAccountId", remoteAccountId,
		"remoteRegion", state.remoteRegion,
	)

	instanceId, err := state.client.CreateVpcPeerConnection(
		ctx,
		state.localVpcId,
		remoteAccountId,
		state.remoteRegion,
		remoteVpcId,
		obj.GetName(),
	)
	if err != nil {
		meta.RemoveStatusCondition(&obj.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
		meta.SetStatusCondition(&obj.Status.Conditions, metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
			Message: fmt.Sprintf("Failed creating VpcPeerConnection: %s", err.Error()),
		})
		obj.Status.State = string(cloudcontrolv1beta1.StateWarning)
		return composed.PatchStatus(obj).
			ErrorLogMessage("Error patching AliCloud VpcPeering status after create failure").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	obj.Status.Id = instanceId
	obj.Status.VpcId = state.localVpcId

	return composed.PatchStatus(obj).
		ErrorLogMessage("Error patching AliCloud VpcPeering status with connection ID").
		FailedError(composed.StopWithRequeue).
		SuccessErrorNil().
		Run(ctx, state)
}
