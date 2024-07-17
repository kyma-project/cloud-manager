package nfsbackupschedule

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"
)

type deleteNfsBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *deleteNfsBackupSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *deleteNfsBackupSuite) TestWhenNfsScheduleIsDeleting() {
	obj := deletingBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with NfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := deleteNfsBackup(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *deleteNfsBackupSuite) TestWhenRunNotDue() {
	runTime := time.Now().Add(time.Minute).UTC()
	obj := nfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with NfsBackupSchedule
	state, err := factory.newStateWith(obj)
	state.nextRunTime = runTime
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := deleteNfsBackup(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *deleteNfsBackupSuite) TestWhenRunAlreadyCompleted() {

	runTime := time.Now().UTC()
	obj := nfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Update the creation run time to the next run time
	obj.Status.LastDeleteRun = &metav1.Time{Time: runTime.UTC()}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//Get state object with NfsBackupSchedule
	state, err := factory.newStateWith(obj)
	state.nextRunTime = runTime
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := deleteNfsBackup(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *deleteNfsBackupSuite) TestWhenNoMaxRetentionSet() {

	runTime := time.Now().UTC()
	obj := nfsBackupSchedule.DeepCopy()
	obj.Spec.MaxRetentionDays = 0
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with NfsBackupSchedule
	state, err := factory.newStateWith(obj)
	state.nextRunTime = runTime
	suite.Nil(err)

	//Invoke API under test
	err, _ = deleteNfsBackup(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeue, err)

	fromK8s := &v1beta1.NfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: nfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.NotNil(fromK8s.Status.LastDeleteRun)
	suite.Equal(runTime.Unix(), fromK8s.Status.LastDeleteRun.Time.Unix())
	suite.Nil(fromK8s.Status.NextRunTimes)
	suite.Nil(fromK8s.Status.LastDeletedBackups)
}

func (suite *deleteNfsBackupSuite) testDeleteBackup(scope *cloudcontrolv1beta1.Scope, createBackup2 bool) {

	runTime := time.Now().UTC()

	obj := nfsBackupSchedule.DeepCopy()
	obj.Spec.MaxRetentionDays = 1
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with NfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	state.Scope = scope
	state.nextRunTime = runTime

	//Create provider specific backup objects
	var backup1, backup2 client.Object
	backup1Meta := metav1.ObjectMeta{
		Name:              "test-backup-1",
		Namespace:         nfsBackupSchedule.Namespace,
		CreationTimestamp: metav1.Time{Time: time.Now()},
		Labels: map[string]string{
			v1beta1.LabelScheduleName:      nfsBackupSchedule.Name,
			v1beta1.LabelScheduleNamespace: nfsBackupSchedule.Namespace,
		},
	}
	backup2Meta := metav1.ObjectMeta{
		Name:              "test-backup-2",
		Namespace:         nfsBackupSchedule.Namespace,
		CreationTimestamp: metav1.Time{Time: time.Now().AddDate(0, 0, -2)},
		Labels: map[string]string{
			v1beta1.LabelScheduleName:      nfsBackupSchedule.Name,
			v1beta1.LabelScheduleNamespace: nfsBackupSchedule.Namespace,
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
	awsSpec := v1beta1.AwsNfsVolumeBackupSpec{
		Source: v1beta1.AwsNfsVolumeBackupSource{
			Volume: v1beta1.VolumeRef{
				Name:      obj.Spec.NfsVolumeRef.Name,
				Namespace: obj.Spec.NfsVolumeRef.Namespace,
			},
		},
	}
	switch scope.Spec.Provider {
	case cloudcontrolv1beta1.ProviderGCP:
		backup1 = &v1beta1.GcpNfsVolumeBackup{
			ObjectMeta: backup1Meta,
			Spec:       gcpSpec,
		}
		backup2 = &v1beta1.GcpNfsVolumeBackup{
			ObjectMeta: backup2Meta,
			Spec:       gcpSpec,
		}
	case cloudcontrolv1beta1.ProviderAws:
		backup1 = &v1beta1.AwsNfsVolumeBackup{
			ObjectMeta: backup1Meta,
			Spec:       awsSpec,
		}
		backup2 = &v1beta1.AwsNfsVolumeBackup{
			ObjectMeta: backup2Meta,
			Spec:       awsSpec,
		}
	default:
		suite.Fail("Invalid provider")
	}
	err = factory.skrCluster.K8sClient().Create(ctx, backup1)
	suite.Nil(err)
	if createBackup2 {
		err = factory.skrCluster.K8sClient().Create(ctx, backup2)
		suite.Nil(err)
	}

	//Invoke API under test
	err, _ = deleteNfsBackup(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeue, err)

	fromK8s := &v1beta1.NfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: nfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.NotNil(fromK8s.Status.LastDeleteRun)
	suite.Equal(runTime.Unix(), fromK8s.Status.LastDeleteRun.Time.Unix())

	var backup client.Object
	switch scope.Spec.Provider {
	case cloudcontrolv1beta1.ProviderGCP:
		backup = &v1beta1.GcpNfsVolumeBackup{}
	case cloudcontrolv1beta1.ProviderAws:
		backup = &v1beta1.AwsNfsVolumeBackup{}
	default:
		suite.Fail("Invalid provider")
	}

	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: backup1.GetName(),
			Namespace: nfsBackupSchedule.Namespace},
		backup)
	suite.Nil(err)
}

func (suite *deleteNfsBackupSuite) TestDeleteGcpBackup() {
	suite.testDeleteBackup(&gcpScope, true)
}

func (suite *deleteNfsBackupSuite) TestDeleteAwsBackup() {
	suite.testDeleteBackup(&awsScope, true)
}

func (suite *deleteNfsBackupSuite) TestDeleteGcpBackupFailure() {
	suite.testDeleteBackup(&gcpScope, false)
}

func (suite *deleteNfsBackupSuite) TestDeleteAwsBackupFailure() {
	suite.testDeleteBackup(&awsScope, false)
}

func TestDeleteNfsBackupSuite(t *testing.T) {
	suite.Run(t, new(deleteNfsBackupSuite))
}
