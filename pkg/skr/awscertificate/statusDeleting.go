package awscertificate

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusDeleting(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	cert := state.ObjAsAwsCertificate()

	sp := composed.NewStatusPatcherComposed(cert)

	sp.MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
		c.SetStatusDeleting()
	})

	//cert.SetStatusDeleting()

	return sp.
		OnSuccess(composed.Continue).
		OnFailure(composed.Log("Failed to patch AwsCertificate status with deleting condition")).
		Run(ctx, state.Cluster().K8sClient())
}
