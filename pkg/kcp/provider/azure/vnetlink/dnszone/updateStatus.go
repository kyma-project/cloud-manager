package dnszone

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	changed := false

	if state.vnetLink == nil ||
		state.vnetLink.Properties == nil ||
		state.vnetLink.Properties.VirtualNetworkLinkState == nil {
		return nil, ctx
	}

	if state.ObjAsAzureVNetLink().Status.State != string(*state.vnetLink.Properties.VirtualNetworkLinkState) {
		state.ObjAsAzureVNetLink().Status.State = string(*state.vnetLink.Properties.VirtualNetworkLinkState)
		changed = true
	}

	if meta.RemoveStatusCondition(state.ObjAsAzureVNetLink().Conditions(), cloudcontrolv1beta1.ConditionTypeError) {
		changed = true
	}

	if meta.SetStatusCondition(state.ObjAsAzureVNetLink().Conditions(), metav1.Condition{
		Type:    cloudcontrolv1beta1.ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  cloudcontrolv1beta1.ReasonReady,
		Message: cloudcontrolv1beta1.ReasonReady,
	}) {
		changed = true
	}

	if !changed {
		return nil, ctx
	}

	return composed.UpdateStatus(state.ObjAsAzureVNetLink()).
		ErrorLogMessage("Error updating KCP AzureVNetLink status to ready").
		SuccessLogMsg("AzureVNetLink status updated to ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
