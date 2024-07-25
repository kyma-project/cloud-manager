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

type allocateIpRangeSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *allocateIpRangeSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *allocateIpRangeSuite) TestWhenStatusHasCidr() {
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
	ipRange.Status.Cidr = "10.10.10.10/28"

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Invoke the function under test
	err, resCtx := allocateIpRange(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()
	// check no error condition in status
	assert.Len(suite.T(), ipRange.Status.Conditions, 0)
}

func (suite *allocateIpRangeSuite) TestWhenSpecHasCidr() {
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

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Invoke the function under test
	err, resCtx := allocateIpRange(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()
	// check no error condition in status
	assert.Len(suite.T(), ipRange.Status.Conditions, 0)
}

func TestAllocateIpRange(t *testing.T) {
	suite.Run(t, new(allocateIpRangeSuite))
}
