package gcpnfsbackupschedule

import (
	"context"
	"time"

	"github.com/gorhill/cronexpr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/backupschedule"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	SkrCluster composed.StateCluster

	// Common scheduling
	Scheduler      *backupschedule.ScheduleCalculator
	cronExpression *cronexpr.Expression
	nextRunTime    time.Time
	createRunDone  bool
	deleteRunDone  bool

	// GCP-specific (concrete types, no interfaces)
	Scope   *cloudcontrolv1beta1.Scope
	Source  *cloudresourcesv1beta1.GcpNfsVolume
	Backups []*cloudresourcesv1beta1.GcpNfsVolumeBackup

	env abstractions.Environment
}

// Implement backupschedule.ScheduleState interface

func (s *State) ObjAsBackupSchedule() backupschedule.BackupSchedule {
	return s.Obj().(backupschedule.BackupSchedule)
}

func (s *State) GetScheduleCalculator() *backupschedule.ScheduleCalculator {
	return s.Scheduler
}

func (s *State) GetCronExpression() *cronexpr.Expression  { return s.cronExpression }
func (s *State) SetCronExpression(e *cronexpr.Expression) { s.cronExpression = e }
func (s *State) GetNextRunTime() time.Time                { return s.nextRunTime }
func (s *State) SetNextRunTime(t time.Time)               { s.nextRunTime = t }
func (s *State) IsCreateRunCompleted() bool               { return s.createRunDone }
func (s *State) SetCreateRunCompleted(v bool)             { s.createRunDone = v }
func (s *State) IsDeleteRunCompleted() bool               { return s.deleteRunDone }
func (s *State) SetDeleteRunCompleted(v bool)             { s.deleteRunDone = v }

// GCP-specific getter
func (s *State) ObjAsGcpNfsBackupSchedule() *cloudresourcesv1beta1.GcpNfsBackupSchedule {
	return s.Obj().(*cloudresourcesv1beta1.GcpNfsBackupSchedule)
}

type StateFactory interface {
	NewState(ctx context.Context, baseState composed.State) (*State, error)
}

func NewStateFactory(kymaRef klog.ObjectRef, kcpCluster composed.StateCluster,
	skrCluster composed.StateCluster, env abstractions.Environment, clk clock.Clock) StateFactory {
	return &stateFactory{
		kymaRef:    kymaRef,
		kcpCluster: kcpCluster,
		skrCluster: skrCluster,
		env:        env,
		clk:        clk,
	}
}

type stateFactory struct {
	kymaRef    klog.ObjectRef
	kcpCluster composed.StateCluster
	skrCluster composed.StateCluster
	env        abstractions.Environment
	clk        clock.Clock
}

func (f *stateFactory) NewState(ctx context.Context, baseState composed.State) (*State, error) {
	return &State{
		State:      baseState,
		KymaRef:    f.kymaRef,
		KcpCluster: f.kcpCluster,
		SkrCluster: f.skrCluster,
		env:        f.env,
		Scheduler:  backupschedule.NewScheduleCalculator(f.clk, 1*time.Second),
	}, nil
}
