package backupschedule

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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

	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	err, _ = addFinalizer(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Contains(suite.T(), state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func (suite *addFinalizerSuite) TestDoNotAddFinalizerOnDeletingObject() {

	deletingObj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(deletingObj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	state.Obj().SetFinalizers([]string{})
	assert.Nil(suite.T(), err)

	//Call addFinalizer
	err, _ = addFinalizer(ctx, state)
	assert.Nil(suite.T(), err)
	assert.NotContains(suite.T(), state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func TestAddFinalizer(t *testing.T) {
	suite.Run(t, new(addFinalizerSuite))
}
