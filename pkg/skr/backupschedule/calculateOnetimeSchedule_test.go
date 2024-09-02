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

type calculateOnetimeScheduleSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *calculateOnetimeScheduleSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *calculateOnetimeScheduleSuite) TestWhenNfsScheduleIsDeleting() {

	obj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := calculateOnetimeSchedule(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *calculateOnetimeScheduleSuite) TestRecurringSchedule() {

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
	err, _ctx := calculateOnetimeSchedule(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *calculateOnetimeScheduleSuite) TestAlreadySetSchedule() {

	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the schedule
	obj.Status.NextRunTimes = []string{time.Now().Format(time.RFC3339)}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := calculateOnetimeSchedule(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *calculateOnetimeScheduleSuite) TestScheduleWithStartTime() {

	obj := gcpNfsBackupSchedule.DeepCopy()
	startTime := time.Now().Add(time.Hour)
	obj.Spec.StartTime = &metav1.Time{Time: startTime}
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ = calculateOnetimeSchedule(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeue, err)

	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsBackupSchedule.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.Equal(1, len(fromK8s.Status.NextRunTimes))
	suite.Equal(startTime.UTC().Format(time.RFC3339), fromK8s.Status.NextRunTimes[0])
}

func (suite *calculateOnetimeScheduleSuite) TestScheduleWithNoStartTime() {

	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ = calculateOnetimeSchedule(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeue, err)

	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsBackupSchedule.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.Equal(1, len(fromK8s.Status.NextRunTimes))
	runTime, err := time.Parse(time.RFC3339, fromK8s.Status.NextRunTimes[0])
	suite.Nil(err)
	suite.GreaterOrEqual(time.Second*1, time.Since(runTime))
}

func TestCalculateOnetimeScheduleSuite(t *testing.T) {
	suite.Run(t, new(calculateOnetimeScheduleSuite))
}
