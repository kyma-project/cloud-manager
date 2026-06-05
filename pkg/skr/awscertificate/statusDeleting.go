package awscertificate

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusDeleting(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	sp := composed.NewStatusPatcherComposed(state.ObjAsAwsCertificate())

	state.ObjAsAwsCertificate().SetStatusDeleting()

	return sp.
		OnSuccess(composed.Continue).
		OnFailure(composed.Log("Failed to patch AwsCertificate status with deleting condition")).
		Run(ctx, state.Cluster().K8sClient())
}
