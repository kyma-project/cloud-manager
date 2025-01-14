package backupschedule

import (
	"context"
	"github.com/adhocore/gronx"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	SkrCluster composed.StateCluster

	SourceRef composed.ObjWithConditions
	Backups   []client.Object

	gronx              *gronx.Gronx
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
}

func (f *stateFactory) NewState(ctx context.Context, baseState composed.State) (*State, error) {

	return &State{
		State:      baseState,
		KymaRef:    f.kymaRef,
		KcpCluster: f.kcpCluster,
		SkrCluster: f.skrCluster,
		env:        f.env,
		gronx:      gronx.New(),
	}, nil
}

func (s *State) ObjAsBackupSchedule() BackupSchedule {
	return s.Obj().(BackupSchedule)
}

func (s *State) NextN(fromTime time.Time, n uint) ([]time.Time, error) {
	var times []time.Time
	if n <= 0 {
		return times, nil
	}

	expression := s.ObjAsBackupSchedule().GetActiveSchedule()
	if fromTime.IsZero() {
		fromTime = time.Now().UTC()
	}

	fromTime, err := gronx.NextTickAfter(expression, fromTime, true)
	if err != nil {
		return times, nil
	}
	times = append(times, fromTime)
	for range n - 1 {
		fromTime, err = gronx.NextTickAfter(expression, fromTime, false)
		if err != nil {
			return times, err
		}
		times = append(times, fromTime)
	}
	return times, nil
}
