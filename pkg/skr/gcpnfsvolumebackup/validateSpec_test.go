package gcpnfsvolumebackup

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type validateSpecSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *validateSpecSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *validateSpecSuite) TestValidateLocationEmpty() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Spec.Location = ""
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	err, _ = validateLocation(ctx, state)
	if feature.GcpNfsVolumeAutomaticLocationAllocation.Value(suite.ctx) {
		assert.Nil(suite.T(), err)
	} else {
		assert.Equal(suite.T(), composed.StopAndForget, err)
	}
}

func (suite *validateSpecSuite) TestValidateLocationNotEmpty() {
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
	err, _ = validateLocation(ctx, state)
	assert.Nil(suite.T(), err)
}

func TestValidateSpec(t *testing.T) {
	suite.Run(t, new(validateSpecSuite))
}
