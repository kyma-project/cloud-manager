package awsnfsvolumebackup

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	backuptypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type updateStatusSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *updateStatusSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *updateStatusSuite) TestUpdateStatusOnDeletingObjAndNoRecoveryPoint() {

	deletingObj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	suite.Nil(err)
	state.recoveryPoint = nil

	//Call updateStatus
	err, _ctx := updateStatus(ctx, state)
	suite.Equal(composed.StopAndForget, err)
	suite.Nil(_ctx)
}

func (suite *updateStatusSuite) TestUpdateStatusOnDeletingObjWithRecoveryPoint() {

	deletingObj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	suite.Nil(err)
	state.recoveryPoint = &backup.DescribeRecoveryPointOutput{}

	//Call updateStatus
	err, _ctx := updateStatus(ctx, state)
	suite.Equal(composed.StopWithRequeueDelay(util.Timing.T1000ms()), err)
	suite.Nil(_ctx)
}

func (suite *updateStatusSuite) TestUpdateStatusWhenObjectReady() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the recovery point object
	state.recoveryPoint = &backup.DescribeRecoveryPointOutput{
		Status: backuptypes.RecoveryPointStatusCompleted,
	}

	//Set the condition to ready
	meta.SetStatusCondition(obj.Conditions(), metav1.Condition{
		Type:    cloudresourcesv1beta1.ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  "AfsNfsVolumeBackup is ready",
		Message: "AfsNfsVolumeBackup is ready",
	})
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	err, _ = updateStatus(ctx, state)
	suite.Equal(composed.StopAndForget, err)

	fromK8s := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: awsNfsVolumeBackup.Name,
			Namespace: awsNfsVolumeBackup.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(cloudresourcesv1beta1.StateReady, fromK8s.Status.State)
	readyCondition := meta.FindStatusCondition(fromK8s.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	suite.NotNil(readyCondition)
	suite.Equal(metav1.ConditionTrue, readyCondition.Status)

}

func (suite *updateStatusSuite) TestUpdateStatusWhenObjectNotReady() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolumeBackup
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the recovery point object
	state.recoveryPoint = &backup.DescribeRecoveryPointOutput{
		Status: backuptypes.RecoveryPointStatusCompleted,
	}
	//Update the ready condition
	obj.Status.Conditions = []metav1.Condition{}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	err, _ = updateStatus(ctx, state)
	suite.Equal(composed.StopAndForget, err)

	fromK8s := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: awsNfsVolumeBackup.Name,
			Namespace: awsNfsVolumeBackup.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(cloudresourcesv1beta1.StateReady, fromK8s.Status.State)
	readyCondition := meta.FindStatusCondition(fromK8s.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	suite.NotNil(readyCondition)
	suite.Equal(metav1.ConditionTrue, readyCondition.Status)
}

func (suite *updateStatusSuite) TestUpdateStatusWhenBackupJobFailed() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolumeBackup
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the recovery point object
	state.backupJob = &backup.DescribeBackupJobOutput{
		State: backuptypes.BackupJobStateFailed,
	}

	err, _ = updateStatus(ctx, state)
	suite.Equal(composed.StopAndForget, err)

	fromK8s := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: awsNfsVolumeBackup.Name,
			Namespace: awsNfsVolumeBackup.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(cloudresourcesv1beta1.StateError, fromK8s.Status.State)
	errCondition := meta.FindStatusCondition(fromK8s.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
	suite.NotNil(errCondition)
	suite.Equal(metav1.ConditionTrue, errCondition.Status)
}

func (suite *updateStatusSuite) TestUpdateStatusWhenRecoveryPointError() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolumeBackup
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the recovery point object
	state.recoveryPoint = &backup.DescribeRecoveryPointOutput{
		Status: backuptypes.RecoveryPointStatusExpired,
	}

	err, _ = updateStatus(ctx, state)
	suite.Equal(composed.StopAndForget, err)

	fromK8s := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: awsNfsVolumeBackup.Name,
			Namespace: awsNfsVolumeBackup.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(cloudresourcesv1beta1.StateError, fromK8s.Status.State)
	errCondition := meta.FindStatusCondition(fromK8s.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
	suite.NotNil(errCondition)
	suite.Equal(metav1.ConditionTrue, errCondition.Status)
}

func (suite *updateStatusSuite) TestUpdateStatusWithNoJobOrRecoveryPoint() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolumeBackup
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the recovery point object
	state.backupJob = nil
	state.recoveryPoint = nil

	err, _ctx := updateStatus(ctx, state)
	suite.Equal(composed.StopWithRequeueDelay(util.Timing.T1000ms()), err)
	suite.Nil(_ctx)
}

func TestUpdateStatus(t *testing.T) {
	suite.Run(t, new(updateStatusSuite))
}
