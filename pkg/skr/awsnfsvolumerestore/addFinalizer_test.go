package awsnfsvolumerestore

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type addFinalizerSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *addFinalizerSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *addFinalizerSuite) TestAddFinalizer() {

	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with AwsNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	err, _ = addFinalizer(ctx, state)
	suite.Equal(composed.StopWithRequeue, err)
	suite.Contains(state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func (suite *addFinalizerSuite) TestAddFinalizerWhenAlreadyExists() {

	obj := awsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	controllerutil.AddFinalizer(obj, api.CommonFinalizerDeletionHook)

	err, _ = addFinalizer(ctx, state)
	suite.Nil(err)
	suite.Contains(state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func (suite *addFinalizerSuite) TestDoNotAddFinalizerOnDeletingObject() {

	deletingObj := deletingAwsNfsVolumeRestore.DeepCopy()
	factory, err := newStateFactoryWithObj(deletingObj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	suite.Nil(err)
	state.Obj().SetFinalizers([]string{})

	//Call addFinalizer
	err, _ = addFinalizer(ctx, state)
	suite.Nil(err)
	suite.NotContains(state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func TestAddFinalizer(t *testing.T) {
	suite.Run(t, new(addFinalizerSuite))
}
