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

type deleteBackupsSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *deleteBackupsSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *deleteBackupsSuite) TestWhenNfsScheduleIsDeleting() {
	obj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := deleteBackups(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *deleteBackupsSuite) TestWhenRunNotDue() {
	runTime := time.Now().Add(time.Minute).UTC()
	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	state.nextRunTime = runTime
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := deleteBackups(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *deleteBackupsSuite) TestWhenRunAlreadyCompleted() {

	runTime := time.Now().UTC()
	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Update the creation run time to the next run time
	obj.Status.LastDeleteRun = &metav1.Time{Time: runTime.UTC()}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	state.nextRunTime = runTime
	state.deleteRunCompleted = true
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := deleteBackups(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *deleteBackupsSuite) TestWhenNoBackupsExist() {

	runTime := time.Now().UTC()
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.MaxRetentionDays = 0
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	state.nextRunTime = runTime
	suite.Nil(err)

	//Invoke API under test
	err, _ = deleteBackups(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeue, err)

	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.NotNil(fromK8s.Status.LastDeleteRun)
	suite.Equal(runTime.Unix(), fromK8s.Status.LastDeleteRun.Time.Unix())
	suite.Nil(fromK8s.Status.NextRunTimes)
	suite.Nil(fromK8s.Status.LastDeletedBackups)
}

func (suite *deleteBackupsSuite) TestWhenNoMaxRetentionSet() {

	runTime := time.Now().UTC()
	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.MaxRetentionDays = 0
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	state.nextRunTime = runTime
	suite.Nil(err)

	//Create provider specific backup objects
	err = factory.skrCluster.K8sClient().Create(ctx, gcpBackup1.DeepCopy())
	suite.Nil(err)
	state.Backups = append(state.Backups, gcpBackup1)

	//Invoke API under test
	err, _ = deleteBackups(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeue, err)

	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.NotNil(fromK8s.Status.LastDeleteRun)
	suite.Equal(runTime.Unix(), fromK8s.Status.LastDeleteRun.Time.Unix())
	suite.Nil(fromK8s.Status.NextRunTimes)
	suite.Nil(fromK8s.Status.LastDeletedBackups)
}

func (suite *deleteBackupsSuite) testDeleteBackup(createBackup2 bool) {

	runTime := time.Now().UTC()

	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.MaxRetentionDays = 1
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	state.Scope = &gcpScope
	state.nextRunTime = runTime

	//Create provider specific backup objects
	err = factory.skrCluster.K8sClient().Create(ctx, gcpBackup1.DeepCopy())
	suite.Nil(err)

	if createBackup2 {
		err = factory.skrCluster.K8sClient().Create(ctx, gcpBackup2.DeepCopy())
		suite.Nil(err)
	}

	state.Backups = append(state.Backups, gcpBackup1)
	state.Backups = append(state.Backups, gcpBackup2)

	//Invoke API under test
	err, _ = deleteBackups(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeue, err)

	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.NotNil(fromK8s.Status.LastDeleteRun)
	suite.Equal(runTime.Unix(), fromK8s.Status.LastDeleteRun.Time.Unix())

	backup := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpBackup1.GetName(),
			Namespace: gcpNfsBackupSchedule.Namespace},
		backup)
	suite.Nil(err)
}

func (suite *deleteBackupsSuite) TestDeleteGcpBackup() {
	suite.testDeleteBackup(true)
}

func (suite *deleteBackupsSuite) TestDeleteGcpBackupFailure() {
	suite.testDeleteBackup(false)
}

func TestDeleteBackupsSuite(t *testing.T) {
	suite.Run(t, new(deleteBackupsSuite))
}
