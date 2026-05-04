package sapnfsvolumesnapshotrestore

import (
	"context"

	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shares"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
	"k8s.io/klog/v2"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	SkrCluster composed.StateCluster

	Scope             *cloudcontrolv1beta1.Scope
	SourceSnapshot    *cloudresourcesv1beta1.SapNfsVolumeSnapshot
	DestinationVolume *cloudresourcesv1beta1.SapNfsVolume
	CreatedVolume     *cloudresourcesv1beta1.SapNfsVolume

	// shareId is the Manila share ID of the destination volume (for in-place restore)
	shareId string
	// share holds the current Manila share state (for polling during revert)
	share *shares.Share

	snapshotClient sapclient.SnapshotClient
	shareClient    sapclient.ShareClient
	provider       sapclient.SapClientProvider[sapclient.SnapshotClient]
}

func (s *State) ObjAsSapNfsVolumeSnapshotRestore() *cloudresourcesv1beta1.SapNfsVolumeSnapshotRestore {
	return s.Obj().(*cloudresourcesv1beta1.SapNfsVolumeSnapshotRestore)
}

type StateFactory interface {
	NewState(ctx context.Context, baseState composed.State) (*State, error)
}

func NewStateFactory(
	scopeProvider scopeprovider.ScopeProvider,
	kcpCluster composed.StateCluster,
	skrCluster composed.StateCluster,
	provider sapclient.SapClientProvider[sapclient.SnapshotClient],
) StateFactory {
	return &stateFactory{
		scopeProvider: scopeProvider,
		kcpCluster:    kcpCluster,
		skrCluster:    skrCluster,
		provider:      provider,
	}
}

type stateFactory struct {
	scopeProvider scopeprovider.ScopeProvider
	kcpCluster    composed.StateCluster
	skrCluster    composed.StateCluster
	provider      sapclient.SapClientProvider[sapclient.SnapshotClient]
}

func (f *stateFactory) NewState(ctx context.Context, baseState composed.State) (*State, error) {
	kymaRef, err := f.scopeProvider.GetScope(ctx, baseState.Name())
	if err != nil {
		return nil, err
	}
	return &State{
		State:      baseState,
		KymaRef:    kymaRef,
		KcpCluster: f.kcpCluster,
		SkrCluster: f.skrCluster,
		provider:   f.provider,
	}, nil
}
