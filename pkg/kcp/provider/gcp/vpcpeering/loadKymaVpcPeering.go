package vpcpeering

import (
	"context"
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

	logger.Info("[KCP GCP VpcPeering loadKymaVpcPeering] Loading Kyma VPC Peering")

	kymaVpcPeering, err := state.client.GetVpcPeering(ctx, state.getKymaVpcPeeringName(), state.localNetwork.Status.Network.Gcp.GcpProject, state.localNetwork.Status.Network.Gcp.NetworkName)
	if err != nil {
		logger.Error(err, "Error loading Kyma Vpc Peering")
		state.ObjAsVpcPeering().Status.State = v1beta1.VirtualNetworkPeeringStateDisconnected
		meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonFailedCreatingVpcPeeringConnection,
			Message: "Error loading Kyma Vpc Peering",
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating status since it was not possible to load the Kyma Vpc Peering",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	logger.Info("[KCP GCP VpcPeering loadKymaVpcPeering] Kyma VPC Peering loaded")
	state.kymaVpcPeering = kymaVpcPeering
	return nil, nil
}
