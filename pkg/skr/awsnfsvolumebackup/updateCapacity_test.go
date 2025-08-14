package awsnfsvolumebackup

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/backup"
	backuptypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type updateCapacitySuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *updateCapacitySuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *updateCapacitySuite) TestWhenBackupIsDeleting() {

	obj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := updateCapacity(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *updateCapacitySuite) TestWhenRecoveryPointNotExists() {

	obj := awsNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.StateProcessing
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := updateCapacity(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)

	fromK8s := &v1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(v1beta1.StateProcessing, fromK8s.Status.State)
}

func (suite *updateCapacitySuite) TestWhenUpdateIsNotDue() {

	config.AwsConfig.EfsCapacityCheckInterval = time.Duration(6 * time.Hour)

	obj := awsNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.StateCreating
	obj.Status.LastCapacityUpdate = &metav1.Time{Time: time.Now().Add(-1 * time.Hour)}
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the recovery point object
	var backupSize int64 = 1075
	state.recoveryPoint = &backup.DescribeRecoveryPointOutput{
		Status:            backuptypes.RecoveryPointStatusCompleted,
		BackupSizeInBytes: &backupSize,
	}

	err, _ctx := updateCapacity(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)

	fromK8s := &v1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(v1beta1.StateCreating, fromK8s.Status.State)
}

func (suite *updateCapacitySuite) TestWhenLastCapacityUpdateNotExists() {

	obj := awsNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.StateCreating
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the recovery point object
	var backupSize int64 = 15676
	state.recoveryPoint = &backup.DescribeRecoveryPointOutput{
		Status:            backuptypes.RecoveryPointStatusCompleted,
		BackupSizeInBytes: &backupSize,
	}

	err, _ctx := updateCapacity(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Equal(ctx, _ctx)

	fromK8s := &v1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(v1beta1.StateCreating, fromK8s.Status.State)
	suite.Equal(backupSize, fromK8s.Status.Capacity.Value())
	suite.GreaterOrEqual(1*time.Second, time.Since(fromK8s.Status.LastCapacityUpdate.Time))
}

func (suite *updateCapacitySuite) TestWhenLastCapacityUpdateIsZero() {

	obj := awsNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.StateCreating
	obj.Status.LastCapacityUpdate = &metav1.Time{}
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the recovery point object
	var backupSize int64 = 15726
	state.recoveryPoint = &backup.DescribeRecoveryPointOutput{
		Status:            backuptypes.RecoveryPointStatusCompleted,
		BackupSizeInBytes: &backupSize,
	}

	err, _ctx := updateCapacity(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Equal(ctx, _ctx)

	fromK8s := &v1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(v1beta1.StateCreating, fromK8s.Status.State)
	suite.Equal(backupSize, fromK8s.Status.Capacity.Value())
	suite.GreaterOrEqual(1*time.Second, time.Since(fromK8s.Status.LastCapacityUpdate.Time))
}

func TestUpdateCapacity(t *testing.T) {
	suite.Run(t, new(updateCapacitySuite))
}
