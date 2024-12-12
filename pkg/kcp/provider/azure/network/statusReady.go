package network

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statusReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	changed := false

	// status.networkType
	if state.ObjAsNetwork().Status.NetworkType != state.ObjAsNetwork().Spec.Type {
		state.ObjAsNetwork().Status.NetworkType = state.ObjAsNetwork().Spec.Type
		changed = true
	}

	// status.state
	if state.ObjAsNetwork().Status.State != string(cloudcontrolv1beta1.StateReady) {
		state.ObjAsNetwork().Status.State = string(cloudcontrolv1beta1.StateReady)
		changed = true
	}

	// more than one cond
	if len(state.ObjAsNetwork().Status.Conditions) != 1 {
		changed = true
	}

	// ready condition
	cond := meta.FindStatusCondition(*state.ObjAsNetwork().Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
	if cond == nil {
		changed = true
	} else if cond.Status != metav1.ConditionTrue || cond.Reason != cloudcontrolv1beta1.ReasonReady {
		changed = true
	}

	// status network reference
	if state.ObjAsNetwork().Status.Network == nil {
		state.ObjAsNetwork().Status.Network = &cloudcontrolv1beta1.NetworkReference{}
		changed = true
	}
	if state.ObjAsNetwork().Status.Network.Azure == nil {
		state.ObjAsNetwork().Status.Network.Azure = &cloudcontrolv1beta1.AzureNetworkReference{}
		changed = true
	}
	if state.ObjAsNetwork().Status.Network.Azure.TenantId != state.Scope().Spec.Scope.Azure.TenantId {
		state.ObjAsNetwork().Status.Network.Azure.TenantId = state.Scope().Spec.Scope.Azure.TenantId
		changed = true
	}
	if state.ObjAsNetwork().Status.Network.Azure.SubscriptionId != state.Scope().Spec.Scope.Azure.SubscriptionId {
		state.ObjAsNetwork().Status.Network.Azure.SubscriptionId = state.Scope().Spec.Scope.Azure.SubscriptionId
		changed = true
	}
	if state.ObjAsNetwork().Status.Network.Azure.ResourceGroup != state.resourceGroupName {
		state.ObjAsNetwork().Status.Network.Azure.ResourceGroup = state.resourceGroupName
		changed = true
	}
	if state.ObjAsNetwork().Status.Network.Azure.NetworkName != state.vnetName {
		state.ObjAsNetwork().Status.Network.Azure.NetworkName = state.vnetName
		changed = true
	}

	if !changed {
		return nil, nil
	}

	return composed.PatchStatus(state.ObjAsNetwork()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "Ready",
		}).
		ErrorLogMessage("Error patching Azure KCP Network status after setting ready condition").
		Run(ctx, state)
}
