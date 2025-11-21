package gcpnfsvolume

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

	err, _ = removeFinalizer(ctx, state)
	assert.Equal(s.T(), composed.StopAndForget, err)
	assert.NotContains(s.T(), state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func (s *removeFinalizerSuite) TestDonNotRemoveFinalizerIfKcpNfsInstanceExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Add the finalizer to the object
	nfsVol := deletedGcpNfsVolume.DeepCopy()
	controllerutil.AddFinalizer(nfsVol, api.CommonFinalizerDeletionHook)
	err = factory.skrCluster.K8sClient().Update(ctx, nfsVol)
	assert.Nil(s.T(), err)

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(nfsVol)
	assert.Nil(s.T(), err)
	state.KcpNfsInstance = &gcpNfsInstanceToDelete

	err, _ = removeFinalizer(ctx, state)
	assert.Nil(s.T(), err)
	assert.Contains(s.T(), state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func (s *removeFinalizerSuite) TestDoNotRemoveFinalizerIfObjectIsNotDeleting() {
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

	//Call removeFinalizer
	err, _ = removeFinalizer(ctx, state)
	assert.Nil(s.T(), err)
	assert.Contains(s.T(), state.Obj().GetFinalizers(), api.CommonFinalizerDeletionHook)
}

func TestRemoveFinalizer(t *testing.T) {
	suite.Run(t, new(removeFinalizerSuite))
}
