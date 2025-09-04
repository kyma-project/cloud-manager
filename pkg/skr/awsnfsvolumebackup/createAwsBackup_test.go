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

func (s *createAwsBackupSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *createAwsBackupSuite) TestCreateAwsBackupOnDeletingObj() {

	deletingObj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	s.Nil(err)

	state.vault = nil

	//Call createAwsBackup
	err, _ctx := createAwsBackup(ctx, state)
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *createAwsBackupSuite) TestCreateAwsBackupWhenIdsAreNotNil() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//update jobId and Id fields
	obj.Status.Id = "123456"
	obj.Status.JobId = "abcdef"
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	//Call createAwsBackup
	err, _ctx := createAwsBackup(ctx, state)
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *createAwsBackupSuite) TestCreateAwsBackupWhenIdsAreNil() {

	obj := awsNfsVolumeBackup.DeepCopy()
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

	//update jobId and Id fields with empty values
	obj.Status.State = ""
	obj.Status.Id = ""
	obj.Status.JobId = ""
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	//createAwsClient
	err, _ = createAwsClient(ctx, state)
	s.Nil(err)

	//loadVault
	err, _ = loadLocalVault(ctx, state)
	s.Nil(err)

	//Invoke API under test
	err, _ = createAwsBackup(ctx, state)
	s.Equal(composed.StopWithRequeueDelay(util.Timing.T1000ms()), err)

	fromK8s := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	s.Nil(err)

	s.Equal(cloudresourcesv1beta1.StateCreating, obj.Status.State)
	s.NotNil(obj.Status.Id)
	s.NotNil(obj.Status.JobId)
}

func TestCreateAwsBackup(t *testing.T) {
	suite.Run(t, new(createAwsBackupSuite))
}
