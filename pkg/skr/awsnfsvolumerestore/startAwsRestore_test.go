package awsnfsvolumerestore

import (
	"context"
	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type startAwsRestoreSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *startAwsRestoreSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *startAwsRestoreSuite) TestStartAwsRestoreDeleting() {

	obj := deletingAwsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	err, ctx = startAwsRestore(ctx, state)
	suite.Nil(err)
	suite.Nil(ctx)
}

func (suite *startAwsRestoreSuite) TestStartAwsRestoreWithJobId() {

	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	err, ctx = startAwsRestore(ctx, state)
	suite.Nil(err)
	suite.Nil(ctx)
}

func (suite *startAwsRestoreSuite) TestStartAwsRestoreWithoutJobId() {

	obj := awsNfsVolumeRestore.DeepCopy()
	obj.Status.JobId = ""
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	//load the scope object into state
	awsScope := scope.DeepCopy()
	state.SetScope(awsScope)
	state.skrAwsNfsVolumeBackup = awsNfsVolumeBackup.DeepCopy()
	state.skrAwsNfsVolume = awsNfsVolume.DeepCopy()

	//createAwsClient
	err, _ = createAwsClient(ctx, state)
	suite.Nil(err)

	err, _ctx := startAwsRestore(ctx, state)
	suite.Equal(composed.StopWithRequeueDelay(util.Timing.T1000ms()), err)
	suite.Equal(ctx, _ctx)
	result := &cloudresourcesv1beta1.AwsNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Namespace: obj.Namespace, Name: obj.Name}, result)
	suite.Nil(err)
	suite.Equal(cloudresourcesv1beta1.JobStateInProgress, result.Status.State)
	suite.NotEmpty(result.Status.JobId)
}

func TestStartAwsRestore(t *testing.T) {
	suite.Run(t, new(startAwsRestoreSuite))
}
