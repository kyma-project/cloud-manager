package awscertificate

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	cert := state.ObjAsAwsCertificate()
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("Updating status to Ready")

	return composed.NewStatusPatcherComposed(cert).
		MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
			// Set ARN (set by importCertificate in state.certificateArn)
			if state.certificateArn != "" {
				c.Status.Arn = state.certificateArn
			}

			// Update expiration date if certificate detail is loaded
			if state.certificateDetail != nil {
				if state.certificateDetail.NotAfter != nil {
					c.Status.ExpirationDate = &metav1.Time{
						Time: ptr.Deref(state.certificateDetail.NotAfter, metav1.Now().Time),
					}
				}
			}
			// Set Ready status
			c.SetStatusReady()
		}).
		OnStatusChanged(composed.Log("AwsCertificate is Ready")).
		Run(ctx, state.Cluster().K8sClient())
}
