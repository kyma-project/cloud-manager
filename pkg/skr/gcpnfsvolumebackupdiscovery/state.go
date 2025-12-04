package gcpnfsvolumebackupdiscovery

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client"
	"google.golang.org/api/file/v1"
	"k8s.io/klog/v2"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	SkrCluster composed.StateCluster

	Scope *cloudcontrolv1beta1.Scope

	fileBackupClient client.FileBackupClient

	backups []*file.Backup
}

type StateFactory interface {
	NewState(ctx context.Context, baseState composed.State) (*State, error)
}

func NewStateFactory(kymaRef klog.ObjectRef, kcpCluster composed.StateCluster, skrCluster composed.StateCluster,
	fileBackupClientProvider gcpclient.ClientProvider[client.FileBackupClient], env abstractions.Environment) StateFactory {

	return &stateFactory{
		kymaRef:                  kymaRef,
		kcpCluster:               kcpCluster,
		skrCluster:               skrCluster,
		fileBackupClientProvider: fileBackupClientProvider,
		env:                      env,
	}
}

type stateFactory struct {
	kymaRef                  klog.ObjectRef
	kcpCluster               composed.StateCluster
	skrCluster               composed.StateCluster
	fileBackupClientProvider gcpclient.ClientProvider[client.FileBackupClient]
	env                      abstractions.Environment
}

func (f *stateFactory) NewState(ctx context.Context, baseState composed.State) (*State, error) {
	fbc, err := f.fileBackupClientProvider(
		ctx,
		config.GcpConfig.CredentialsFile,
	)
	if err != nil {
		return nil, err
	}
	return &State{
		State:            baseState,
		KymaRef:          f.kymaRef,
		KcpCluster:       f.kcpCluster,
		SkrCluster:       f.skrCluster,
		fileBackupClient: fbc,
	}, nil
}

func (s *State) ObjAsGcpNfsVolumeBackupDiscovery() *cloudresourcesv1beta1.GcpNfsVolumeBackupDiscovery {
	return s.Obj().(*cloudresourcesv1beta1.GcpNfsVolumeBackupDiscovery)
}
