package awsnfsvolumebackup

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type setIdempotencyTokenSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *setIdempotencyTokenSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *setIdempotencyTokenSuite) TestSetIdempotencyTokenOnDeletingObject() {

	deletingObj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	suite.Nil(err)
	state.Obj().SetFinalizers([]string{})

	//Call setIdempotencyToken
	err, _ctx := setIdempotencyToken(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *setIdempotencyTokenSuite) TestSetIdempotencyTokenWhenNfsVolumeReady() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the state to ready
	obj.Status.IdempotencyToken = uuid.NewString()
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	err, _ctx := setIdempotencyToken(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *setIdempotencyTokenSuite) TestSetIdempotencyTokenWhenEmpty() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ = setIdempotencyToken(ctx, state)
	suite.Equal(composed.StopWithRequeue, err)

	fromK8s := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.NotNil(fromK8s.Status.IdempotencyToken)
}

func TestSetIdempotencyToken(t *testing.T) {
	suite.Run(t, new(setIdempotencyTokenSuite))
}
