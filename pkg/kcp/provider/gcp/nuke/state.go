package nuke

import (
	"context"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsbackupclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client"
	"google.golang.org/api/file/v1"
)

type StateFactory interface {
	NewState(ctx context.Context, nukeState nuketypes.State) (focal.State, error)
}

func NewStateFactory(
	fileBackupClientProvider gcpclient.ClientProvider[gcpnfsbackupclient.FileBackupClient],
	env abstractions.Environment) StateFactory {
	return stateFactory{
		fileBackupClientProvider: fileBackupClientProvider,
		env:                      env,
	}
}

type stateFactory struct {
	fileBackupClientProvider gcpclient.ClientProvider[gcpnfsbackupclient.FileBackupClient]
	env                      abstractions.Environment
}

func (f stateFactory) NewState(ctx context.Context, nukeState nuketypes.State) (focal.State, error) {
	fbc, err := f.fileBackupClientProvider(
		ctx,
		f.env.Get("GCP_SA_JSON_KEY_PATH"),
	)
	if err != nil {
		return nil, err
	}
	return &State{
		State:            nukeState,
		fileBackupClient: fbc,
	}, nil
}

type State struct {
	nuketypes.State
	fileBackupClient  gcpnfsbackupclient.FileBackupClient
	ProviderResources []*nuketypes.ProviderResourceKindState
}

type GcpBackup struct {
	*file.Backup
}

func (b GcpBackup) GetId() string {
	return b.Name
}

func (b GcpBackup) GetObject() interface{} {
	return b.Backup
}

type ProviderNukeStatus struct {
	v1beta1.NukeStatus
}
