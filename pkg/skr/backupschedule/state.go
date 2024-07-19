package backupschedule

import (
	"context"
	"github.com/gorhill/cronexpr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
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

	SourceRef composed.ObjWithConditions

	cronExpression *cronexpr.Expression
	nextRunTime    time.Time

	Scope *cloudcontrolv1beta1.Scope
	env   abstractions.Environment
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
	}, nil
}

func (s *State) ObjAsBackupSchedule() BackupSchedule {
	return s.Obj().(BackupSchedule)
}
