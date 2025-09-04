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

type evaluateNextRunSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *evaluateNextRunSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *evaluateNextRunSuite) TestWhenNfsScheduleIsDeleting() {
	obj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ctx := evaluateNextRun(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *evaluateNextRunSuite) TestWhenNextRunTimesIsNotSet() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = evaluateNextRun(ctx, state)

	//validate expected return values
	s.Equal(composed.StopAndForget, err)

	//Get the object from K8s
	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	s.Nil(err)
	s.Equal(v1beta1.JobStateError, fromK8s.Status.State)
	s.Equal(metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	s.Equal(v1beta1.JobStateError, fromK8s.Status.Conditions[0].Type)
}

func (s *evaluateNextRunSuite) TestWhenNextRunTimesIsNotParseable() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Update the next run time to an invalid time
	obj.Status.NextRunTimes = []string{"invalid-time"}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = evaluateNextRun(ctx, state)

	//validate expected return values
	s.Equal(composed.StopAndForget, err)

	//Get the object from K8s
	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	s.Nil(err)
	s.Equal(v1beta1.JobStateError, fromK8s.Status.State)
	s.Equal(metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	s.Equal(v1beta1.JobStateError, fromK8s.Status.Conditions[0].Type)
}

func (s *evaluateNextRunSuite) TestWhenNextRunTimeIsNotDueYet() {
	now := time.Now().UTC()
	offset := 1 * time.Hour
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.EndTime = &metav1.Time{Time: now.Add(5 * time.Hour)}
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Update the next run time with current time
	obj.Status.NextRunTimes = []string{now.Add(offset).Format(time.RFC3339)}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = evaluateNextRun(ctx, state)

	//validate expected return values
	result, err := composed.HandleWithoutLogging(err, ctx)
	delay := result.RequeueAfter
	s.Nil(err)
	s.NotNil(delay)
	s.Greater(delay, time.Duration(int64(0.95*float64(offset))))
	s.LessOrEqual(delay, offset)
}

func (s *evaluateNextRunSuite) TestWhenScheduleJustRun() {
	now := time.Now().UTC()
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.Schedule = "* * * * *"

	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Update the lastCreateRun with current time
	obj.Status.NextRunTimes = []string{now.Format(time.RFC3339)}
	obj.Status.LastCreateRun = &metav1.Time{Time: now}
	obj.Status.LastDeleteRun = &metav1.Time{Time: now}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = evaluateNextRun(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeue, err)

	//Get the object from K8s
	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	s.Nil(err)
	s.Nil(fromK8s.Status.NextRunTimes)
}

func (s *evaluateNextRunSuite) TestWhenNextRunTimesIsValid() {
	now := time.Now().UTC()
	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Update the next run time with current time
	obj.Status.NextRunTimes = []string{now.Format(time.RFC3339)}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	//Invoke API under test
	err, _ctx := evaluateNextRun(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
	s.Equal(now.Unix(), state.nextRunTime.Unix())
}

func TestEvaluateNextRunSuite(t *testing.T) {
	suite.Run(t, new(evaluateNextRunSuite))
}
