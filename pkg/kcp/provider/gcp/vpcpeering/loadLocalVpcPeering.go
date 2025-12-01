package vpcpeering

import (
	"context"

	pb "cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func loadLocalVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.localVpcPeering != nil ||
		(state.remotePeeringOperation != nil && state.remotePeeringOperation.Status == ptr.To(pb.Operation_DONE) && state.remotePeeringOperation.GetError().String() != "") {
		return nil, nil
	}

	kymaVpcPeering, err := state.client.GetVpcPeering(ctx, state.getKymaVpcPeeringName(), state.LocalNetwork().Status.Network.Gcp.GcpProject, state.LocalNetwork().Status.Network.Gcp.NetworkName)
	if err != nil {
		logger.Error(err, "Error loading Local Vpc Peering")
		state.ObjAsVpcPeering().Status.State = v1beta1.VirtualNetworkPeeringStateDisconnected
		meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonFailedCreatingVpcPeeringConnection,
			Message: "Error loading Local Vpc Peering",
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating status since it was not possible to load the Local Vpc Peering",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	if kymaVpcPeering != nil {
		state.localVpcPeering = kymaVpcPeering
		logger.Info("Local VPC Peering loaded", "localVpcPeering", state.getKymaVpcPeeringName())
	}

	return nil, nil
}
