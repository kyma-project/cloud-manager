package wafpolicy

import (
	"context"

	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	if state.ObjAsWafPolicy().Status.Id != "" {
		return nil, ctx
	}

	id := uuid.NewString()

	if state.ObjAsWafPolicy().Labels == nil {
		state.ObjAsWafPolicy().Labels = map[string]string{}
	}
	state.ObjAsWafPolicy().Labels[cloudresourcesv1beta1.LabelId] = id

	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR WafPolicy with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR WafPolicy updated with ID label")

	return composed.NewStatusPatcherComposed(state.ObjAsWafPolicy()).
		MutateStatus(func(p *cloudresourcesv1beta1.WafPolicy) {
			p.Status.Id = id
		}).
		OnFailure(composed.Log("Error updating SKR WafPolicy with status ID"), composed.Requeue).
		OnSuccess(composed.RequeueAfter(rate.Default.Quick.When(state.ObjAsWafPolicy()))).
		OnStatusChanged(composed.Log("SKR WafPolicy updated with status ID")).
		Run(ctx, state.Cluster().K8sClient())
}
