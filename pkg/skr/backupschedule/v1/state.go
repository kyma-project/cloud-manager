package v1

import (
	"context"
	"time"

	"github.com/gorhill/cronexpr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/backupschedule"
	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	SkrCluster composed.StateCluster

	SourceRef composed.ObjWithConditions
	Backups   []client.Object

	cronExpression     *cronexpr.Expression
	nextRunTime        time.Time
	createRunCompleted bool
	deleteRunCompleted bool

	backupImpl backupImpl
	Scope      *cloudcontrolv1beta1.Scope
	env        abstractions.Environment
}

type StateFactory interface {
	NewState(ctx context.Context, baseState composed.State) (*State, error)
}

func NewStateFactory(scopeProvider scopeprovider.ScopeProvider, kcpCluster composed.StateCluster, skrCluster composed.StateCluster,
	env abstractions.Environment) StateFactory {

	return &stateFactory{
		scopeProvider: scopeProvider,
		kcpCluster:    kcpCluster,
		skrCluster:    skrCluster,
		env:           env,
	}
}

type stateFactory struct {
	scopeProvider scopeprovider.ScopeProvider
	kcpCluster    composed.StateCluster
	skrCluster    composed.StateCluster
	env           abstractions.Environment
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
		env:        f.env,
	}, nil
}

func (s *State) ObjAsBackupSchedule() backupschedule.BackupSchedule {
	return s.Obj().(backupschedule.BackupSchedule)
}

func (s *State) ObjAsGcpNfsBackupSchedule() *cloudresourcesv1beta1.GcpNfsBackupSchedule {
	return s.Obj().(*cloudresourcesv1beta1.GcpNfsBackupSchedule)
}
