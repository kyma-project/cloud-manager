package backupschedule

import (
	"context"
	"github.com/gorhill/cronexpr"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
	"time"
)

type calculateRecurringScheduleSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *calculateRecurringScheduleSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *calculateRecurringScheduleSuite) TestWhenNfsScheduleIsDeleting() {

	obj := deletingGcpBackupSchedule.DeepCopy()
	obj.Spec.Schedule = "* * * * *"
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ctx := calculateRecurringSchedule(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *calculateRecurringScheduleSuite) TestOnetimeSchedule() {

	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.Schedule = ""
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ctx := calculateRecurringSchedule(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *calculateRecurringScheduleSuite) TestAlreadySetSchedule() {

	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.Schedule = "* * * * *"
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the schedule
	now := time.Now()
	obj.Status.Schedule = obj.Spec.Schedule
	obj.Status.NextRunTimes = []string{now.Format(time.RFC3339)}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	//Set cron expression in state
	expr, err := cronexpr.Parse(obj.Spec.Schedule)
	s.Nil(err)
	state.cronExpression = expr

	//Invoke API under test
	err, _ctx := calculateRecurringSchedule(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *calculateRecurringScheduleSuite) testSchedule(schedule string, start, end *time.Time, expectedState string, expectedRunTimes ...time.Time) {

	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.Schedule = schedule
	if start != nil {
		obj.Spec.StartTime = &metav1.Time{Time: *start}
	}
	if end != nil {
		obj.Spec.EndTime = &metav1.Time{Time: *end}
	}
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the Status.schedule
	obj.Status.Schedule = obj.Spec.Schedule
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	//Set cron expression in state
	expr, err := cronexpr.Parse(obj.Spec.Schedule)
	s.Nil(err)
	state.cronExpression = expr

	//Invoke API under test
	err, _ = calculateRecurringSchedule(ctx, state)

	//validate expected return values
	s.NotNil(err)

	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsBackupSchedule.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	s.Nil(err)
	s.Equal(len(expectedRunTimes), len(fromK8s.Status.NextRunTimes))

	for i, expected := range expectedRunTimes {
		actual, err := time.Parse(time.RFC3339, fromK8s.Status.NextRunTimes[i])
		s.Nil(err)
		s.WithinDuration(expected, actual, time.Minute*1)
	}
}

func (s *calculateRecurringScheduleSuite) TestForEveryMinute() {
	now := time.Now()
	schedule := "* * * * *"
	expectedState := v1beta1.JobStateActive
	expectedTimes := []time.Time{
		time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location()).UTC(),
		time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+2, 0, 0, now.Location()).UTC(),
		time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+3, 0, 0, now.Location()).UTC(),
	}
	s.testSchedule(schedule, nil, nil, expectedState, expectedTimes...)
}

func (s *calculateRecurringScheduleSuite) TestForEveryMinuteWithStartTime() {
	now := time.Now()
	schedule := "* * * * *"
	start := now.Add(time.Hour * 24 * 5)
	expectedState := v1beta1.JobStateActive
	expectedTimes := []time.Time{
		time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute()+1, 0, 0, now.Location()).UTC(),
		time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute()+2, 0, 0, now.Location()).UTC(),
		time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute()+3, 0, 0, now.Location()).UTC(),
	}
	s.testSchedule(schedule, &start, nil, expectedState, expectedTimes...)
}

func (s *calculateRecurringScheduleSuite) TestForEveryMinuteWithEndTime() {
	now := time.Now().UTC()
	schedule := "* * * * *"
	end := now.AddDate(1, 0, 0)

	expectedState := v1beta1.JobStateActive
	expectedTimes := []time.Time{
		time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location()).UTC(),
		time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+2, 0, 0, now.Location()).UTC(),
		time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+3, 0, 0, now.Location()).UTC(),
	}

	s.testSchedule(schedule, nil, &end, expectedState, expectedTimes...)
}

func (s *calculateRecurringScheduleSuite) TestForEveryMonth() {
	now := time.Now().UTC()
	schedule := "0 0 1 * *"
	expectedState := v1beta1.JobStateActive
	expectedTimes := []time.Time{
		time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).UTC(),
		time.Date(now.Year(), now.Month()+2, 1, 0, 0, 0, 0, now.Location()).UTC(),
		time.Date(now.Year(), now.Month()+3, 1, 0, 0, 0, 0, now.Location()).UTC(),
	}
	s.testSchedule(schedule, nil, nil, expectedState, expectedTimes...)
}

func (s *calculateRecurringScheduleSuite) TestForEveryMonthWithStartTime() {
	now := time.Now().UTC()
	schedule := "0 0 1 * *"
	start := now.Add(time.Hour * 24)
	expectedState := v1beta1.JobStateActive
	expectedTimes := []time.Time{
		time.Date(start.Year(), start.Month()+1, 1, 0, 0, 0, 0, now.Location()).UTC(),
		time.Date(start.Year(), start.Month()+2, 1, 0, 0, 0, 0, now.Location()).UTC(),
		time.Date(start.Year(), start.Month()+3, 1, 0, 0, 0, 0, now.Location()).UTC(),
	}
	s.testSchedule(schedule, &start, nil, expectedState, expectedTimes...)
}

func (s *calculateRecurringScheduleSuite) TestForEveryMonthWithEndTime() {
	now := time.Now().UTC()
	schedule := "0 0 1 * *"
	start := now.Add(time.Hour * 24)
	end := now.AddDate(1, 0, 0)

	expectedState := v1beta1.JobStateActive
	expectedTimes := []time.Time{
		time.Date(start.Year(), start.Month()+1, 1, 0, 0, 0, 0, now.Location()).UTC(),
		time.Date(start.Year(), start.Month()+2, 1, 0, 0, 0, 0, now.Location()).UTC(),
		time.Date(start.Year(), start.Month()+3, 1, 0, 0, 0, 0, now.Location()).UTC(),
	}

	s.testSchedule(schedule, &start, &end, expectedState, expectedTimes...)
}

func TestCalculateRecurringScheduleSuite(t *testing.T) {
	suite.Run(t, new(calculateRecurringScheduleSuite))
}
