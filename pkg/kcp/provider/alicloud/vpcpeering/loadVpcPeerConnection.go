package vpcpeering

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcpeering/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadVpcPeerConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.vpcPeering != nil {
		return nil, ctx
	}

	obj := state.ObjAsVpcPeering()

	if obj.Status.Id != "" {
		info, err := state.client.GetVpcPeerConnection(ctx, obj.Status.Id)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error loading AliCloud VpcPeerConnection by ID", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
		if info != nil {
			state.vpcPeering = info
			return nil, ctx
		}
		// Status.Id is set but the cloud object is gone — surface error and stop.
		meta.SetStatusCondition(obj.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
			Message: fmt.Sprintf("VpcPeerConnection %s not found", obj.Status.Id),
		})
		obj.Status.State = string(cloudcontrolv1beta1.StateError)
		return composed.PatchStatus(obj).
			ErrorLogMessage("Error patching AliCloud VpcPeering status after connection not found").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	if state.localVpcId == "" {
		return nil, ctx
	}

	list, err := state.client.ListVpcPeerConnections(ctx, state.localVpcId, obj.GetName())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing AliCloud VpcPeerConnections", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	for _, item := range list {
		if alicloudvpcpeeringclient.IsTerminal(item.Status) {
			continue
		}
		// Re-get by id so state holds the full GetVpcPeerConnectionAttribute body (the list
		// projection omits AcceptingVpc/CIDRs that downstream route actions need).
		info, err := state.client.GetVpcPeerConnection(ctx, item.InstanceId)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error loading AliCloud VpcPeerConnection found by name", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
		if info != nil {
			state.vpcPeering = info
			break
		}
	}

	return nil, ctx
}
