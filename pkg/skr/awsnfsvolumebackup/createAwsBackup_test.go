package awsnfsvolumebackup

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type createAwsBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *createAwsBackupSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *createAwsBackupSuite) TestCreateAwsBackupOnDeletingObj() {

	deletingObj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	suite.Nil(err)

	state.vault = nil

	//Call createAwsBackup
	err, _ctx := createAwsBackup(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *createAwsBackupSuite) TestCreateAwsBackupWhenIdsAreNotNil() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//update jobId and Id fields
	obj.Status.Id = "123456"
	obj.Status.JobId = "abcdef"
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//Call createAwsBackup
	err, _ctx := createAwsBackup(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *createAwsBackupSuite) TestCreateAwsBackupWhenIdsAreNil() {

	obj := awsNfsVolumeBackup.DeepCopy()
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

	//update jobId and Id fields with empty values
	obj.Status.State = ""
	obj.Status.Id = ""
	obj.Status.JobId = ""
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	//createAwsClient
	err, _ = createAwsClient(ctx, state)
	suite.Nil(err)

	//loadVault
	err, _ = loadLocalVault(ctx, state)
	suite.Nil(err)

	//Invoke API under test
	err, _ = createAwsBackup(ctx, state)
	suite.Equal(composed.StopWithRequeueDelay(util.Timing.T1000ms()), err)

	fromK8s := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(cloudresourcesv1beta1.StateCreating, obj.Status.State)
	suite.NotNil(obj.Status.Id)
	suite.NotNil(obj.Status.JobId)
}

func TestCreateAwsBackup(t *testing.T) {
	suite.Run(t, new(createAwsBackupSuite))
}
