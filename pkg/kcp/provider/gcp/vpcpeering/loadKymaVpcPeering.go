package vpcpeering

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadKymaVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.kymaVpcPeering != nil {
		return nil, nil
	}

	logger.Info("Loading Kyma VPC Peering")

	//Using cm- prefix to make it clear it's a cloud manager resource
	//Using obj name suffix as the peering name since it is unique within the kcp namespace
	kymaVpcPeering, err := state.client.GetVpcPeering(ctx, state.getKymaVpcPeeringName(), state.localNetwork.Spec.Network.Reference.Gcp.GcpProject, state.localNetwork.Spec.Network.Reference.Gcp.NetworkName)
	if err != nil {
		logger.Error(err, "Error loading Kyma Vpc Peering")
		meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonFailedCreatingVpcPeeringConnection,
			Message: fmt.Sprintf("Error loading Kyma Vpc Peering: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating status since it was not possible to load the Kyma Vpc Peering",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	state.kymaVpcPeering = kymaVpcPeering
	return nil, nil
}
