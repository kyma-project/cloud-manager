package gcpnfsvolume

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type removePersistenceVolumeFinalizerSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *removePersistenceVolumeFinalizerSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *removePersistenceVolumeFinalizerSuite) TestRemoveFinalizer() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(&deletedGcpNfsVolume)
	assert.Nil(s.T(), err)
	state.PV = &pvDeletingGcpNfsVolume

	//Create PV.
	err = state.SkrCluster.K8sClient().Create(ctx, state.PV)
	assert.Nil(s.T(), err)

	err, _ = removePersistenceVolumeFinalizer(ctx, state)
	assert.Nil(s.T(), err)

	pvName := fmt.Sprintf("%s--%s", deletedGcpNfsVolume.Namespace, deletedGcpNfsVolume.Name)
	pv := corev1.PersistentVolume{}
	err = state.SkrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: pvName}, &pv)
	assert.Nil(s.T(), err)

	assert.NotContains(s.T(), pv.GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func (s *removePersistenceVolumeFinalizerSuite) TestContinueIfPVNotExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(&deletedGcpNfsVolume)
	assert.Nil(s.T(), err)

	err, _ = removePersistenceVolumeFinalizer(ctx, state)
	assert.Nil(s.T(), err)
}

func (s *removePersistenceVolumeFinalizerSuite) TestDoNotRemoveFinalizerIfObjectIsNotDeleting() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with Deleted GcpNfsVolume
	state, err := factory.newState()
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), err)

	//Call removePersistenceVolumeFinalizer
	err, _ = removePersistenceVolumeFinalizer(ctx, state)
	assert.Nil(s.T(), err)
	assert.Contains(s.T(), state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func TestRemovePersistenceVolumeFinalizer(t *testing.T) {
	suite.Run(t, new(removePersistenceVolumeFinalizerSuite))
}
