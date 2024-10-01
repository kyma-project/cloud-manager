package awsnfsvolumerestore

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type removeFinalizerSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *removeFinalizerSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *removeFinalizerSuite) TestRemoveFinalizer() {

	deletingObj := deletingAwsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(deletingObj)
	suite.Nil(err)
	err, _ = removeFinalizer(ctx, state)
	suite.Equal(composed.StopAndForget, err)
	suite.Equal(0, len(state.Obj().GetFinalizers()))
}

func (suite *removeFinalizerSuite) TestDoNotRemoveFinalizerIfNotDeleting() {
	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//First add finalizer
	err, _ = addFinalizer(ctx, state)
	suite.Equal(composed.StopWithRequeue, err)

	//Call removeFinalizer
	err, _ = removeFinalizer(ctx, state)
	suite.Nil(err)
	suite.Equal(1, len(state.Obj().GetFinalizers()))
}

func TestRemoveFinalizer(t *testing.T) {
	suite.Run(t, new(removeFinalizerSuite))
}
