package certificate

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	cert := state.ObjAsCertificate()

	return composed.NewStatusPatcherComposed(cert).
		MutateStatus(func(c *cloudresourcesv1beta1.Certificate) {
			if state.certificateArn != "" {
				c.Status.ProviderId = state.certificateArn
			}
			if state.certificateDetail != nil {
				c.SetStatusReady()
			}
		}).
		OnStatusChanged(composed.Log("Certificate status updated")).
		Run(ctx, state.Cluster().K8sClient())
}
