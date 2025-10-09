package awsnfsvolumebackup

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
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

func (s *deleteAwsBackupSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *deleteAwsBackupSuite) TestDeleteAwsBackupWhenNotDeleting() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Call deleteAwsBackup
	err, _ctx := deleteLocalAwsBackup(ctx, state)
	s.Nil(err)
	s.Equal(ctx, _ctx)
}

func (s *deleteAwsBackupSuite) TestDeleteAwsBackupWhenRecoveryPointIsNil() {

	obj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Call deleteAwsBackup
	err, _ctx := deleteLocalAwsBackup(ctx, state)
	s.Nil(err)
	s.Equal(ctx, _ctx)
}

func (s *deleteAwsBackupSuite) TestDeleteAwsBackupAfterCreatingBackup() {

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

	//Invoke API under test
	err, _ = deleteLocalAwsBackup(ctx, state)
	s.Equal(composed.StopWithRequeue, err)

	fromK8s := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	s.Nil(err)

	s.Equal(cloudresourcesv1beta1.StateDeleting, obj.Status.State)
	s.NotNil(obj.Status.Id)
	s.NotNil(obj.Status.JobId)
}

func TestDeleteAwsBackup(t *testing.T) {
	suite.Run(t, new(deleteAwsBackupSuite))
}
