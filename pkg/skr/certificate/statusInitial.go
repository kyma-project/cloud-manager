package certificate

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusInitial(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	sp := composed.NewStatusPatcherComposed(state.ObjAsCertificate())
	if !sp.IsStale() {
		return nil, ctx
	}

	return sp.
		MutateStatus(func(c *cloudresourcesv1beta1.Certificate) {
			c.SetStatusProcessing()
		}).
		OnSuccess(composed.Continue).
		Run(ctx, state.Cluster().K8sClient())
}
