package gcpnfsvolumebackupdiscovery

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func clientCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	state.fileBackupClient = state.fileBackupClientProvider(state.Scope.Spec.Scope.Gcp.Project)

	return nil, ctx
}
