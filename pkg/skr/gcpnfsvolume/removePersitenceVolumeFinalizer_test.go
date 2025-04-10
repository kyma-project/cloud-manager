package gcpnfsvolume

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type removePersistenceVolumeFinalizerSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *removePersistenceVolumeFinalizerSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *removePersistenceVolumeFinalizerSuite) TestRemoveFinalizer() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newStateWith(&deletedGcpNfsVolume)
	state.PV = &pvDeletingGcpNfsVolume

	//Create PV.
	err = state.SkrCluster.K8sClient().Create(ctx, state.PV)
	assert.Nil(suite.T(), err)

	err, _ = removePersistenceVolumeFinalizer(ctx, state)
	assert.Nil(suite.T(), err)

	pvName := fmt.Sprintf("%s--%s", deletedGcpNfsVolume.Namespace, deletedGcpNfsVolume.Name)
	pv := corev1.PersistentVolume{}
	err = state.SkrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: pvName}, &pv)
	assert.Nil(suite.T(), err)

	assert.NotContains(suite.T(), pv.GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func (suite *removePersistenceVolumeFinalizerSuite) TestContinueIfPVNotExists() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state := factory.newStateWith(&deletedGcpNfsVolume)

	err, _ = removePersistenceVolumeFinalizer(ctx, state)
	assert.Nil(suite.T(), err)
}

func (suite *removePersistenceVolumeFinalizerSuite) TestDoNotRemoveFinalizerIfObjectIsNotDeleting() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with Deleted GcpNfsVolume
	state := factory.newState()
	assert.Nil(suite.T(), err)

	//Call removePersistenceVolumeFinalizer
	err, _ = removePersistenceVolumeFinalizer(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Contains(suite.T(), state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func TestRemovePersistenceVolumeFinalizer(t *testing.T) {
	suite.Run(t, new(removePersistenceVolumeFinalizerSuite))
}
