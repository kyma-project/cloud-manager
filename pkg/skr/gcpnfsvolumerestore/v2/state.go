package v2

import (
	"context"
	"fmt"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsbackupclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	v2client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client/v2"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/klog/v2"
)

// State represents the v2 state using modern GCP protobuf types
type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	SkrCluster composed.StateCluster

	Scope             *cloudcontrolv1beta1.Scope
	GcpNfsVolume      *cloudresourcesv1beta1.GcpNfsVolume
	SrcBackupFullPath string

	fileBackup *filestorepb.Backup // Modern protobuf type (for permission checks)

	fileRestoreClient v2client.FileRestoreClient
	fileBackupClient  gcpnfsbackupclientv2.FileBackupClient
}

type StateFactory interface {
	NewState(ctx context.Context, baseState composed.State) (*State, error)
}

func NewStateFactory(
	kymaRef klog.ObjectRef,
	kcpCluster composed.StateCluster,
	skrCluster composed.StateCluster,
	fileRestoreClientProvider gcpclient.GcpClientProvider[v2client.FileRestoreClient],
	fileBackupClientProvider gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient],
) StateFactory {
	return &stateFactory{
		kymaRef:                   kymaRef,
		kcpCluster:                kcpCluster,
		skrCluster:                skrCluster,
		fileRestoreClientProvider: fileRestoreClientProvider,
		fileBackupClientProvider:  fileBackupClientProvider,
	}
}

type stateFactory struct {
	kymaRef                   klog.ObjectRef
	kcpCluster                composed.StateCluster
	skrCluster                composed.StateCluster
	fileRestoreClientProvider gcpclient.GcpClientProvider[v2client.FileRestoreClient]
	fileBackupClientProvider  gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]
}

func (f *stateFactory) NewState(ctx context.Context, baseState composed.State) (*State, error) {
	return &State{
		State:             baseState,
		KymaRef:           f.kymaRef,
		KcpCluster:        f.kcpCluster,
		SkrCluster:        f.skrCluster,
		fileRestoreClient: f.fileRestoreClientProvider(),
		fileBackupClient:  f.fileBackupClientProvider(),
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

	allowedAll, existsAll := labels[ConvertToAccessibleFromKey("all")]
	if existsAll && allowedAll == util.GcpLabelBackupAccessibleFrom {
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
