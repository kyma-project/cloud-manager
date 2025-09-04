package awsnfsvolumerestore

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type removeFinalizerSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *removeFinalizerSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *removeFinalizerSuite) TestRemoveFinalizer() {

	deletingObj := deletingAwsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(deletingObj)
	s.Nil(err)
	err, _ = removeFinalizer(ctx, state)
	s.Equal(composed.StopAndForget, err)
	s.Equal(0, len(state.Obj().GetFinalizers()))
}

func (s *removeFinalizerSuite) TestDoNotRemoveFinalizerIfNotDeleting() {
	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//First add finalizer
	err, _ = addFinalizer(ctx, state)
	s.Equal(composed.StopWithRequeue, err)

	//Call removeFinalizer
	err, _ = removeFinalizer(ctx, state)
	s.Nil(err)
	s.Equal(1, len(state.Obj().GetFinalizers()))
}

func TestRemoveFinalizer(t *testing.T) {
	suite.Run(t, new(removeFinalizerSuite))
}
