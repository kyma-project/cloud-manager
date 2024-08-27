package awsnfsvolumebackup

import (
	"context"
	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
	"github.com/stretchr/testify/suite"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type loadAwsBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *loadAwsBackupSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *loadAwsBackupSuite) TestLoadAwsBackupWhenIdIsNil() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//update jobId and Id fields with empty values
	obj.Status.Id = ""
	obj.Status.JobId = ""
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//Call loadAwsBackup
	err, _ctx := loadAwsBackup(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *loadAwsBackupSuite) TestLoadAwsBackupWhenJobNotExists() {

	obj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//update jobId and Id fields with empty values
	obj.Status.Id = "123456"
	obj.Status.JobId = "abcdef"
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//load the scope object into state
	awsScope := scope.DeepCopy()
	state.SetScope(awsScope)

	//createAwsClient
	err, _ = createAwsClient(ctx, state)
	suite.Nil(err)

	//Call loadAwsBackup
	err, _ctx := loadAwsBackup(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *loadAwsBackupSuite) TestLoadAwsBackupAfterCreatingBackup() {

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
	obj.Status.Id = state.awsClient.ParseRecoveryPointId(ptr.Deref(res.RecoveryPointArn, ""))
	obj.Status.JobId = ptr.Deref(res.BackupJobId, "")
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//loadAwsBackup
	err, _ = loadAwsBackup(ctx, state)
	suite.Nil(err)

}

func TestLoadAwsBackup(t *testing.T) {
	suite.Run(t, new(loadAwsBackupSuite))
}
