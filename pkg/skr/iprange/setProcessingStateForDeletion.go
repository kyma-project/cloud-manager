package iprange

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func setProcessingStateForDeletion(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR IpRange is NOT marked for deletion, do not delete mirror in KCP
		return nil, ctx
	}
	if state.KcpIpRange == nil || composed.IsMarkedForDeletion(state.KcpIpRange) {
		return nil, ctx // KCP IpRange is already marked for deletion, so it already passed Processing state
	}

	// Skip setting Processing if a DeleteWhileUsed warning exists to avoid ping-pong with preventDeleteOn* actions
	warningCondition := meta.FindStatusCondition(*state.ObjAsIpRange().Conditions(), cloudresourcesv1beta1.ConditionTypeWarning)
	if warningCondition != nil && warningCondition.Status == metav1.ConditionTrue && warningCondition.Reason == cloudresourcesv1beta1.ConditionTypeDeleteWhileUsed {
		return nil, ctx
	}

	if state.ObjAsIpRange().State() != cloudresourcesv1beta1.StateProcessing {
		state.ObjAsIpRange().SetState(cloudresourcesv1beta1.StateProcessing)
		err := state.UpdateObjStatus(ctx)
		if client.IgnoreNotFound(err) != nil {
			// No reason to halt the flow if we can't set the processing state here as it will go to "deleting" or "error" state soon
			return composed.LogErrorAndReturn(err, "Error updating SKR IpRange status with Processing state", nil, ctx)
		}
	}

	return nil, ctx
}
