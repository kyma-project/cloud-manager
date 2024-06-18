package v2

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type copyCidrToStatusSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *copyCidrToStatusSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *copyCidrToStatusSuite) TestWhenCidrIsSet() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()

	assert.Equal(suite.T(), "", ipRange.Status.Cidr)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Invoke the function under test
	err, resCtx := copyCidrToStatus(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), ctx, resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()
	assert.Equal(suite.T(), ipRange.Spec.Cidr, ipRange.Status.Cidr)

	// check error condition in status
	assert.Len(suite.T(), ipRange.Status.Conditions, 0)
}

func (suite *copyCidrToStatusSuite) TestWhenCidrIsNotSet() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()
	ipRange.Spec.Cidr = ""

	assert.Equal(suite.T(), "", ipRange.Status.Cidr)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Invoke the function under test
	err, resCtx := copyCidrToStatus(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), ctx, resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()
	assert.Equal(suite.T(), "", ipRange.Status.Cidr)

	// check error condition in status
	assert.Len(suite.T(), ipRange.Status.Conditions, 0)
}

func TestCopyCidrToStatus(t *testing.T) {
	suite.Run(t, new(copyCidrToStatusSuite))
}
