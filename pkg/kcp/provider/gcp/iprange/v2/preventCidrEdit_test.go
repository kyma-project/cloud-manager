package v2

import (
	"context"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type preventCidrEditSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *preventCidrEditSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *preventCidrEditSuite) TestWhenCidrIsSetEqualReady() {
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
	ipRange.Status.Cidr = ipRange.Spec.Cidr
	ipRange.Status.Conditions = []metav1.Condition{
		{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Message: "Ready",
		},
	}

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Invoke the function under test
	err, resCtx := preventCidrEdit(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	newIpRange := state.ObjAsIpRange()
	assert.Equal(suite.T(), newIpRange.Status.Cidr, ipRange.Status.Cidr)

	// check error condition in status
	assert.Len(suite.T(), newIpRange.Status.Conditions, 0)
}

func (suite *preventCidrEditSuite) TestWhenCidrIsSetNotEqualNotReady() {
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
	err, resCtx := preventCidrEdit(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	newIpRange := state.ObjAsIpRange()
	assert.Equal(suite.T(), newIpRange.Status.Cidr, ipRange.Status.Cidr)

	// check error condition in status
	assert.Len(suite.T(), newIpRange.Status.Conditions, 0)
}

func (suite *preventCidrEditSuite) TestWhenCidrIsSetNotEqualReady() {
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
	ipRange.Status.Conditions = []metav1.Condition{
		{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Message: "Ready",
		},
	}

	ipRange.Status.Cidr = "10.10.10.10/28"

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Invoke the function under test
	err, resCtx := preventCidrEdit(ctx, state)
	assert.Equal(suite.T(), composed.StopAndForget, err)
	assert.Equal(suite.T(), ctx, resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	newIpRange := state.ObjAsIpRange()
	assert.Equal(suite.T(), newIpRange.Status.Cidr, ipRange.Status.Cidr)

	// check error condition in status
	assert.Len(suite.T(), newIpRange.Status.Conditions, 1)
}

func (suite *preventCidrEditSuite) TestWhenSpecCidrIsNotSetReady() {
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
	ipRange.Status.Cidr = "10.10.10.10/28"
	ipRange.Status.Conditions = []metav1.Condition{
		{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Message: "Ready",
		},
	}
	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Invoke the function under test
	err, resCtx := preventCidrEdit(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)
}

func (suite *preventCidrEditSuite) TestWhenStatusCidrIsNotSetReady() {
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
	ipRange.Status.Conditions = []metav1.Condition{
		{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Message: "Ready",
		},
	}
	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Invoke the function under test
	err, resCtx := preventCidrEdit(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)
}

func TestPreventCidrEdit(t *testing.T) {
	suite.Run(t, new(preventCidrEditSuite))
}
