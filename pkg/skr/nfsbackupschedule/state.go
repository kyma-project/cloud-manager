package nfsbackupschedule

import (
	"context"
	"github.com/gorhill/cronexpr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"k8s.io/klog/v2"
	"time"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	SkrCluster composed.StateCluster

	NfsVolume composed.ObjWithConditions

	cronExpression *cronexpr.Expression
	nextRunTime    time.Time

	Scope     *cloudcontrolv1beta1.Scope
	env       abstractions.Environment
	gcpConfig *client.GcpConfig
}

type StateFactory interface {
	NewState(ctx context.Context, baseState composed.State) (*State, error)
}

func NewStateFactory(kymaRef klog.ObjectRef, kcpCluster composed.StateCluster, skrCluster composed.StateCluster,
	env abstractions.Environment) StateFactory {

	return &stateFactory{
		kymaRef:    kymaRef,
		kcpCluster: kcpCluster,
		skrCluster: skrCluster,
		env:        env,
		gcpConfig:  client.GetGcpConfig(env),
	}
}

type stateFactory struct {
	kymaRef    klog.ObjectRef
	kcpCluster composed.StateCluster
	skrCluster composed.StateCluster
	env        abstractions.Environment
	gcpConfig  *client.GcpConfig
}

func (f *stateFactory) NewState(ctx context.Context, baseState composed.State) (*State, error) {

	return &State{
		State:      baseState,
		KymaRef:    f.kymaRef,
		KcpCluster: f.kcpCluster,
		SkrCluster: f.skrCluster,
		env:        f.env,
		gcpConfig:  f.gcpConfig,
	}, nil
}

func (s *State) ObjAsNfsBackupSchedule() *cloudresourcesv1beta1.NfsBackupSchedule {
	return s.Obj().(*cloudresourcesv1beta1.NfsBackupSchedule)
}
