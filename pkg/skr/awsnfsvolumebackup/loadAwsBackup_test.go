package awsnfsvolumebackup

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
	"github.com/stretchr/testify/suite"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type loadAwsBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *loadAwsBackupSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *loadAwsBackupSuite) TestLoadAwsBackupWhenIdIsNil() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//update jobId and Id fields with empty values
	obj.Status.Id = ""
	obj.Status.JobId = ""
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	//Call loadAwsBackup
	err, _ctx := loadLocalAwsBackup(ctx, state)
	s.Nil(err)
	s.Equal(ctx, _ctx)
}

func (s *loadAwsBackupSuite) TestLoadAwsBackupWhenJobNotExists() {

	obj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//update jobId and Id fields with empty values
	obj.Status.Id = "123456"
	obj.Status.JobId = "abcdef"
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	//load the scope object into state
	awsScope := scope.DeepCopy()
	state.SetScope(awsScope)

	//createAwsClient
	err, _ = createAwsClient(ctx, state)
	s.Nil(err)

	//Call loadAwsBackup
	err, _ctx := loadLocalAwsBackup(ctx, state)
	s.Nil(err)
	s.Equal(ctx, _ctx)
}

func (s *loadAwsBackupSuite) TestLoadAwsBackupAfterCreatingBackup() {

	obj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//load the scope object into state
	awsScope := scope.DeepCopy()
	state.SetScope(awsScope)

	//createAwsClient
	err, _ = createAwsClient(ctx, state)
	s.Nil(err)

	//loadVault
	err, _ = loadLocalVault(ctx, state)
	s.Nil(err)

	//createAwsBackup
	res, err := state.awsClient.StartBackupJob(ctx, &client.StartBackupJobInput{
		BackupVaultName:   state.GetVaultName(),
		IamRoleArn:        state.roleName,
		ResourceArn:       state.GetFileSystemArn(),
		RecoveryPointTags: state.GetTags(),
	})
	s.Nil(err)

	//update jobId and Id fields with empty values
	obj.Status.State = cloudresourcesv1beta1.StateReady
	obj.Status.Id = awsutil.ParseArnResourceId(ptr.Deref(res.RecoveryPointArn, ""))
	obj.Status.JobId = ptr.Deref(res.BackupJobId, "")
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	//loadAwsBackup
	err, _ = loadLocalAwsBackup(ctx, state)
	s.Nil(err)

}

func TestLoadAwsBackup(t *testing.T) {
	suite.Run(t, new(loadAwsBackupSuite))
}
