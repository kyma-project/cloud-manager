package backupschedule

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
	"time"
)

type createBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *createBackupSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *createBackupSuite) TestWhenScheduleIsDeleting() {
	obj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with Gcp BackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := createBackup(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *createBackupSuite) TestWhenRunNotDue() {
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
	err, _ctx := createBackup(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *createBackupSuite) TestWhenRunAlreadyCompleted() {

	runTime := time.Now().UTC()
	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Update the creation run time to the next run time
	obj.Status.LastCreateRun = &metav1.Time{Time: runTime.UTC()}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	state.nextRunTime = runTime
	state.createRunCompleted = true
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := createBackup(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *createBackupSuite) TestCreateGcpBackup() {

	runTime := time.Now().UTC()

	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	state.Scope = &gcpScope
	state.nextRunTime = runTime
	state.backupImpl = &backupImplGcpNfs{}

	index := obj.Status.BackupIndex + 1

	//Invoke API under test
	err, _ = createBackup(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeue, err)

	bkupName := fmt.Sprintf("%s-%d-%s", obj.Name, index, state.nextRunTime.UTC().Format("20060102-150405"))

	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.Equal(v1beta1.JobStateActive, fromK8s.Status.State)
	suite.Equal(bkupName, fromK8s.Status.LastCreatedBackup.Name)

	bkup := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: bkupName,
			Namespace: gcpNfsBackupSchedule.Namespace},
		bkup)
	suite.Nil(err)
	suite.Equal(bkupName, bkup.GetName())
}

func (suite *createBackupSuite) TestCreateGcpBackupFailure() {

	runTime := time.Now().UTC()
	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	state.Scope = &gcpScope
	state.nextRunTime = runTime
	state.backupImpl = &backupImplGcpNfs{}

	index := obj.Status.BackupIndex + 1

	bkupName := fmt.Sprintf("%s-%d-%s", obj.Name, index, state.nextRunTime.UTC().Format("20060102-150405"))
	//Create backup with the same name to simulate failure
	bkup := &v1beta1.GcpNfsVolumeBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bkupName,
			Namespace: gcpNfsBackupSchedule.Namespace,
		},
		Spec: v1beta1.GcpNfsVolumeBackupSpec{
			Source: v1beta1.GcpNfsVolumeBackupSource{
				Volume: v1beta1.GcpNfsVolumeRef{
					Name:      obj.Spec.NfsVolumeRef.Name,
					Namespace: obj.Spec.NfsVolumeRef.Namespace,
				},
			},
			Location: "us-west1-a",
		},
	}

	err = factory.skrCluster.K8sClient().Create(ctx, bkup)
	suite.Nil(err)

	//Invoke API under test
	err, _ = createBackup(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeue, err)

	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.Equal(v1beta1.JobStateActive, fromK8s.Status.State)
}

func TestCreateBackupSuite(t *testing.T) {
	suite.Run(t, new(createBackupSuite))
}
