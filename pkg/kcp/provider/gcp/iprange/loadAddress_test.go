package iprange

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/compute/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
	"testing"
)

type loadAddressSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *loadAddressSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *loadAddressSuite) TestWhenNotFound() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, getUrlCompute) {
				//Return 404
				http.Error(w, "Not Found", http.StatusNotFound)
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
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
	err, resCtx := loadAddress(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	// check error condition in status
	assert.Len(suite.T(), ipRange.Status.Conditions, 0)
}

func (suite *loadAddressSuite) TestWhenErrorResponse() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, getUrlCompute) {
				//Return 404
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
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
	err, resCtx := loadAddress(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeue, err)
	assert.Nil(suite.T(), resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	// check error condition in status
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, ipRange.Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, ipRange.Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonGcpError, ipRange.Status.Conditions[0].Reason)
}

func (suite *loadAddressSuite) TestWhenGcpAddressHasWrongVPC() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp *compute.Address
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, getUrlCompute) {
				resp = &compute.Address{
					Address:      ipAddr,
					Name:         gcpIpRange.Spec.RemoteRef.Name,
					Network:      "not-matching-vpc",
					PrefixLength: int64(prefix),
				}
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
		b, err := json.Marshal(resp)
		if err != nil {
			assert.Fail(suite.T(), "unable to marshal request: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(suite.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
		return
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
	err, resCtx := loadAddress(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeue, err)
	assert.Nil(suite.T(), resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	// check error condition in status
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, ipRange.Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, ipRange.Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonGcpError, ipRange.Status.Conditions[0].Reason)
}

func (suite *loadAddressSuite) TestWhenGcpAddressHasCorrectVPC() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp *compute.Address
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, getUrlCompute) {
				resp = &compute.Address{
					Address:      ipAddr,
					Name:         gcpIpRange.Spec.RemoteRef.Name,
					Network:      "test-vpc",
					PrefixLength: int64(prefix),
				}
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
		b, err := json.Marshal(resp)
		if err != nil {
			assert.Fail(suite.T(), "unable to marshal request: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(suite.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
		return
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
	err, resCtx := loadAddress(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	// check error condition in status
	assert.Len(suite.T(), ipRange.Status.Conditions, 0)
	assert.NotNil(suite.T(), state.address)
}

func TestLoadAddress(t *testing.T) {
	suite.Run(t, new(loadAddressSuite))
}
