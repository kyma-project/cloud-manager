package backupschedule

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type checkCompletedSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *checkCompletedSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *checkCompletedSuite) TestWhenNfsScheduleIsDeleting() {
	obj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ctx := checkCompleted(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *checkCompletedSuite) TestWhenScheduleIsNotDone() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Status.State = v1beta1.JobStateProcessing
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ctx := checkCompleted(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *checkCompletedSuite) TestWhenAlreadyCompleted() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Status.State = v1beta1.JobStateDone
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = checkCompleted(ctx, state)

	//validate expected return values
	s.Equal(composed.StopAndForget, err)
	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsBackupSchedule.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), v1beta1.JobStateDone, fromK8s.Status.State)
}

func (s *checkCompletedSuite) TestValidCronExpression() {
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
	err, _ = checkCompleted(ctx, state)

	//validate expected return values
	s.Nil(err)
}

func (s *checkCompletedSuite) TestAfterEndTime() {
	now := time.Now().UTC()
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.EndTime = &metav1.Time{Time: now.Add(-2 * time.Hour)}
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
	err, _ = checkCompleted(ctx, state)

	//validate expected return values
	s.Equal(composed.StopAndForget, err)

	//Get the object from K8s
	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	s.Nil(err)
	s.Equal(v1beta1.JobStateDone, fromK8s.Status.State)
	s.Nil(fromK8s.Status.NextRunTimes)
}

func (s *checkCompletedSuite) TestCompletedOnetimeSchedule() {
	now := time.Now().UTC()
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.Schedule = ""

	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Update the lastCreateRun with current time
	obj.Status.LastCreateRun = &metav1.Time{Time: now}
	obj.Status.LastDeleteRun = &metav1.Time{Time: now}
	obj.Status.BackupCount = 0
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	//Invoke API under test
	err, _ = checkCompleted(ctx, state)

	//validate expected return values
	s.Equal(composed.StopAndForget, err)

	//Get the object from K8s
	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	s.Nil(err)
	s.Equal(v1beta1.JobStateDone, fromK8s.Status.State)
	s.Nil(fromK8s.Status.NextRunTimes)
}

func TestCheckCompletedSuite(t *testing.T) {
	suite.Run(t, new(checkCompletedSuite))
}
