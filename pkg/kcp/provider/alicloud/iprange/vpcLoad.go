package iprange

import (
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func vpcLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.Network().Status.Network == nil || state.Network().Status.Network.Alicloud == nil {
		logger.Info("AliCloud network reference not yet available in Network status, requeueing")
		return composed.StopWithRequeue, ctx
	}

	networkName := state.Network().Status.Network.Alicloud.NetworkName

	vpcs, err := state.client.DescribeVpcs(ctx, networkName)
	if err != nil {
		logger.Error(err, "Error loading AliCloud VPC for IpRange")
		return composed.StopWithRequeue, ctx
	}

	if len(vpcs) == 0 {
		logger.Error(errors.New("no vpc"), "AliCloud VPC for IpRange not found", "networkName", networkName)
		state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:               cloudcontrolv1beta1.ConditionTypeError,
				Status:             metav1.ConditionTrue,
				ObservedGeneration: state.ObjAsIpRange().Generation,
				Reason:             cloudcontrolv1beta1.ReasonCloudProviderError,
				Message:            "VPC not found",
			}).
			ErrorLogMessage("Error patching AliCloud KCP IpRange status when VPC not found").
			Run(ctx, state)
	}

	state.vpcId = vpcs[0].VpcId

	return nil, ctx
}
