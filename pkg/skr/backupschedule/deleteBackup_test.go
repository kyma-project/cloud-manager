package backupschedule

import (
	"context"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"
)

type deleteBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *deleteBackupSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *deleteBackupSuite) TestWhenNfsScheduleIsDeleting() {
	obj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := deleteBackup(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *deleteBackupSuite) TestWhenRunNotDue() {
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
	err, _ctx := deleteBackup(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *deleteBackupSuite) TestWhenRunAlreadyCompleted() {

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
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := deleteBackup(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *deleteBackupSuite) TestWhenNoMaxRetentionSet() {

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
	err, _ = deleteBackup(ctx, state)

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

func (suite *deleteBackupSuite) testDeleteBackup(createBackup2 bool) {

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
	var backup1, backup2 client.Object
	backup1Meta := metav1.ObjectMeta{
		Name:              "test-backup-1",
		Namespace:         gcpNfsBackupSchedule.Namespace,
		CreationTimestamp: metav1.Time{Time: time.Now()},
		Labels: map[string]string{
			v1beta1.LabelScheduleName:      gcpNfsBackupSchedule.Name,
			v1beta1.LabelScheduleNamespace: gcpNfsBackupSchedule.Namespace,
		},
	}
	backup2Meta := metav1.ObjectMeta{
		Name:              "test-backup-2",
		Namespace:         gcpNfsBackupSchedule.Namespace,
		CreationTimestamp: metav1.Time{Time: time.Now().AddDate(0, 0, -2)},
		Labels: map[string]string{
			v1beta1.LabelScheduleName:      gcpNfsBackupSchedule.Name,
			v1beta1.LabelScheduleNamespace: gcpNfsBackupSchedule.Namespace,
		},
	}
	gcpSpec := v1beta1.GcpNfsVolumeBackupSpec{
		Source: v1beta1.GcpNfsVolumeBackupSource{
			Volume: v1beta1.GcpNfsVolumeRef{
				Name:      obj.Spec.NfsVolumeRef.Name,
				Namespace: obj.Spec.NfsVolumeRef.Namespace,
			},
		},
		Location: "us-west1-a",
	}
	backup1 = &v1beta1.GcpNfsVolumeBackup{
		ObjectMeta: backup1Meta,
		Spec:       gcpSpec,
	}
	backup2 = &v1beta1.GcpNfsVolumeBackup{
		ObjectMeta: backup2Meta,
		Spec:       gcpSpec,
	}
	err = factory.skrCluster.K8sClient().Create(ctx, backup1)
	suite.Nil(err)
	if createBackup2 {
		err = factory.skrCluster.K8sClient().Create(ctx, backup2)
		suite.Nil(err)
	}

	//Invoke API under test
	err, _ = deleteBackup(ctx, state)

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
		types.NamespacedName{Name: backup1.GetName(),
			Namespace: gcpNfsBackupSchedule.Namespace},
		backup)
	suite.Nil(err)
}

func (suite *deleteBackupSuite) TestDeleteGcpBackup() {
	suite.testDeleteBackup(true)
}

func (suite *deleteBackupSuite) TestDeleteGcpBackupFailure() {
	suite.testDeleteBackup(false)
}

func TestDeleteNfsBackupSuite(t *testing.T) {
	suite.Run(t, new(deleteBackupSuite))
}
