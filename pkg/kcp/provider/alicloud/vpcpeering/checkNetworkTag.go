package vpcpeering

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	peeringconfig "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func checkNetworkTag(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	if state.vpcPeering != nil {
		return nil, ctx
	}

	remoteVpcId := state.RemoteNetwork().Status.Network.Alicloud.VpcId
	tags, err := state.remoteClient.DescribeVpc(ctx, remoteVpcId)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error describing remote AliCloud VPC for tag check", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	shootName := state.Scope().Spec.ShootName
	_, hasNetworkTag := tags[peeringconfig.VpcPeeringConfig.NetworkTag]
	_, hasShootTag := tags[shootName]
	if hasNetworkTag || hasShootTag {
		return nil, ctx
	}

	logger.Info("Remote AliCloud VPC has no matching network tag", "remoteVpcId", remoteVpcId)

	return composed.PatchStatus(obj).
		SetCondition(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonFailedLoadingRemoteVpcNetwork,
			Message: "Loaded remote VPC has no matching tags",
		}).
		ErrorLogMessage("Error updating VpcPeering status due to remote vpc tag mismatch").
		FailedError(composed.StopWithRequeue).
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
		Run(ctx, state)
}
