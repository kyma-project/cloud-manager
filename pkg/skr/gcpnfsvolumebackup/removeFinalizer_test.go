package gcpnfsvolumebackup

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
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
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	deletingObj := deletingGpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, deletingObj)
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
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)

	//First add finalizer
	err, _ = addFinalizer(ctx, state)
	assert.Nil(suite.T(), err)
	//Call removeFinalizer
	err, _ = removeFinalizer(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), len(state.Obj().GetFinalizers()), 1)
}

func TestRemoveFinalizer(t *testing.T) {
	suite.Run(t, new(removeFinalizerSuite))
}
