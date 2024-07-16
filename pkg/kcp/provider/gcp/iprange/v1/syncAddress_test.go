package v1

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
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

type syncAddressSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *syncAddressSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *syncAddressSuite) TestCreateSuccess() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp *compute.Operation
		switch r.Method {
		case http.MethodPost:
			if strings.HasSuffix(r.URL.Path, urlGlobalAddress) {
				resp = &compute.Operation{
					Name: opIdentifier,
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
	state.addressOp = client.ADD

	//Invoke the function under test
	err, _ = syncAddress(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(state.gcpConfig.GcpOperationWaitTime), err)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	// check error operationId  in status
	assert.Equal(suite.T(), opIdentifier, ipRange.Status.OpIdentifier)
}

func (suite *syncAddressSuite) TestCreateFailure() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			if strings.HasSuffix(r.URL.Path, urlGlobalAddress) {
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
	state.addressOp = client.ADD

	//Invoke the function under test
	err, _ = syncAddress(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime), err)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	// check error condition in status
	assert.Len(suite.T(), ipRange.Status.Conditions, 1)
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, ipRange.Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, ipRange.Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonGcpError, ipRange.Status.Conditions[0].Reason)
}
func (suite *syncAddressSuite) TestUpdate() {
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
	state.addressOp = client.MODIFY

	//Invoke the function under test
	err, resCtx := syncAddress(ctx, state)
	assert.Equal(suite.T(), composed.StopAndForget, err)
	assert.Nil(suite.T(), resCtx)
}

func (suite *syncAddressSuite) TestDeleteSuccess() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp *compute.Operation
		switch r.Method {
		case http.MethodDelete:
			if strings.HasSuffix(r.URL.Path, getUrlCompute) {
				resp = &compute.Operation{
					Name: opIdentifier,
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
	state.addressOp = client.DELETE

	//Invoke the function under test
	err, _ = syncAddress(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(state.gcpConfig.GcpOperationWaitTime), err)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	// check error operationId  in status
	assert.Equal(suite.T(), opIdentifier, ipRange.Status.OpIdentifier)
}

func (suite *syncAddressSuite) TestDeleteFailure() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp *compute.Operation
		switch r.Method {
		case http.MethodDelete:
			if strings.HasSuffix(r.URL.Path, getUrlCompute) {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
	state.addressOp = client.DELETE

	//Invoke the function under test
	err, _ = syncAddress(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime), err)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	// check error condition in status
	assert.Len(suite.T(), ipRange.Status.Conditions, 1)
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, ipRange.Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, ipRange.Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonGcpError, ipRange.Status.Conditions[0].Reason)
}

func TestSyncAddress(t *testing.T) {
	suite.Run(t, new(syncAddressSuite))
}
