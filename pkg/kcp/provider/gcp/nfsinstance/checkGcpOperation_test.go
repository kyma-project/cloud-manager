package nfsinstance

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
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

func (suite *checkGcpOperationSuite) TestCheckGcpOperationNoOperation() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstance()
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx := checkGcpOperation(ctx, testState.State)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)
}

func (suite *checkGcpOperationSuite) TestCheckGcpOperationFailedOperation() {
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
	gcpNfsInstance := getGcpNfsInstance()
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "/projects/test-project/locations/us-west1/operations/create-operation")
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, _ = checkGcpOperation(ctx, testState.State)
	assert.Error(suite.T(), err)
	assert.Len(suite.T(), testState.State.ObjAsNfsInstance().Status.Conditions, 1)
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, testState.State.ObjAsNfsInstance().Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, testState.State.ObjAsNfsInstance().Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonGcpError, testState.State.ObjAsNfsInstance().Status.Conditions[0].Reason)
	assert.Equal(suite.T(), "", testState.State.ObjAsNfsInstance().Status.OpIdentifier)
}

func (suite *checkGcpOperationSuite) TestCheckGcpOperationNotDoneOperation() {
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
	gcpNfsInstance := getGcpNfsInstance()
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "/projects/test-project/locations/us-west1/operations/create-operation")
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx := checkGcpOperation(ctx, testState.State)
	assert.Nil(suite.T(), resCtx)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(testState.gcpConfig.GcpRetryWaitTime), err)
}

func (suite *checkGcpOperationSuite) TestCheckGcpOperationSuccessfulOperation() {
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
	gcpNfsInstance := getGcpNfsInstance()
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "/projects/test-project/locations/us-west1/operations/create-operation")
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx := checkGcpOperation(ctx, testState.State)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)
}
func TestCheckGcpOperation(t *testing.T) {
	suite.Run(t, new(checkGcpOperationSuite))
}
