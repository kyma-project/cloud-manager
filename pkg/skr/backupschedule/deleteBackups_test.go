package backupschedule

import (
	"context"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	suite.Equal(runTime.Unix(), fromK8s.Status.LastDeleteRun.Unix())
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
	suite.Equal(runTime.Unix(), fromK8s.Status.LastDeleteRun.Unix())
	suite.Nil(fromK8s.Status.NextRunTimes)
	suite.Nil(fromK8s.Status.LastDeletedBackups)
}

func (suite *deleteBackupsSuite) testDeleteBackup(backup1, backup2, backup client.Object, maxDays, maxReady, maxFailed int, b1Exists, b2Exists bool) {

	runTime := time.Now().UTC()

	obj := gcpNfsBackupSchedule.DeepCopy()
	obj.Spec.MaxRetentionDays = maxDays
	obj.Spec.MaxReadyBackups = maxReady
	obj.Spec.MaxFailedBackups = maxFailed
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
	err = factory.skrCluster.K8sClient().Create(ctx, backup1)
	suite.Nil(err)
	state.Backups = append(state.Backups, backup1)

	if backup2 != nil {
		err = factory.skrCluster.K8sClient().Create(ctx, backup2)
		suite.Nil(err)
		state.Backups = append(state.Backups, backup2)
	}

	//Invoke API under test
	err, _ = deleteBackups(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeue, err)

	//Validate schedule data
	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.NotNil(fromK8s.Status.LastDeleteRun)
	suite.Equal(runTime.Unix(), fromK8s.Status.LastDeleteRun.Unix())

	//Validate backup1
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: backup1.GetName(),
			Namespace: obj.Namespace},
		backup)
	if b1Exists {
		suite.Nil(err)
	} else {
		suite.True(apierrors.IsNotFound(err))
	}

	//Validate backup2
	if backup2 != nil {
		err = factory.skrCluster.K8sClient().Get(ctx,
			types.NamespacedName{Name: backup2.GetName(),
				Namespace: obj.Namespace},
			backup)
		if b2Exists {
			suite.Nil(err)
		} else {
			suite.True(apierrors.IsNotFound(err))
		}
	}
}

func (suite *deleteBackupsSuite) TestMaxRetentionDays() {
	backup1 := gcpBackup1.DeepCopy()
	backup2 := gcpBackup2.DeepCopy()
	backup := &v1beta1.GcpNfsVolumeBackup{}
	suite.testDeleteBackup(backup1, backup2, backup, 1, 100, 5, true, false)
}

func (suite *deleteBackupsSuite) TestMaxReadyBackups() {
	backup1 := gcpBackup1.DeepCopy()
	backup1.Status.State = v1beta1.StateReady
	backup2 := gcpBackup2.DeepCopy()
	backup2.Status.State = v1beta1.StateReady
	backup := &v1beta1.GcpNfsVolumeBackup{}
	suite.testDeleteBackup(backup1, backup2, backup, 375, 1, 5, true, false)
}

func (suite *deleteBackupsSuite) TestMaxFailedBackups() {
	backup1 := gcpBackup1.DeepCopy()
	backup1.Status.State = v1beta1.StateFailed
	backup2 := gcpBackup2.DeepCopy()
	backup2.Status.State = v1beta1.StateFailed
	backup := &v1beta1.GcpNfsVolumeBackup{}
	suite.testDeleteBackup(backup1, backup2, backup, 375, 100, 1, true, false)
}

func (suite *deleteBackupsSuite) TestDeleteGcpBackupFailure() {
	backup1 := gcpBackup1.DeepCopy()
	backup := &v1beta1.GcpNfsVolumeBackup{}
	suite.testDeleteBackup(backup1, nil, backup, 1, 100, 5, true, false)
}

func TestDeleteBackupsSuite(t *testing.T) {
	suite.Run(t, new(deleteBackupsSuite))
}
