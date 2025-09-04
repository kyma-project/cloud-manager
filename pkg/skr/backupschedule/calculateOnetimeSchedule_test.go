package backupschedule

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type calculateOnetimeScheduleSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *calculateOnetimeScheduleSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *calculateOnetimeScheduleSuite) TestWhenNfsScheduleIsDeleting() {

	obj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ctx := calculateOnetimeSchedule(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *calculateOnetimeScheduleSuite) TestRecurringSchedule() {

	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.Schedule = "* * * * *"
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ctx := calculateOnetimeSchedule(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *calculateOnetimeScheduleSuite) TestAlreadySetSchedule() {

	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the schedule
	obj.Status.NextRunTimes = []string{time.Now().Format(time.RFC3339)}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	//Invoke API under test
	err, _ctx := calculateOnetimeSchedule(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *calculateOnetimeScheduleSuite) TestScheduleWithStartTime() {

	obj := gcpNfsBackupSchedule.DeepCopy()
	startTime := time.Now().Add(time.Hour)
	obj.Spec.StartTime = &metav1.Time{Time: startTime}
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = calculateOnetimeSchedule(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeue, err)

	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsBackupSchedule.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	s.Nil(err)
	s.Equal(1, len(fromK8s.Status.NextRunTimes))
	s.Equal(startTime.UTC().Format(time.RFC3339), fromK8s.Status.NextRunTimes[0])
}

func (s *calculateOnetimeScheduleSuite) TestScheduleWithNoStartTime() {

	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = calculateOnetimeSchedule(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeue, err)

	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsBackupSchedule.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	s.Nil(err)
	s.Equal(1, len(fromK8s.Status.NextRunTimes))
	runTime, err := time.Parse(time.RFC3339, fromK8s.Status.NextRunTimes[0])
	s.Nil(err)
	s.GreaterOrEqual(time.Second*1, time.Since(runTime))
}

func (s *calculateOnetimeScheduleSuite) TestScheduleWithLastCreateRun() {

	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the schedule
	lastCreateRun := &metav1.Time{Time: time.Now()}
	obj.Status.LastCreateRun = lastCreateRun
	obj.Status.BackupCount = 1
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = calculateOnetimeSchedule(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeue, err)

	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsBackupSchedule.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	s.Nil(err)
	s.Equal(1, len(fromK8s.Status.NextRunTimes))
	expectedNextRun := time.Date(lastCreateRun.Year(), lastCreateRun.Month(), lastCreateRun.Day()+1,
		lastCreateRun.Hour()+1, 0, 0, 0, time.UTC)
	runTime, err := time.Parse(time.RFC3339, fromK8s.Status.NextRunTimes[0])
	s.Nil(err)
	s.Equal(expectedNextRun, runTime)
}

func TestCalculateOnetimeScheduleSuite(t *testing.T) {
	suite.Run(t, new(calculateOnetimeScheduleSuite))
}
