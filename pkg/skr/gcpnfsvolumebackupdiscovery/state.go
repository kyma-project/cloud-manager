package gcpnfsvolumebackupdiscovery

import (
	"context"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsbackupclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
	"k8s.io/klog/v2"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	SkrCluster composed.StateCluster

	Scope *cloudcontrolv1beta1.Scope

	fileBackupClientProvider gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]
	fileBackupClient         gcpnfsbackupclientv2.FileBackupClient

	backups []*filestorepb.Backup
}

type StateFactory interface {
	NewState(ctx context.Context, baseState composed.State) (*State, error)
}

func NewStateFactory(scopeProvider scopeprovider.ScopeProvider, kcpCluster composed.StateCluster, skrCluster composed.StateCluster,
	fileBackupClientProvider gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]) StateFactory {

	return &stateFactory{
		scopeProvider:            scopeProvider,
		kcpCluster:               kcpCluster,
		skrCluster:               skrCluster,
		fileBackupClientProvider: fileBackupClientProvider,
	}
}

type stateFactory struct {
	scopeProvider            scopeprovider.ScopeProvider
	kcpCluster               composed.StateCluster
	skrCluster               composed.StateCluster
	fileBackupClientProvider gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]
}

func (f *stateFactory) NewState(ctx context.Context, baseState composed.State) (*State, error) {
	kymaRef, err := f.scopeProvider.GetScope(ctx, baseState.Name())
	if err != nil {
		return nil, err
	}
	return &State{
		State:                    baseState,
		KymaRef:                  kymaRef,
		KcpCluster:               f.kcpCluster,
		SkrCluster:               f.skrCluster,
		fileBackupClientProvider: f.fileBackupClientProvider,
	}, nil
}

func (s *State) ObjAsGcpNfsVolumeBackupDiscovery() *cloudresourcesv1beta1.GcpNfsVolumeBackupDiscovery {
	return s.Obj().(*cloudresourcesv1beta1.GcpNfsVolumeBackupDiscovery)
}
