package gcpnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// createBackupClient constructs the V2 filestore backup client once the Scope is loaded.
// It must run after loadScope (state.Scope is nil at NewState time) and before loadBackup.
func createBackupClient(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.Scope == nil || state.Scope.Spec.Scope.Gcp == nil {
		return nil, ctx
	}

	state.fileBackupClient = state.fileBackupClientProvider(state.Scope.Spec.Scope.Gcp.Project)

	return nil, ctx
}
