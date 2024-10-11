package v2

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/servicenetworking/v1"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
	"testing"
)

type checkGcpOperationSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *checkGcpOperationSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *checkGcpOperationSuite) TestWhenOpIdentifierIsNil() {
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
	err, resCtx := checkGcpOperation(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)
}

func (suite *checkGcpOperationSuite) TestWhenSvcNwOperationNotComplete() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var opResp *servicenetworking.Operation
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, opIdentifier) {
				opResp = &servicenetworking.Operation{
					Name: "create-operation",
					Done: false,
				}
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
		b, err := json.Marshal(opResp)
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

	//Get state object with IpRange
	ipRange := gcpIpRange.DeepCopy()
	ipRange.Status.State = client.DeletePsaConnection
	ipRange.Status.OpIdentifier = opIdentifier
	err = factory.kcpCluster.K8sClient().Status().Update(ctx, ipRange)
	assert.Nil(suite.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Invoke the function under test
	err, resCtx := checkGcpOperation(ctx, state)
	assert.Nil(suite.T(), resCtx)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime), err)
}

func (suite *checkGcpOperationSuite) TestWhenSvcNwOperationSuccessful() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var opResp *servicenetworking.Operation
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, opIdentifier) {
				opResp = &servicenetworking.Operation{
					Name: "create-operation",
					Done: true,
				}
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
		b, err := json.Marshal(opResp)
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

	//Get state object with IpRange
	ipRange := gcpIpRange.DeepCopy()
	ipRange.Status.State = client.DeletePsaConnection
	ipRange.Status.OpIdentifier = opIdentifier
	err = factory.kcpCluster.K8sClient().Status().Update(ctx, ipRange)
	assert.Nil(suite.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Invoke the function under test
	err, resCtx := checkGcpOperation(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)
}

func (suite *checkGcpOperationSuite) TestWhenComputeOperationNotComplete() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var opResp *compute.Operation
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, opIdentifier) {
				opResp = &compute.Operation{
					Name:   opIdentifier,
					Status: "PENDING",
				}
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
		b, err := json.Marshal(opResp)
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

	//Get state object with IpRange
	ipRange := gcpIpRange.DeepCopy()
	ipRange.Status.State = client.DeleteAddress
	ipRange.Status.OpIdentifier = opIdentifier
	err = factory.kcpCluster.K8sClient().Status().Update(ctx, ipRange)
	assert.Nil(suite.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Invoke the function under test
	err, resCtx := checkGcpOperation(ctx, state)
	assert.Nil(suite.T(), resCtx)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime), err)
}

func (suite *checkGcpOperationSuite) TestWhenComputeOperationSuccessful() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var opResp *compute.Operation
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, opIdentifier) {
				opResp = &compute.Operation{
					Name:   opIdentifier,
					Status: "DONE",
				}
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
		b, err := json.Marshal(opResp)
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

	//Get state object with IpRange
	ipRange := gcpIpRange.DeepCopy()
	ipRange.Status.State = client.DeleteAddress
	ipRange.Status.OpIdentifier = opIdentifier
	err = factory.kcpCluster.K8sClient().Status().Update(ctx, ipRange)
	assert.Nil(suite.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Invoke the function under test
	err, resCtx := checkGcpOperation(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)
}

func TestCheckGcpOperation(t *testing.T) {
	suite.Run(t, new(checkGcpOperationSuite))
}
