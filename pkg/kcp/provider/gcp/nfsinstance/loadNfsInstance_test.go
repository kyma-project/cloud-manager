package nfsinstance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type loadNfsInstanceSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *loadNfsInstanceSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *loadNfsInstanceSuite) TestLoadNfsInstanceNotFound() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances/cm-test-gcp-nfs-instance") {
				//Return 404
				http.Error(w, "Not Found", http.StatusNotFound)
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
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
	err, resCtx := loadNfsInstance(ctx, testState.State)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)
}

func (s *loadNfsInstanceSuite) TestLoadNfsInstanceOtherErrors() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances/cm-test-gcp-nfs-instance") {
				//Return 500
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
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
	err, _ = loadNfsInstance(ctx, testState.State)
	assert.NotNil(s.T(), err)
	// check error condition in status
	assert.Equal(s.T(), v1beta1.ConditionTypeError, testState.State.ObjAsNfsInstance().Status.Conditions[0].Type)
	assert.Equal(s.T(), metav1.ConditionTrue, testState.State.ObjAsNfsInstance().Status.Conditions[0].Status)
	assert.Equal(s.T(), v1beta1.ReasonGcpError, testState.State.ObjAsNfsInstance().Status.Conditions[0].Reason)
}

func (s *loadNfsInstanceSuite) TestLoadNfsInstanceSuccess() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances/cm-test-gcp-nfs-instance") {
				//Return 200
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"name":"test-gcp-nfs-volume"}`))
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
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
	err, resCtx := loadNfsInstance(ctx, testState.State)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)
	assert.NotNil(s.T(), testState.fsInstance)
}

func (s *loadNfsInstanceSuite) TestLoadNfsInstanceNotFoundWith403() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances/cm-test-gcp-nfs-instance") {
				//Return 403 (invalid location)
				http.Error(w, "Location us-west15-a is not found or access is unauthorized.", http.StatusForbidden)
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
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
	err, resCtx := loadNfsInstance(ctx, testState.State)
	// 403 should be treated as not found - no error returned
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)
	// fsInstance should be nil since resource not found
	assert.Nil(s.T(), testState.fsInstance)
}

func TestLoadNfsInstance(t *testing.T) {
	suite.Run(t, new(loadNfsInstanceSuite))
}
