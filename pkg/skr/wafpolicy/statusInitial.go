package wafpolicy

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusInitial(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	sp := composed.NewStatusPatcherComposed(state.ObjAsWafPolicy())
	if !sp.IsStale() {
		return nil, ctx
	}

	return sp.
		MutateStatus(func(obj *cloudresourcesv1beta1.WafPolicy) {
			obj.SetStatusProcessing()
		}).
		OnSuccess(composed.Log("Setting initial status on WafPolicy"), composed.Continue).
		OnFailure(composed.Log("Error setting initial status on WafPolicy")).
		Run(ctx, state.Cluster().K8sClient())
}
