package sapnfsvolumesnapshot

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/snapshots"
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

	Scope        *cloudcontrolv1beta1.Scope
	SapNfsVolume *cloudresourcesv1beta1.SapNfsVolume

	snapshot *snapshots.Snapshot

	snapshotClient sapclient.SnapshotClient
	provider       sapclient.SapClientProvider[sapclient.SnapshotClient]
}

func (s *State) ObjAsSapNfsVolumeSnapshot() *cloudresourcesv1beta1.SapNfsVolumeSnapshot {
	return s.Obj().(*cloudresourcesv1beta1.SapNfsVolumeSnapshot)
}

func (s *State) OpenStackSnapshotName() string {
	return fmt.Sprintf("cm-%s", s.ObjAsSapNfsVolumeSnapshot().Status.Id)
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

func NewReconcilerFactory() *ReconcilerFactory {
	return &ReconcilerFactory{}
}

type ReconcilerFactory struct{}

func (f *ReconcilerFactory) New(
	scopeProvider scopeprovider.ScopeProvider,
	kcpCluster composed.StateCluster,
	skrCluster composed.StateCluster,
	provider sapclient.SapClientProvider[sapclient.SnapshotClient],
) *Reconciler {
	composedStateFactory := composed.NewStateFactory(skrCluster)
	stateFactory := NewStateFactory(scopeProvider, kcpCluster, skrCluster, provider)
	return &Reconciler{
		composedStateFactory: composedStateFactory,
		stateFactory:         stateFactory,
	}
}
