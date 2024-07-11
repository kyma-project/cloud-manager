package v1

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/servicenetworking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (suite *checkGcpOperationSuite) TestWhenSvcNwGetOperationFailed() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, opIdentifier) {
				http.Error(w, "Operation failed", http.StatusInternalServerError)
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

	//Get state object with IpRange
	ipRange := gcpIpRange.DeepCopy()
	ipRange.Status.State = client.DeletePsaConnection
	ipRange.Status.OpIdentifier = opIdentifier
	err = factory.kcpCluster.K8sClient().Status().Update(ctx, ipRange)
	assert.Nil(suite.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Invoke the function under test
	err, _ = checkGcpOperation(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeue, err)

	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	assert.Len(suite.T(), ipRange.Status.Conditions, 1)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ConditionTypeError, ipRange.Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, ipRange.Status.Conditions[0].Status)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ReasonGcpError, ipRange.Status.Conditions[0].Reason)
	assert.Equal(suite.T(), opIdentifier, ipRange.Status.OpIdentifier)
}

func (suite *checkGcpOperationSuite) TestWhenSvcNwOperationFailed() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var opResp *servicenetworking.Operation
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, opIdentifier) {
				opResp = &servicenetworking.Operation{
					Name: "create-operation",
					Done: true,
					Error: &servicenetworking.Status{
						Code:    500,
						Message: "Operation failed",
					},
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
	err, _ = checkGcpOperation(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeue, err)

	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	assert.Len(suite.T(), ipRange.Status.Conditions, 1)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ConditionTypeError, ipRange.Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, ipRange.Status.Conditions[0].Status)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ReasonGcpError, ipRange.Status.Conditions[0].Reason)
	assert.Equal(suite.T(), "", ipRange.Status.OpIdentifier)
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
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime), err)
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

func (suite *checkGcpOperationSuite) TestWhenComputeGetOperationFailed() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, opIdentifier) {
				http.Error(w, "Operation failed", http.StatusInternalServerError)
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

	//Get state object with IpRange
	ipRange := gcpIpRange.DeepCopy()
	ipRange.Status.State = client.DeleteAddress
	ipRange.Status.OpIdentifier = opIdentifier
	err = factory.kcpCluster.K8sClient().Status().Update(ctx, ipRange)
	assert.Nil(suite.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Invoke the function under test
	err, _ = checkGcpOperation(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeue, err)

	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	assert.Len(suite.T(), ipRange.Status.Conditions, 1)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ConditionTypeError, ipRange.Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, ipRange.Status.Conditions[0].Status)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ReasonGcpError, ipRange.Status.Conditions[0].Reason)
	assert.Equal(suite.T(), opIdentifier, ipRange.Status.OpIdentifier)
}

func (suite *checkGcpOperationSuite) TestWhenComputeOperationFailed() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var opResp *compute.Operation
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, opIdentifier) {
				opResp = &compute.Operation{
					Error: &compute.OperationError{
						Errors: []*compute.OperationErrorErrors{{
							Code:    "500",
							Message: "Operation Failed",
						}},
					},
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
	err, _ = checkGcpOperation(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeue, err)

	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	assert.Len(suite.T(), ipRange.Status.Conditions, 1)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ConditionTypeError, ipRange.Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, ipRange.Status.Conditions[0].Status)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ReasonGcpError, ipRange.Status.Conditions[0].Reason)
	assert.Equal(suite.T(), "", ipRange.Status.OpIdentifier)
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
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime), err)
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
