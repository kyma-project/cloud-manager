package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
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
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	deletingObj := deletingGpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, deletingObj)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(deletingObj)
	assert.Nil(s.T(), err)
	err, _ = removeFinalizer(ctx, state)
	assert.Equal(s.T(), composed.StopAndForget, err)
	assert.Equal(s.T(), len(state.Obj().GetFinalizers()), 0)
}

func (s *removeFinalizerSuite) TestDoNotRemoveFinalizerIfNotDeleting() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)

	//First add finalizer
	err, _ = addFinalizer(ctx, state)
	assert.Nil(s.T(), err)
	//Call removeFinalizer
	err, _ = removeFinalizer(ctx, state)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), len(state.Obj().GetFinalizers()), 1)
}

func TestRemoveFinalizer(t *testing.T) {
	suite.Run(t, new(removeFinalizerSuite))
}
