package defaultiprange

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

func waitIpRangeReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	isReady := meta.IsStatusConditionTrue(state.GetSkrIpRange().Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	if isReady {
		return nil, ctx
	}

	return composed.StopWithRequeue, ctx
}
