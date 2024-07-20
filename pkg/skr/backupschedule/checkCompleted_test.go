package backupschedule

import (
	"context"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
	"time"
)

type checkCompletedSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *checkCompletedSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *checkCompletedSuite) TestWhenNfsScheduleIsDeleting() {
	obj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := checkCompleted(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *checkCompletedSuite) TestWhenScheduleIsNotDone() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Status.State = v1beta1.JobStateProcessing
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := checkCompleted(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *checkCompletedSuite) TestWhenAlreadyCompleted() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Status.State = v1beta1.JobStateDone
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ = checkCompleted(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopAndForget, err)
	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsBackupSchedule.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), v1beta1.JobStateDone, fromK8s.Status.State)
}

func (suite *checkCompletedSuite) TestValidCronExpression() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.Schedule = "* * * * *"
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ = checkCompleted(ctx, state)

	//validate expected return values
	suite.Nil(err)
}

func (suite *checkCompletedSuite) TestAfterEndTime() {
	now := time.Now().UTC()
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.EndTime = &metav1.Time{Time: now.Add(-2 * time.Hour)}
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
	err, _ = checkCompleted(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopAndForget, err)

	//Get the object from K8s
	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.Equal(v1beta1.JobStateDone, fromK8s.Status.State)
	suite.Nil(fromK8s.Status.NextRunTimes)
}

func (suite *checkCompletedSuite) TestCompletedOnetimeSchedule() {
	now := time.Now().UTC()
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.Schedule = ""

	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Update the lastCreateRun with current time
	obj.Status.LastCreateRun = &metav1.Time{Time: now}
	obj.Status.LastDeleteRun = &metav1.Time{Time: now}
	obj.Status.BackupCount = 0
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ = checkCompleted(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopAndForget, err)

	//Get the object from K8s
	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.Equal(v1beta1.JobStateDone, fromK8s.Status.State)
	suite.Nil(fromK8s.Status.NextRunTimes)
}

func TestCheckCompletedSuite(t *testing.T) {
	suite.Run(t, new(checkCompletedSuite))
}
