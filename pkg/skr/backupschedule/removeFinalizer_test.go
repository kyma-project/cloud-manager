package backupschedule

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
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

	deletingObj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(deletingObj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(deletingObj)
	assert.Nil(suite.T(), err)
	err, _ = removeFinalizer(ctx, state)
	assert.Equal(suite.T(), composed.StopAndForget, err)
	assert.Equal(suite.T(), len(state.Obj().GetFinalizers()), 0)
}

func (suite *removeFinalizerSuite) TestDoNotRemoveFinalizerIfNotDeleting() {

	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)

	//First add finalizer
	err, _ = addFinalizer(ctx, state)
	//Call removeFinalizer
	err, _ = removeFinalizer(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), len(state.Obj().GetFinalizers()), 1)
}

func TestRemoveFinalizer(t *testing.T) {
	suite.Run(t, new(removeFinalizerSuite))
}
