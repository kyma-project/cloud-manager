package backupschedule

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type createBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *createBackupSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *createBackupSuite) TestWhenScheduleIsDeleting() {
	obj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with Gcp BackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke API under test
	err, _ctx := createBackup(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *createBackupSuite) TestWhenRunNotDue() {
	runTime := time.Now().Add(time.Minute).UTC()
	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	state.nextRunTime = runTime
	s.Nil(err)

	//Invoke API under test
	err, _ctx := createBackup(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *createBackupSuite) TestWhenRunAlreadyCompleted() {

	runTime := time.Now().UTC()
	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Update the creation run time to the next run time
	obj.Status.LastCreateRun = &metav1.Time{Time: runTime.UTC()}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	state.nextRunTime = runTime
	state.createRunCompleted = true
	s.Nil(err)

	//Invoke API under test
	err, _ctx := createBackup(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *createBackupSuite) TestCreateGcpBackup() {

	runTime := time.Now().UTC()

	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)
	state.Scope = &gcpScope
	state.nextRunTime = runTime
	state.backupImpl = &backupImplGcpNfs{}

	index := obj.Status.BackupIndex + 1

	//Invoke API under test
	err, _ = createBackup(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeue, err)

	bkupName := fmt.Sprintf("%s-%d-%s", obj.Name, index, state.nextRunTime.UTC().Format("20060102-150405"))

	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	s.Nil(err)
	s.Equal(v1beta1.JobStateActive, fromK8s.Status.State)
	s.Equal(bkupName, fromK8s.Status.LastCreatedBackup.Name)

	bkup := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: bkupName,
			Namespace: gcpNfsBackupSchedule.Namespace},
		bkup)
	s.Nil(err)
	s.Equal(bkupName, bkup.GetName())
}

func (s *createBackupSuite) TestCreateGcpBackupFailure() {

	runTime := time.Now().UTC()
	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	s.Nil(err)
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
	s.Nil(err)

	//Invoke API under test
	err, _ = createBackup(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeue, err)

	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	s.Nil(err)
	s.Equal(v1beta1.JobStateActive, fromK8s.Status.State)
}

func TestCreateBackupSuite(t *testing.T) {
	suite.Run(t, new(createBackupSuite))
}
