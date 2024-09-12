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

	if state.ObjAsVpcPeering().Status.State != string(cloudcontrolv1beta1.ReadyState) {
		state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.ReadyState)
		changed = true
	}

	if len(state.ObjAsVpcPeering().Status.Conditions) != 1 {
		changed = true
	}
	cond := meta.FindStatusCondition(state.ObjAsVpcPeering().Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
	if cond == nil {
		changed = true
	} else if cond.Status != metav1.ConditionTrue || cond.Reason != cloudcontrolv1beta1.ConditionTypeReady || cond.Message != cloudcontrolv1beta1.ConditionTypeReady {
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

	if !changed {
		return nil, ctx
	}

	return composed.PatchStatus(state.ObjAsVpcPeering()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: cloudcontrolv1beta1.ReasonReady,
		}).
		ErrorLogMessage("Error patching KCP VpcPeering status to ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
