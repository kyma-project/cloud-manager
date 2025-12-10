package nfsinstance

import (
	"context"
	"encoding/json"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type checkGcpOperationSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *checkGcpOperationSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *checkGcpOperationSuite) TestCheckGcpOperationNoOperation() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstance()
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx := checkGcpOperation(ctx, testState.State)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)
}

func (s *checkGcpOperationSuite) TestCheckGcpOperationFailedOperation() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var opResp *file.Operation
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/operations/create-operation") {
				opResp = &file.Operation{
					Name: "create-operation",
					Done: true,
					Error: &file.Status{
						Code:    500,
						Message: "Operation failed",
					},
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
	gcpNfsInstance := getGcpNfsInstance()
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "/projects/test-project/locations/us-west1/operations/create-operation")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	err, _ = checkGcpOperation(ctx, testState.State)
	assert.Error(s.T(), err)
	assert.Len(s.T(), testState.ObjAsNfsInstance().Status.Conditions, 1)
	assert.Equal(s.T(), v1beta1.ConditionTypeError, testState.ObjAsNfsInstance().Status.Conditions[0].Type)
	assert.Equal(s.T(), metav1.ConditionTrue, testState.ObjAsNfsInstance().Status.Conditions[0].Status)
	assert.Equal(s.T(), v1beta1.ReasonGcpError, testState.ObjAsNfsInstance().Status.Conditions[0].Reason)
	assert.Equal(s.T(), "", testState.ObjAsNfsInstance().Status.OpIdentifier)
}

func (s *checkGcpOperationSuite) TestCheckGcpOperationNotDoneOperation() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var opResp *file.Operation
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/operations/create-operation") {
				opResp = &file.Operation{
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
	gcpNfsInstance := getGcpNfsInstance()
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "/projects/test-project/locations/us-west1/operations/create-operation")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx := checkGcpOperation(ctx, testState.State)
	assert.Nil(s.T(), resCtx)
	assert.Equal(s.T(), composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), err)
}

func (s *checkGcpOperationSuite) TestCheckGcpOperationSuccessfulOperation() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var opResp *file.Operation
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/operations/create-operation") {
				opResp = &file.Operation{
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
	gcpNfsInstance := getGcpNfsInstance()
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "/projects/test-project/locations/us-west1/operations/create-operation")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx := checkGcpOperation(ctx, testState.State)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)
}
func TestCheckGcpOperation(t *testing.T) {
	suite.Run(t, new(checkGcpOperationSuite))
}
