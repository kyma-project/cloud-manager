package awsnfsvolumerestore

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type addFinalizerSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *addFinalizerSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *addFinalizerSuite) TestAddFinalizer() {

	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)
	err, _ = addFinalizer(ctx, state)
	s.Equal(composed.StopWithRequeue, err)
	s.Contains(state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func (s *addFinalizerSuite) TestAddFinalizerWhenAlreadyExists() {

	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	s.Nil(err)

	controllerutil.AddFinalizer(obj, api.CommonFinalizerDeletionHook)

	err, _ = addFinalizer(ctx, state)
	s.Nil(err)
	s.Contains(state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func (s *addFinalizerSuite) TestDoNotAddFinalizerOnDeletingObject() {

	deletingObj := deletingAwsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	s.Nil(err)
	state.Obj().SetFinalizers([]string{})

	//Call addFinalizer
	err, _ = addFinalizer(ctx, state)
	s.Nil(err)
	s.NotContains(state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func TestAddFinalizer(t *testing.T) {
	suite.Run(t, new(addFinalizerSuite))
}
