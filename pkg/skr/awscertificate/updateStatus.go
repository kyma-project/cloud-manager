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

	if state.certificateDetail == nil {
		return composed.NewStatusPatcherComposed(cert).
			MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
				// Update ARN from state
				c.Status.Arn = state.certificateArn
			}).
			OnStatusChanged(composed.Log("AwsCertificate ARN updated, requeing")).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	return composed.NewStatusPatcherComposed(cert).
		MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
			// Update expiration date from certificate detail
			if state.certificateDetail.NotAfter != nil {
				c.Status.ExpirationDate = &metav1.Time{
					Time: *state.certificateDetail.NotAfter,
				}
			}
			// Set Ready status (certificate detail is always loaded by importCertificate)
			c.SetStatusReady()
		}).
		OnStatusChanged(composed.Log("AwsCertificate is Ready")).
		Run(ctx, state.Cluster().K8sClient())
}
