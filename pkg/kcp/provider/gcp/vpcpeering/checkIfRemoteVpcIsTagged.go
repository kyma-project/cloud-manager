package vpcpeering

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	peeringconfig "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func checkIfRemoteVpcIsTagged(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	tags, err := state.client.GetRemoteNetworkTags(ctx, state.remoteNetwork.Status.Network.Gcp.NetworkName, state.remoteNetwork.Status.Network.Gcp.GcpProject)
	if err != nil {
		// This can be caused by multiple things, like missing permissions, network not found, etc.
		logger.Error(err, "[KCP GCP VPCPeering checkIfRemoteNetworkIsTagged] Error fetching GCP remote network tags")
		return composed.UpdateStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
				Message: "Error fetching GCP remote network tags",
			}).
			ErrorLogMessage("Error updating VPC Peering while fetching remote network tags").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(5*util.Timing.T60000ms())).
			Run(ctx, state)
	}

	var isVpcTagged bool

	for _, tag := range tags {
		if strings.Contains(tag, state.Scope().Spec.ShootName) || strings.Contains(tag, peeringconfig.VpcPeeringConfig.NetworkTag) {
			isVpcTagged = true
			break
		}
	}

	if !isVpcTagged {
		logger.Info("[KCP GCP VPCPeering checkIfRemoteNetworkIsTagged] Remote network is not tagged with the kyma shoot name", "shootName", state.Scope().Spec.ShootName)
		state.ObjAsVpcPeering().Status.State = cloudcontrolv1beta1.VirtualNetworkPeeringStateDisconnected
		return composed.UpdateStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
				Message: fmt.Sprintf("Error creating VpcPeering, remote VPC does not have a tag with the key: %s", state.Scope().Spec.ShootName),
			}).
			ErrorLogMessage("Error creating Remote VpcPeering").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(5*util.Timing.T60000ms())).
			Run(ctx, state)
	}

	return nil, nil
}
