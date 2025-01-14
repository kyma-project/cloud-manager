package backupschedule

import (
	"context"
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

func (suite *calculateRecurringScheduleSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *calculateRecurringScheduleSuite) TestWhenNfsScheduleIsDeleting() {

	obj := deletingGcpBackupSchedule.DeepCopy()
	obj.Spec.Schedule = "* * * * *"
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := calculateRecurringSchedule(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *calculateRecurringScheduleSuite) TestOnetimeSchedule() {

	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.Schedule = ""
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := calculateRecurringSchedule(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *calculateRecurringScheduleSuite) TestAlreadySetSchedule() {

	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.Schedule = "* * * * *"
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the schedule
	now := time.Now()
	obj.Status.Schedule = obj.Spec.Schedule
	obj.Status.NextRunTimes = []string{now.Format(time.RFC3339)}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := calculateRecurringSchedule(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *calculateRecurringScheduleSuite) testSchedule(schedule string, start, end *time.Time, expectedState string, expectedRunTimes ...time.Time) {

	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.Schedule = schedule
	if start != nil {
		obj.Spec.StartTime = &metav1.Time{Time: *start}
	}
	if end != nil {
		obj.Spec.EndTime = &metav1.Time{Time: *end}
	}
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the Status.schedule
	obj.Status.Schedule = obj.Spec.Schedule
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ = calculateRecurringSchedule(ctx, state)

	//validate expected return values
	suite.NotNil(err)

	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsBackupSchedule.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.Equal(len(expectedRunTimes), len(fromK8s.Status.NextRunTimes))

	for i, t := range expectedRunTimes {
		suite.Equal(t.Format(time.RFC3339), fromK8s.Status.NextRunTimes[i])
	}
}

func (suite *calculateRecurringScheduleSuite) TestForEveryMinute() {
	now := time.Now()
	schedule := "* * * * *"
	expectedState := v1beta1.JobStateActive
	expectedTimes := []time.Time{
		time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location()).UTC(),
		time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+2, 0, 0, now.Location()).UTC(),
		time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+3, 0, 0, now.Location()).UTC(),
	}
	suite.testSchedule(schedule, nil, nil, expectedState, expectedTimes...)
}

func (suite *calculateRecurringScheduleSuite) TestForEveryMinuteWithStartTime() {
	now := time.Now()
	schedule := "* * * * *"
	start := now.Add(time.Hour * 24 * 5)
	expectedState := v1beta1.JobStateActive
	expectedTimes := []time.Time{
		time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute()+1, 0, 0, now.Location()).UTC(),
		time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute()+2, 0, 0, now.Location()).UTC(),
		time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute()+3, 0, 0, now.Location()).UTC(),
	}
	suite.testSchedule(schedule, &start, nil, expectedState, expectedTimes...)
}

func (suite *calculateRecurringScheduleSuite) TestForEveryMinuteWithEndTime() {
	now := time.Now().UTC()
	schedule := "* * * * *"
	end := now.AddDate(1, 0, 0)

	expectedState := v1beta1.JobStateActive
	expectedTimes := []time.Time{
		time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location()).UTC(),
		time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+2, 0, 0, now.Location()).UTC(),
		time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+3, 0, 0, now.Location()).UTC(),
	}

	suite.testSchedule(schedule, nil, &end, expectedState, expectedTimes...)
}

func (suite *calculateRecurringScheduleSuite) TestForEveryMonth() {
	now := time.Now().UTC()
	schedule := "0 0 1 * *"
	expectedState := v1beta1.JobStateActive
	expectedTimes := []time.Time{
		time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).UTC(),
		time.Date(now.Year(), now.Month()+2, 1, 0, 0, 0, 0, now.Location()).UTC(),
		time.Date(now.Year(), now.Month()+3, 1, 0, 0, 0, 0, now.Location()).UTC(),
	}
	suite.testSchedule(schedule, nil, nil, expectedState, expectedTimes...)
}

func (suite *calculateRecurringScheduleSuite) TestForEveryMonthWithStartTime() {
	now := time.Now().UTC()
	schedule := "0 0 1 * *"
	start := now.Add(time.Hour * 24)
	expectedState := v1beta1.JobStateActive
	expectedTimes := []time.Time{
		time.Date(start.Year(), start.Month()+1, 1, 0, 0, 0, 0, now.Location()).UTC(),
		time.Date(start.Year(), start.Month()+2, 1, 0, 0, 0, 0, now.Location()).UTC(),
		time.Date(start.Year(), start.Month()+3, 1, 0, 0, 0, 0, now.Location()).UTC(),
	}
	suite.testSchedule(schedule, &start, nil, expectedState, expectedTimes...)
}

func (suite *calculateRecurringScheduleSuite) TestForEveryMonthWithEndTime() {
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

	suite.testSchedule(schedule, &start, &end, expectedState, expectedTimes...)
}

func TestCalculateRecurringScheduleSuite(t *testing.T) {
	suite.Run(t, new(calculateRecurringScheduleSuite))
}
