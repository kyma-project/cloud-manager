package backupschedule

import (
	"context"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/suite"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"testing"
	"time"
)

type deleteCascadeSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *deleteCascadeSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *deleteCascadeSuite) TestWhenScheduleIsNotDeleting() {
	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := deleteCascade(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *deleteCascadeSuite) TestWhenDeleteCascadeIsNotSet() {
	runTime := time.Now().Add(time.Minute).UTC()
	obj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	state.nextRunTime = runTime
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := deleteCascade(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *deleteCascadeSuite) TestWhenThereAreNoBackups() {

	obj := deletingGcpBackupSchedule.DeepCopy()
	obj.Spec.DeleteCascade = true
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := deleteCascade(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *deleteCascadeSuite) TestWhenNoBackupsExist() {

	obj := deletingGcpBackupSchedule.DeepCopy()
	obj.Spec.DeleteCascade = true
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke API under test
	err, _ctx := deleteCascade(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *deleteCascadeSuite) TestDeleteSchedule() {

	obj := deletingGcpBackupSchedule.DeepCopy()
	obj.Spec.DeleteCascade = true
	factory, err := newTestStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsBackupSchedule
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err = factory.skrCluster.K8sClient().Create(ctx, gcpBackup1.DeepCopy())
	suite.Nil(err)
	err = factory.skrCluster.K8sClient().Create(ctx, gcpBackup2.DeepCopy())
	suite.Nil(err)

	state.Backups = append(state.Backups, gcpBackup1)
	state.Backups = append(state.Backups, gcpBackup2)

	//Invoke API under test
	err, _ = deleteCascade(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeueDelay(util.Timing.T1000ms()), err)

	fromK8s := &v1beta1.GcpNfsBackupSchedule{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: gcpNfsBackupSchedule.Namespace},
		fromK8s)
	suite.Nil(err)

	backup := &v1beta1.GcpNfsVolumeBackup{}
	//Check whether the gcpBackup1 is deleted
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpBackup1.GetName(),
			Namespace: obj.Namespace},
		backup)
	suite.Equal(true, apierrors.IsNotFound(err))

	//Check whether the gcpBackup2 is deleted
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpBackup2.GetName(),
			Namespace: obj.Namespace},
		backup)
	suite.Equal(true, apierrors.IsNotFound(err))
}

func TestDeleteCascadeSuite(t *testing.T) {
	suite.Run(t, new(deleteCascadeSuite))
}
