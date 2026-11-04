package vpcpeering

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcpeering/client"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateSuccessStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsVpcPeering()

	if state.vpcPeering == nil {
		return nil, ctx
	}

	changed := meta.RemoveStatusCondition(obj.Conditions(), cloudcontrolv1beta1.ConditionTypeError)

	remoteVpcId := alicloudvpcpeeringclient.RemoteVpcId(state.vpcPeering)
	if obj.Status.RemoteId != remoteVpcId {
		changed = true
		obj.Status.RemoteId = remoteVpcId
	}

	// Already provisioned and nothing changed: skip the patch to avoid re-triggering reconciliation.
	if !changed &&
		obj.Status.Id != "" &&
		obj.Status.RemoteId != "" &&
		meta.IsStatusConditionTrue(*obj.Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
		return nil, ctx
	}

	obj.Status.State = string(cloudcontrolv1beta1.StateReady)

	return composed.PatchStatus(obj).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "VpcPeering is provisioned",
		}).
		ErrorLogMessage("Error patching AliCloud VpcPeering with success status").
		SuccessLogMsg("AliCloud VpcPeering is ready").
		FailedError(composed.StopWithRequeue).
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
