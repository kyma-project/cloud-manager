package nfsbackupschedule

import (
	"context"
	"fmt"
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

type createNfsBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *createNfsBackupSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *createNfsBackupSuite) TestWhenNfsScheduleIsDeleting() {
	obj := deletingBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with NfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := createNfsBackup(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *createNfsBackupSuite) TestWhenRunNotDue() {
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
	err, _ctx := createNfsBackup(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *createNfsBackupSuite) TestWhenRunAlreadyCompleted() {

	runTime := time.Now().UTC()
	obj := nfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Update the creation run time to the next run time
	obj.Status.LastCreateRun = &metav1.Time{Time: runTime.UTC()}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//Get state object with NfsBackupSchedule
	state, err := factory.newStateWith(obj)
	state.nextRunTime = runTime
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := createNfsBackup(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *createNfsBackupSuite) testCreateBackup(scope *cloudcontrolv1beta1.Scope) {

	runTime := time.Now().UTC()

	obj := nfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with NfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	state.Scope = scope
	state.nextRunTime = runTime
	index := obj.Status.BackupIndex + 1

	//Invoke API under test
	err, _ = createNfsBackup(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeue, err)

	bkupName := fmt.Sprintf("%s-%d-%s", obj.Name, index, state.nextRunTime.UTC().Format("20060102-150405"))

	fromK8s := &v1beta1.NfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: nfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.Equal(v1beta1.JobStateActive, fromK8s.Status.State)
	suite.Equal(bkupName, fromK8s.Status.LastCreatedBackup.Name)

	var bkup client.Object
	switch scope.Spec.Provider {
	case cloudcontrolv1beta1.ProviderGCP:
		bkup = &v1beta1.GcpNfsVolumeBackup{}
	case cloudcontrolv1beta1.ProviderAws:
		bkup = &v1beta1.AwsNfsVolumeBackup{}
	default:
		suite.Fail("Invalid provider")
	}

	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: bkupName,
			Namespace: nfsBackupSchedule.Namespace},
		bkup)
	suite.Nil(err)
	suite.Equal(bkupName, bkup.GetName())
}

func (suite *createNfsBackupSuite) TestCreateGcpBackup() {
	suite.testCreateBackup(&gcpScope)
}

func (suite *createNfsBackupSuite) TestCreateAwsBackup() {
	suite.testCreateBackup(&awsScope)
}

func (suite *createNfsBackupSuite) testCreateBackupFailure(scope *cloudcontrolv1beta1.Scope) {

	runTime := time.Now().UTC()
	obj := nfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with NfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	state.Scope = scope
	state.nextRunTime = runTime
	index := obj.Status.BackupIndex + 1

	bkupName := fmt.Sprintf("%s-%d-%s", obj.Name, index, state.nextRunTime.UTC().Format("20060102-150405"))
	//Create backup with the same name to simulate failure
	var bkup client.Object
	switch scope.Spec.Provider {
	case cloudcontrolv1beta1.ProviderGCP:
		bkup = &v1beta1.GcpNfsVolumeBackup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bkupName,
				Namespace: nfsBackupSchedule.Namespace,
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
	case cloudcontrolv1beta1.ProviderAws:
		bkup = &v1beta1.AwsNfsVolumeBackup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bkupName,
				Namespace: nfsBackupSchedule.Namespace,
			},
			Spec: v1beta1.AwsNfsVolumeBackupSpec{
				Source: v1beta1.AwsNfsVolumeBackupSource{
					Volume: v1beta1.VolumeRef{
						Name:      obj.Spec.NfsVolumeRef.Name,
						Namespace: obj.Spec.NfsVolumeRef.Namespace,
					},
				},
			},
		}
	default:
		suite.Fail("Invalid provider")
	}
	err = factory.skrCluster.K8sClient().Create(ctx, bkup)
	suite.Nil(err)

	//Invoke API under test
	err, _ = createNfsBackup(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime), err)

	fromK8s := &v1beta1.NfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: nfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.Equal(v1beta1.JobStateError, fromK8s.Status.State)
}

func (suite *createNfsBackupSuite) TestCreateGcpBackupFailure() {
	suite.testCreateBackupFailure(&gcpScope)
}

func (suite *createNfsBackupSuite) TestCreateAwsBackupFailure() {
	suite.testCreateBackupFailure(&awsScope)
}

func TestCreateNfsBackupSuite(t *testing.T) {
	suite.Run(t, new(createNfsBackupSuite))
}
