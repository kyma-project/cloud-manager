package gcpnfsvolumebackup

import (
	"context"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	client2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client"
	"google.golang.org/api/file/v1"
	"k8s.io/klog/v2"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	SkrCluster composed.StateCluster

	Scope       *v1beta1.Scope
	NfsInstance *v1beta1.NfsInstance

	fileBackup *file.Backup

	fileBackupClient client.FileBackupClient
	gcpConfig        *client2.GcpConfig
}

type StateFactory interface {
	NewState(ctx context.Context, baseState composed.State) (*State, error)
}

func NewStateFactory(kymaRef klog.ObjectRef, kcpCluster composed.StateCluster, skrCluster composed.StateCluster,
	fileBackupClientProvider client2.ClientProvider[client.FileBackupClient], env abstractions.Environment) StateFactory {

	return &stateFactory{
		kymaRef:                  kymaRef,
		kcpCluster:               kcpCluster,
		skrCluster:               skrCluster,
		fileBackupClientProvider: fileBackupClientProvider,
		env:                      env,
		gcpConfig:                client2.GetGcpConfig(env),
	}
}

type stateFactory struct {
	kymaRef                  klog.ObjectRef
	kcpCluster               composed.StateCluster
	skrCluster               composed.StateCluster
	fileBackupClientProvider client2.ClientProvider[client.FileBackupClient]
	env                      abstractions.Environment
	gcpConfig                *client2.GcpConfig
}

func (f *stateFactory) NewState(ctx context.Context, baseState composed.State) (*State, error) {
	fbc, err := f.fileBackupClientProvider(
		ctx,
		f.env.Get("GCP_SA_JSON_KEY_PATH"),
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
		gcpConfig:        f.gcpConfig,
	}, nil
}

func (s *State) ObjAsGcpNfsVolumeBackup() *cloudresourcesv1beta1.GcpNfsVolumeBackup {
	return s.Obj().(*cloudresourcesv1beta1.GcpNfsVolumeBackup)
}
