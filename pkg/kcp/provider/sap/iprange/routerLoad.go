package iprange

import (
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func routerLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	router, err := state.sapClient.GetRouterByName(ctx, state.Network().Status.Network.OpenStack.NetworkName)
	if err != nil {
		logger.Error(err, "Error loading SAP KCP IpRange openstack router")
		return composed.StopWithRequeue, ctx
	}

	if router == nil {
		logger.Error(errors.New("no router"), "SAP KCP IpRange openstack router not found")
		state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:               cloudcontrolv1beta1.ConditionTypeError,
				Status:             metav1.ConditionTrue,
				ObservedGeneration: state.ObjAsIpRange().Generation,
				Reason:             cloudcontrolv1beta1.ReasonCloudProviderError,
				Message:            "Router not found",
			}).
			ErrorLogMessage("Error patching SAP KCP IpRange status with error router not found").
			Run(ctx, state)
	}

	state.router = router

	return nil, ctx
}
