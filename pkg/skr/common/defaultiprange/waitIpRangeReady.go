package defaultiprange

import (
	"context"
	"fmt"

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

	logger := composed.LoggerFromCtx(ctx)
	logger.
		WithValues("IpRange", fmt.Sprintf("%s/%s", state.GetSkrIpRange().Namespace, state.GetSkrIpRange().Name)).
		Info("IpRange is not ready, requeue delayed")

	return composed.StopWithRequeue, ctx
}
