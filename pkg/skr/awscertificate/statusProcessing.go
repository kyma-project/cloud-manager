package awscertificate

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusProcessing(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	cert := state.ObjAsAwsCertificate()

	if state.certificateDetail != nil && state.certificateNeedsUpdate {
		return composed.NewStatusPatcherComposed(cert).
			MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
				c.SetStatusProcessing()
			}).
			OnSuccess(composed.Continue).
			OnFailure(composed.Log("Error setting processing status on AwsCertificate")).
			Run(ctx, state.Cluster().K8sClient())
	}

	return nil, ctx

}
