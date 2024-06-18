package v2

import (
	"context"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
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

func (suite *allocateIpRangeSuite) TestWhenScopeNoNodes() {
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

	//gcpScope by default doesn't have nodes
	assert.Equal(suite.T(), "", ipRange.Status.Cidr)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Invoke the function under test
	err, resCtx := allocateIpRange(ctx, state)
	assert.Equal(suite.T(), composed.StopAndForget, err)
	assert.Equal(suite.T(), ctx, resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()
	// check error condition in status
	assert.Len(suite.T(), ipRange.Status.Conditions, 1)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ConditionTypeError, ipRange.Status.Conditions[0].Type)
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

func (suite *allocateIpRangeSuite) TestWhenScopeHasNodesNoSpecCidr() {
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
	scope := gcpScope.DeepCopy()
	scope.Spec.Scope.Gcp.Network = cloudcontrolv1beta1.GcpNetwork{
		Nodes:    "10.250.0.0/22",
		Pods:     "10.96.0.0/13",
		Services: "10.104.0.0/13",
	}

	//gcpScope by default doesn't have nodes
	assert.Equal(suite.T(), "", ipRange.Status.Cidr)

	state, err := factory.newStateWithScope(ctx, ipRange, scope)
	assert.Nil(suite.T(), err)

	//Invoke the function under test
	err, resCtx := allocateIpRange(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), ctx, resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()
	// check cidr is set in status
	assert.Greater(suite.T(), len(ipRange.Status.Cidr), 0)
	// check no error condition in status
	assert.Len(suite.T(), ipRange.Status.Conditions, 0)
}

func TestAllocateIpRange(t *testing.T) {
	suite.Run(t, new(allocateIpRangeSuite))
}
