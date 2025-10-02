package iprange

import (
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func networkLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	net, err := state.sapClient.GetNetworkByName(ctx, state.Network().Status.Network.OpenStack.NetworkName)
	if err != nil {
		logger.Error(err, "Error loading Openstack network for IpRange")
		return composed.StopWithRequeue, ctx
	}

	if net == nil {
		logger.Error(errors.New("no network"), "Openstack network for IpRange not found")
		state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:               cloudcontrolv1beta1.ConditionTypeError,
				Status:             metav1.ConditionTrue,
				ObservedGeneration: state.ObjAsIpRange().Generation,
				Reason:             cloudcontrolv1beta1.ReasonCloudProviderError,
				Message:            "Network not found",
			}).
			ErrorLogMessage("Error patching SAP KCP IpRange status with error when network not found").
			Run(ctx, state)
	}

	state.net = net

	return nil, ctx
}
