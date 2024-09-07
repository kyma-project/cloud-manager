package v2

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/servicenetworking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
	"testing"
)

type syncPsaConnectionSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *syncPsaConnectionSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *syncPsaConnectionSuite) TestCreateSuccess() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp *compute.Operation
		switch r.Method {
		case http.MethodPost:
			if strings.HasSuffix(r.URL.Path, urlSvcNetworking) {
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
	state.connectionOp = client.ADD

	//Invoke the function under test
	err, _ = syncPsaConnection(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpOperationWaitTime), err)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	// check error operationId  in status
	assert.Equal(suite.T(), opIdentifier, ipRange.Status.OpIdentifier)
}

func (suite *syncPsaConnectionSuite) TestCreateFailure() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp *compute.Operation
		switch r.Method {
		case http.MethodPost:
			if strings.HasSuffix(r.URL.Path, urlSvcNetworking) {
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
	state.connectionOp = client.ADD

	//Invoke the function under test
	err, _ = syncPsaConnection(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime), err)

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
func (suite *syncPsaConnectionSuite) TestUpdate() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			if strings.HasSuffix(r.URL.Path, getUrlSvcNw) {
				resp := &servicenetworking.Operation{
					Done: true,
					Name: opIdentifier,
				}
				b, err := json.Marshal(resp)
				if err != nil {
					assert.Fail(suite.T(), "unable to marshal request: "+err.Error())
				}
				_, err = w.Write(b)
				if err != nil {
					assert.Fail(suite.T(), "unable to write to provided ResponseWriter: "+err.Error())
				}
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
	state.connectionOp = client.MODIFY
	state.ipRanges = []string{ipRange.Name}

	//Invoke the function under test
	err, _ = syncPsaConnection(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpOperationWaitTime), err)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	// check error operationId  in status
	assert.Equal(suite.T(), opIdentifier, ipRange.Status.OpIdentifier)
}

func (suite *syncPsaConnectionSuite) TestDeleteSuccess() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, getUrlCrmSvc) {
				resp := &cloudresourcemanager.Project{
					Name:          "test-project",
					ProjectId:     "test-project",
					ProjectNumber: 12345,
				}
				b, err := json.Marshal(resp)
				if err != nil {
					assert.Fail(suite.T(), "unable to marshal request: "+err.Error())
				}
				_, err = w.Write(b)
				if err != nil {
					assert.Fail(suite.T(), "unable to write to provided ResponseWriter: "+err.Error())
				}
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		case http.MethodPost:
			if strings.HasSuffix(r.URL.Path, getUrlSvcNw) {
				resp := &servicenetworking.Operation{
					Done: true,
					Name: opIdentifier,
				}
				b, err := json.Marshal(resp)
				if err != nil {
					assert.Fail(suite.T(), "unable to marshal request: "+err.Error())
				}
				_, err = w.Write(b)
				if err != nil {
					assert.Fail(suite.T(), "unable to write to provided ResponseWriter: "+err.Error())
				}
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
	state.connectionOp = client.DELETE

	//Invoke the function under test
	err, _ = syncPsaConnection(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpOperationWaitTime), err)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	// check error operationId  in status
	assert.Equal(suite.T(), opIdentifier, ipRange.Status.OpIdentifier)
}

func (suite *syncPsaConnectionSuite) TestDeleteFailure() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, getUrlCrmSvc) {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		case http.MethodPost:
			if strings.HasSuffix(r.URL.Path, getUrlSvcNw) {
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
	state.connectionOp = client.DELETE

	//Invoke the function under test
	err, _ = syncPsaConnection(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime), err)

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

func TestSyncPsaConnection(t *testing.T) {
	suite.Run(t, new(syncPsaConnectionSuite))
}
