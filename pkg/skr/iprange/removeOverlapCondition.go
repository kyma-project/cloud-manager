package iprange

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

func removeOverlapCondition(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, ctx
	}

	condError := meta.FindStatusCondition(*state.ObjAsIpRange().Conditions(), cloudresourcesv1beta1.ConditionTypeError)
	if condError == nil {
		return nil, ctx
	}

	if condError.Reason != cloudresourcesv1beta1.ConditionReasonCidrOverlap {
		return nil, ctx
	}

	meta.RemoveStatusCondition(state.ObjAsIpRange().Conditions(), cloudresourcesv1beta1.ConditionTypeError)
	state.ObjAsIpRange().Status.State = cloudresourcesv1beta1.StateProcessing

	return composed.UpdateStatus(state.ObjAsIpRange()).
		ErrorLogMessage("Error updating IpRange status after removed overlap error").
		SuccessLogMsg("Removed overlap error condition").
		SuccessError(composed.StopWithRequeue).
		Run(ctx, st)
}
