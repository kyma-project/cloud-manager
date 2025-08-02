package awsnfsvolumebackup

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type deleteAwsBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *deleteAwsBackupSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *deleteAwsBackupSuite) TestDeleteAwsBackupWhenNotDeleting() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Call deleteAwsBackup
	err, _ctx := deleteAwsBackup(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *deleteAwsBackupSuite) TestDeleteAwsBackupWhenRecoveryPointIsNil() {

	obj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Call deleteAwsBackup
	err, _ctx := deleteAwsBackup(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *deleteAwsBackupSuite) TestDeleteAwsBackupAfterCreatingBackup() {

	obj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//load the scope object into state
	awsScope := scope.DeepCopy()
	state.SetScope(awsScope)

	//createAwsClient
	err, _ = createAwsClient(ctx, state)
	suite.Nil(err)

	//loadVault
	err, _ = loadVault(ctx, state)
	suite.Nil(err)

	//createAwsBackup
	res, err := state.awsClient.StartBackupJob(ctx, &client.StartBackupJobInput{
		BackupVaultName:   state.GetVaultName(),
		IamRoleArn:        state.roleName,
		ResourceArn:       state.GetFileSystemArn(),
		RecoveryPointTags: state.GetTags(),
	})
	suite.Nil(err)

	//update jobId and Id fields with empty values
	obj.Status.State = cloudresourcesv1beta1.StateReady
	obj.Status.Location, _, obj.Status.Id = state.awsClient.ParseRecoveryPointId(ptr.Deref(res.RecoveryPointArn, ""))
	obj.Status.JobId = ptr.Deref(res.BackupJobId, "")
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//loadAwsBackup
	err, _ = loadAwsBackup(ctx, state)
	suite.Nil(err)

	//Invoke API under test
	err, _ = deleteAwsBackup(ctx, state)
	suite.Equal(composed.StopWithRequeue, err)

	fromK8s := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(cloudresourcesv1beta1.StateDeleting, obj.Status.State)
	suite.NotNil(obj.Status.Id)
	suite.NotNil(obj.Status.JobId)
}

func TestDeleteAwsBackup(t *testing.T) {
	suite.Run(t, new(deleteAwsBackupSuite))
}
