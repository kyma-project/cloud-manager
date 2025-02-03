package network

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	networktypes "github.com/kyma-project/cloud-manager/pkg/kcp/network/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func handleNetworkReference(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(networktypes.State)

	if state.ObjAsNetwork().Spec.Network.Reference == nil {
		return nil, ctx
	}

	// prevent delete if used must come before this action!!!

	// Handle a special case for a case when Scope is deleted and related kyma Network is left behind
	// This should clean up orphan networks that were not deleted.
	// For the future, the Scope reconciler is also modified to delete the kyma Network it creates
	// at the moment it delete itself, so this should not happen for those new cases
	if state.Scope() == nil && state.ObjAsNetwork().Spec.Type == cloudcontrolv1beta1.NetworkTypeKyma && !composed.MarkedForDeletionPredicate(ctx, state) {
		err := state.Cluster().K8sClient().Delete(ctx, state.ObjAsNetwork())
		if client.IgnoreNotFound(err) != nil {
			return composed.LogErrorAndReturn(err, "Error deleting KCP kyma Network", composed.StopWithRequeueDelay(util.Timing.T300000ms()), ctx)
		}
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
	}

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return composed.ComposeActions(
			"networkReferenceDelete",
			actions.PatchRemoveCommonFinalizer(),
			composed.StopAndForgetAction,
		)(ctx, state)
	}

	changed := false

	if state.ObjAsNetwork().Status.Network == nil || !state.ObjAsNetwork().Spec.Network.Reference.Equals(state.ObjAsNetwork().Status.Network) {
		state.ObjAsNetwork().Status.Network = state.ObjAsNetwork().Spec.Network.Reference.DeepCopy()
		changed = true
	}

	if state.ObjAsNetwork().Status.State != string(cloudcontrolv1beta1.StateReady) {
		state.ObjAsNetwork().Status.State = string(cloudcontrolv1beta1.StateReady)
		changed = true
	}

	if meta.RemoveStatusCondition(state.ObjAsNetwork().Conditions(), cloudcontrolv1beta1.ConditionTypeError) {
		changed = true
	}
	if meta.SetStatusCondition(state.ObjAsNetwork().Conditions(), metav1.Condition{
		Type:    cloudcontrolv1beta1.ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  cloudcontrolv1beta1.ReasonReady,
		Message: cloudcontrolv1beta1.ReasonReady,
	}) {
		changed = true
	}

	switch {
	case state.ObjAsNetwork().Status.Network.Azure != nil:
		if state.ObjAsNetwork().Status.Network.Azure.SubscriptionId == "" {
			state.ObjAsNetwork().Status.Network.Azure.SubscriptionId = state.Scope().Spec.Scope.Azure.SubscriptionId
			changed = true
		}
		if state.ObjAsNetwork().Status.Network.Azure.TenantId == "" {
			state.ObjAsNetwork().Status.Network.Azure.TenantId = state.Scope().Spec.Scope.Azure.TenantId
			changed = true
		}
	case state.ObjAsNetwork().Status.Network.Aws != nil:
		if state.ObjAsNetwork().Status.Network.Aws.AwsAccountId == "" {
			state.ObjAsNetwork().Status.Network.Aws.AwsAccountId = state.Scope().Spec.Scope.Aws.AccountId
			changed = true
		}
	case state.ObjAsNetwork().Status.Network.Gcp != nil:
		if state.ObjAsNetwork().Status.Network.Gcp.GcpProject == "" {
			state.ObjAsNetwork().Status.Network.Gcp.GcpProject = state.Scope().Spec.Scope.Gcp.Project
			changed = true
		}
	case state.ObjAsNetwork().Status.Network.OpenStack != nil:
		if state.ObjAsNetwork().Status.Network.OpenStack.Domain == "" {
			state.ObjAsNetwork().Status.Network.OpenStack.Domain = state.Scope().Spec.Scope.OpenStack.DomainName
			changed = true
		}
		if state.ObjAsNetwork().Status.Network.OpenStack.Project == "" {
			state.ObjAsNetwork().Status.Network.OpenStack.Project = state.Scope().Spec.Scope.OpenStack.TenantName
			changed = true
		}
	}

	if !changed {
		return composed.StopAndForget, nil
	}

	return composed.PatchStatus(state.ObjAsNetwork()).
		ErrorLogMessage("Error patching Network status after reference copy").
		Run(ctx, state)
}
