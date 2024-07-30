package backupschedule

import (
	"context"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
	"time"
)

type evaluateNextRunSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *evaluateNextRunSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *evaluateNextRunSuite) TestWhenNfsScheduleIsDeleting() {
	obj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := evaluateNextRun(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *evaluateNextRunSuite) TestWhenNextRunTimesIsNotSet() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ = evaluateNextRun(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopAndForget, err)

	//Get the object from K8s
	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.Equal(v1beta1.JobStateError, fromK8s.Status.State)
	suite.Equal(metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	suite.Equal(v1beta1.JobStateError, fromK8s.Status.Conditions[0].Type)
}

func (suite *evaluateNextRunSuite) TestWhenNextRunTimesIsNotParseable() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Update the next run time to an invalid time
	obj.Status.NextRunTimes = []string{"invalid-time"}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ = evaluateNextRun(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopAndForget, err)

	//Get the object from K8s
	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.Equal(v1beta1.JobStateError, fromK8s.Status.State)
	suite.Equal(metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	suite.Equal(v1beta1.JobStateError, fromK8s.Status.Conditions[0].Type)
}

func (suite *evaluateNextRunSuite) TestWhenNextRunTimeIsNotDueYet() {
	now := time.Now().UTC()
	offset := 1 * time.Hour
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.EndTime = &metav1.Time{Time: now.Add(5 * time.Hour)}
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Update the next run time with current time
	obj.Status.NextRunTimes = []string{now.Add(offset).Format(time.RFC3339)}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ = evaluateNextRun(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeueDelay(offset), err)
}

func (suite *evaluateNextRunSuite) TestWhenScheduleJustRun() {
	now := time.Now().UTC()
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.Schedule = "* * * * *"

	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Update the lastCreateRun with current time
	obj.Status.NextRunTimes = []string{now.Format(time.RFC3339)}
	obj.Status.LastCreateRun = &metav1.Time{Time: now}
	obj.Status.LastDeleteRun = &metav1.Time{Time: now}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ = evaluateNextRun(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeue, err)

	//Get the object from K8s
	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.Nil(fromK8s.Status.NextRunTimes)
}

func (suite *evaluateNextRunSuite) TestWhenNextRunTimesIsValid() {
	now := time.Now().UTC()
	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Update the next run time with current time
	obj.Status.NextRunTimes = []string{now.Format(time.RFC3339)}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := evaluateNextRun(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
	suite.Equal(now.Unix(), state.nextRunTime.Unix())
}

func TestEvaluateNextRunSuite(t *testing.T) {
	suite.Run(t, new(evaluateNextRunSuite))
}
