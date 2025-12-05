package v2

import (
	"context"
	"encoding/json"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/servicenetworking/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type checkGcpOperationSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *checkGcpOperationSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *checkGcpOperationSuite) TestWhenOpIdentifierIsNil() {
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
	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Invoke the function under test
	err, resCtx := checkGcpOperation(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)
}

func (s *checkGcpOperationSuite) TestWhenSvcNwOperationNotComplete() {
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
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
		b, err := json.Marshal(opResp)
		if err != nil {
			assert.Fail(s.T(), "unable to marshal request: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(s.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with IpRange
	ipRange := gcpIpRange.DeepCopy()
	ipRange.Status.State = gcpclient.DeletePsaConnection
	ipRange.Status.OpIdentifier = opIdentifier
	err = factory.kcpCluster.K8sClient().Status().Update(ctx, ipRange)
	assert.Nil(s.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Invoke the function under test
	err, resCtx := checkGcpOperation(ctx, state)
	assert.Nil(s.T(), resCtx)
	assert.Equal(s.T(), composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), err)
}

func (s *checkGcpOperationSuite) TestWhenSvcNwOperationSuccessful() {
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
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
		b, err := json.Marshal(opResp)
		if err != nil {
			assert.Fail(s.T(), "unable to marshal request: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(s.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with IpRange
	ipRange := gcpIpRange.DeepCopy()
	ipRange.Status.State = gcpclient.DeletePsaConnection
	ipRange.Status.OpIdentifier = opIdentifier
	err = factory.kcpCluster.K8sClient().Status().Update(ctx, ipRange)
	assert.Nil(s.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Invoke the function under test
	err, resCtx := checkGcpOperation(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)
}

func (s *checkGcpOperationSuite) TestWhenComputeOperationNotComplete() {
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
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
		b, err := json.Marshal(opResp)
		if err != nil {
			assert.Fail(s.T(), "unable to marshal request: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(s.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with IpRange
	ipRange := gcpIpRange.DeepCopy()
	ipRange.Status.State = gcpclient.DeleteAddress
	ipRange.Status.OpIdentifier = opIdentifier
	err = factory.kcpCluster.K8sClient().Status().Update(ctx, ipRange)
	assert.Nil(s.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Invoke the function under test
	err, resCtx := checkGcpOperation(ctx, state)
	assert.Nil(s.T(), resCtx)
	assert.Equal(s.T(), composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), err)
}

func (s *checkGcpOperationSuite) TestWhenComputeOperationSuccessful() {
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
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
		b, err := json.Marshal(opResp)
		if err != nil {
			assert.Fail(s.T(), "unable to marshal request: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(s.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with IpRange
	ipRange := gcpIpRange.DeepCopy()
	ipRange.Status.State = gcpclient.DeleteAddress
	ipRange.Status.OpIdentifier = opIdentifier
	err = factory.kcpCluster.K8sClient().Status().Update(ctx, ipRange)
	assert.Nil(s.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Invoke the function under test
	err, resCtx := checkGcpOperation(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)
}

func TestCheckGcpOperation(t *testing.T) {
	suite.Run(t, new(checkGcpOperationSuite))
}
