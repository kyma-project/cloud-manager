package awscertificate

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	cert := state.ObjAsAwsCertificate()

	return composed.NewStatusPatcherComposed(cert).
		MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
			// ALWAYS save ARN - set by loadCertificate or importCertificate
			if state.certificateArn != "" {
				c.Status.Arn = state.certificateArn
			}

			// Conditional: Update expiration and ready condition only when we have detail
			if state.certificateDetail != nil {
				if state.certificateDetail.NotAfter != nil {
					c.Status.ExpirationDate = &metav1.Time{
						Time: *state.certificateDetail.NotAfter,
					}
				}
				c.SetStatusReady()
			}
		}).
		OnStatusChanged(composed.Log("AwsCertificate status updated")).
		Run(ctx, state.Cluster().K8sClient())
}
