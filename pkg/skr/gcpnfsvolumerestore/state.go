package gcpnfsvolumerestore

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsbackupclientv1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v1"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"google.golang.org/api/file/v1"

	"k8s.io/klog/v2"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	SkrCluster composed.StateCluster

	Scope             *cloudcontrolv1beta1.Scope
	GcpNfsVolume      *cloudresourcesv1beta1.GcpNfsVolume
	SrcBackupFullPath string

	fileBackup *file.Backup

	fileRestoreClient client.FileRestoreClient
	fileBackupClient  gcpnfsbackupclientv1.FileBackupClient
}

type StateFactory interface {
	NewState(ctx context.Context, baseState composed.State) (*State, error)
}

func NewStateFactory(kymaRef klog.ObjectRef, kcpCluster composed.StateCluster, skrCluster composed.StateCluster,
	fileRestoreClientProvider gcpclient.ClientProvider[client.FileRestoreClient],
	fileBackupClientProvider gcpclient.ClientProvider[gcpnfsbackupclientv1.FileBackupClient],
	env abstractions.Environment) StateFactory {

	return &stateFactory{
		kymaRef:                   kymaRef,
		kcpCluster:                kcpCluster,
		skrCluster:                skrCluster,
		fileRestoreClientProvider: fileRestoreClientProvider,
		fileBackupClientProvider:  fileBackupClientProvider,
		env:                       env,
	}
}

type stateFactory struct {
	kymaRef                   klog.ObjectRef
	kcpCluster                composed.StateCluster
	skrCluster                composed.StateCluster
	fileRestoreClientProvider gcpclient.ClientProvider[client.FileRestoreClient]
	fileBackupClientProvider  gcpclient.ClientProvider[gcpnfsbackupclientv1.FileBackupClient]
	env                       abstractions.Environment
}

func (f *stateFactory) NewState(ctx context.Context, baseState composed.State) (*State, error) {
	frc, err := f.fileRestoreClientProvider(ctx, config.GcpConfig.CredentialsFile)
	if err != nil {
		return nil, err
	}

	fbc, err := f.fileBackupClientProvider(ctx, config.GcpConfig.CredentialsFile)
	if err != nil {
		return nil, err
	}

	return &State{
		State:             baseState,
		KymaRef:           f.kymaRef,
		KcpCluster:        f.kcpCluster,
		SkrCluster:        f.skrCluster,
		fileRestoreClient: frc,
		fileBackupClient:  fbc,
	}, nil
}

func (s *State) ObjAsGcpNfsVolumeRestore() *cloudresourcesv1beta1.GcpNfsVolumeRestore {
	return s.Obj().(*cloudresourcesv1beta1.GcpNfsVolumeRestore)
}

func (s *State) IsAllowedToRestoreBackup() bool {
	if s.fileBackup == nil {
		return false
	}

	labels := s.fileBackup.Labels
	if labels == nil {
		return false
	}

	shootName := s.Scope.Spec.ShootName

	managed, exists := s.fileBackup.Labels[gcpclient.ManagedByKey]
	if !exists || managed != gcpclient.ManagedByValue {
		return false
	}

	allowed, exists := labels[ConvertToAccessibleFromKey(shootName)]
	if exists && allowed == util.GcpLabelBackupAccessibleFrom {
		return true
	}

	owner, exists := s.fileBackup.Labels[util.GcpLabelShootName]
	if exists && owner == shootName {
		return true
	}

	return false
}

func ConvertToAccessibleFromKey(name string) string {
	return fmt.Sprintf("cm-allow-%s", name)
}
