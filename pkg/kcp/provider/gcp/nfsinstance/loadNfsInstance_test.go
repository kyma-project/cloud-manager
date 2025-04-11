package nfsinstance

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
	"testing"
)

type loadNfsInstanceSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *loadNfsInstanceSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *loadNfsInstanceSuite) TestLoadNfsInstanceNotFound() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances/cm-test-gcp-nfs-instance") {
				//Return 404
				http.Error(w, "Not Found", http.StatusNotFound)
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
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
	err, resCtx := loadNfsInstance(ctx, testState.State)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)
}

func (suite *loadNfsInstanceSuite) TestLoadNfsInstanceOtherErrors() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances/cm-test-gcp-nfs-instance") {
				//Return 500
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
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
	err, _ = loadNfsInstance(ctx, testState.State)
	assert.NotNil(suite.T(), err)
	// check error condition in status
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, testState.State.ObjAsNfsInstance().Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, testState.State.ObjAsNfsInstance().Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonGcpError, testState.State.ObjAsNfsInstance().Status.Conditions[0].Reason)
}

func (suite *loadNfsInstanceSuite) TestLoadNfsInstanceSuccess() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances/cm-test-gcp-nfs-instance") {
				//Return 200
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"name":"test-gcp-nfs-volume"}`))
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
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
	err, resCtx := loadNfsInstance(ctx, testState.State)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)
	assert.NotNil(suite.T(), testState.fsInstance)
}
func TestLoadNfsInstance(t *testing.T) {
	suite.Run(t, new(loadNfsInstanceSuite))
}
