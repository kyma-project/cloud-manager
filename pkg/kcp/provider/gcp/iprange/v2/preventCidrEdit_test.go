package v2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type preventCidrEditSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *preventCidrEditSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *preventCidrEditSuite) TestWhenCidrIsSetEqualReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)

	//Invoke the function under test
	err, resCtx := preventCidrEdit(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(s.T(), err)
	newIpRange := state.ObjAsIpRange()
	assert.Equal(s.T(), newIpRange.Status.Cidr, ipRange.Status.Cidr)

	// check error condition in status
	assert.Len(s.T(), newIpRange.Status.Conditions, 0)
}

func (s *preventCidrEditSuite) TestWhenCidrIsSetNotEqualNotReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()
	ipRange.Status.Cidr = "10.10.10.10/28"

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Invoke the function under test
	err, resCtx := preventCidrEdit(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(s.T(), err)
	newIpRange := state.ObjAsIpRange()
	assert.Equal(s.T(), newIpRange.Status.Cidr, ipRange.Status.Cidr)

	// check error condition in status
	assert.Len(s.T(), newIpRange.Status.Conditions, 0)
}

func (s *preventCidrEditSuite) TestWhenSpecCidrIsNotSetReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)

	//Invoke the function under test
	err, resCtx := preventCidrEdit(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)
}

func (s *preventCidrEditSuite) TestWhenStatusCidrIsNotSetReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)

	//Invoke the function under test
	err, resCtx := preventCidrEdit(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)
}

func TestPreventCidrEdit(t *testing.T) {
	suite.Run(t, new(preventCidrEditSuite))
}
