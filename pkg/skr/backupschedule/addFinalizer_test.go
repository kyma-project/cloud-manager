package backupschedule

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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

	obj := gcpNfsBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(obj)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)
	err, _ = addFinalizer(ctx, state)
	assert.Nil(s.T(), err)
	assert.Contains(s.T(), state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func (s *addFinalizerSuite) TestDoNotAddFinalizerOnDeletingObject() {

	deletingObj := deletingGcpBackupSchedule.DeepCopy()
	factory, err := newTestStateFactoryWithObj(deletingObj)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	state.Obj().SetFinalizers([]string{})
	assert.Nil(s.T(), err)

	//Call addFinalizer
	err, _ = addFinalizer(ctx, state)
	assert.Nil(s.T(), err)
	assert.NotContains(s.T(), state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func TestAddFinalizer(t *testing.T) {
	suite.Run(t, new(addFinalizerSuite))
}
