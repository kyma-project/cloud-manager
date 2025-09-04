package awsnfsvolumebackup

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type setIdempotencyTokenSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *setIdempotencyTokenSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *setIdempotencyTokenSuite) TestSetIdempotencyTokenOnDeletingObject() {

	deletingObj := deletingAwsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	s.Nil(err)
	state.Obj().SetFinalizers([]string{})

	//Call setIdempotencyToken
	err, _ctx := setIdempotencyToken(ctx, state)
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *setIdempotencyTokenSuite) TestSetIdempotencyTokenWhenNfsVolumeReady() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the state to ready
	obj.Status.IdempotencyToken = uuid.NewString()
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	err, _ctx := setIdempotencyToken(ctx, state)
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *setIdempotencyTokenSuite) TestSetIdempotencyTokenWhenEmpty() {

	obj := awsNfsVolumeBackup.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, _ = setIdempotencyToken(ctx, state)
	s.Equal(composed.StopWithRequeue, err)

	fromK8s := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	s.Nil(err)

	s.NotNil(fromK8s.Status.IdempotencyToken)
}

func TestSetIdempotencyToken(t *testing.T) {
	suite.Run(t, new(setIdempotencyTokenSuite))
}
