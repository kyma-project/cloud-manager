package gcpnfsvolume

import (
	"context"
	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newStateWith(&deletedGcpNfsVolume)

	err, _ = removeFinalizer(ctx, state)
	assert.Equal(suite.T(), composed.StopAndForget, err)
	assert.NotContains(suite.T(), state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func (suite *removeFinalizerSuite) TestDonNotRemoveFinalizerIfKcpNfsInstanceExists() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Add the finalizer to the object
	nfsVol := deletedGcpNfsVolume.DeepCopy()
	controllerutil.AddFinalizer(nfsVol, api.CommonFinalizerDeletionHook)
	err = factory.skrCluster.K8sClient().Update(ctx, nfsVol)
	assert.Nil(suite.T(), err)

	//Get state object with GcpNfsVolume
	state := factory.newStateWith(nfsVol)
	state.KcpNfsInstance = &gcpNfsInstanceToDelete

	err, _ = removeFinalizer(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Contains(suite.T(), state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func (suite *removeFinalizerSuite) TestDoNotRemoveFinalizerIfObjectIsNotDeleting() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with Deleted GcpNfsVolume
	state := factory.newState()
	assert.Nil(suite.T(), err)

	//Call removeFinalizer
	err, _ = removeFinalizer(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Contains(suite.T(), state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func TestRemoveFinalizer(t *testing.T) {
	suite.Run(t, new(removeFinalizerSuite))
}
