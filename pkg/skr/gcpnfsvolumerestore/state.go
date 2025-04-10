package gcpnfsvolumerestore

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client"
	"k8s.io/klog/v2"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	SkrCluster composed.StateCluster

	Scope              *cloudcontrolv1beta1.Scope
	GcpNfsVolume       *cloudresourcesv1beta1.GcpNfsVolume
	GcpNfsVolumeBackup *cloudresourcesv1beta1.GcpNfsVolumeBackup

	fileRestoreClient client.FileRestoreClient
}

type StateFactory interface {
	NewState(ctx context.Context, baseState composed.State) (*State, error)
}

func NewStateFactory(kymaRef klog.ObjectRef, kcpCluster composed.StateCluster, skrCluster composed.StateCluster,
	fileRestoreClientProvider gcpclient.ClientProvider[client.FileRestoreClient], env abstractions.Environment) StateFactory {

	return &stateFactory{
		kymaRef:                   kymaRef,
		kcpCluster:                kcpCluster,
		skrCluster:                skrCluster,
		fileRestoreClientProvider: fileRestoreClientProvider,
		env:                       env,
	}
}

type stateFactory struct {
	kymaRef                   klog.ObjectRef
	kcpCluster                composed.StateCluster
	skrCluster                composed.StateCluster
	fileRestoreClientProvider gcpclient.ClientProvider[client.FileRestoreClient]
	env                       abstractions.Environment
}

func (f *stateFactory) NewState(ctx context.Context, baseState composed.State) (*State, error) {
	fbc, err := f.fileRestoreClientProvider(
		ctx,
		f.env.Get("GCP_SA_JSON_KEY_PATH"),
	)
	if err != nil {
		return nil, err
	}
	return &State{
		State:             baseState,
		KymaRef:           f.kymaRef,
		KcpCluster:        f.kcpCluster,
		SkrCluster:        f.skrCluster,
		fileRestoreClient: fbc,
	}, nil
}

func (s *State) ObjAsGcpNfsVolumeRestore() *cloudresourcesv1beta1.GcpNfsVolumeRestore {
	return s.Obj().(*cloudresourcesv1beta1.GcpNfsVolumeRestore)
}
