package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func statusReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	changed := false

	if state.ObjAsVpcPeering().Status.State != cloudcontrolv1beta1.VirtualNetworkPeeringStateConnected {
		state.ObjAsVpcPeering().Status.State = cloudcontrolv1beta1.VirtualNetworkPeeringStateConnected
		changed = true
	}

	if state.ObjAsVpcPeering().Status.Id != ptr.Deref(state.localPeering.ID, "") {
		state.ObjAsVpcPeering().Status.Id = ptr.Deref(state.localPeering.ID, "")
		changed = true
	}

	if state.ObjAsVpcPeering().Status.RemoteId != ptr.Deref(state.remotePeering.ID, "") {
		state.ObjAsVpcPeering().Status.RemoteId = ptr.Deref(state.remotePeering.ID, "")
		changed = true
	}

	if meta.RemoveStatusCondition(state.ObjAsVpcPeering().Conditions(), cloudcontrolv1beta1.ConditionTypeError) {
		changed = true
	}

	condition := metav1.Condition{
		Type:    cloudcontrolv1beta1.ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  cloudcontrolv1beta1.ReasonReady,
		Message: cloudcontrolv1beta1.ReasonReady,
	}

	if meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), condition) {
		changed = true
	}

	if !changed {
		return nil, ctx
	}

	return composed.PatchStatus(state.ObjAsVpcPeering()).
		ErrorLogMessage("Error patching KCP VpcPeering status to ready").
		SuccessLogMsg("Success patching KCP VpcPeering status to ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
