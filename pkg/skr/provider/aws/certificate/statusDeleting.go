package certificate

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusDeleting(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	cert := state.ObjAsCertificate()

	return composed.NewStatusPatcherComposed(cert).
		MutateStatus(func(c *cloudresourcesv1beta1.Certificate) {
			c.SetStatusDeleting()
		}).
		OnSuccess(composed.Continue).
		OnFailure(composed.Log("Failed to patch Certificate status with deleting condition")).
		Run(ctx, state.Cluster().K8sClient())
}
