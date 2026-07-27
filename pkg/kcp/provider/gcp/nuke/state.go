package nuke

import (
	"context"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsbackupclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
)

// kindFilestoreBackup is the nuke ProviderResourceKindState.Kind for GCP filestore backups.
// Shared by loadNfsBackups (producer) and deleteNfsBackup (consumer) to prevent drift.
const kindFilestoreBackup = "FilestoreBackup"

type StateFactory interface {
	NewState(ctx context.Context, nukeState nuketypes.State) (focal.State, error)
}

func NewStateFactory(
	fileBackupClientProvider gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient],
) StateFactory {
	return stateFactory{
		fileBackupClientProvider: fileBackupClientProvider,
	}
}

type stateFactory struct {
	fileBackupClientProvider gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]
}

func (f stateFactory) NewState(ctx context.Context, nukeState nuketypes.State) (focal.State, error) {
	// The KCP nuke scope is loaded by focal.NewWithOptionalScope() before this factory fires,
	// and this factory only runs behind GcpProviderPredicate, so the GCP project is available.
	fbc := f.fileBackupClientProvider(nukeState.Scope().Spec.Scope.Gcp.Project)
	return &State{
		State:            nukeState,
		fileBackupClient: fbc,
	}, nil
}

type State struct {
	nuketypes.State
	fileBackupClient  gcpnfsbackupclientv2.FileBackupClient
	ProviderResources []*nuketypes.ProviderResourceKindState
}

type GcpBackup struct {
	*filestorepb.Backup
}

func (b GcpBackup) GetId() string {
	return b.GetName()
}

func (b GcpBackup) GetObject() any {
	return b.Backup
}

type ProviderNukeStatus struct {
	v1beta1.NukeStatus
}
