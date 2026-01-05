package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type validateAlwaysSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *validateAlwaysSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *validateAlwaysSuite) TestValidateAlwaysHappy() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// no-op
	}))
	gcpNfsInstanceWithoutStatus := getGcpNfsInstanceWithoutStatus()
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstanceWithoutStatus)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	testState, err := factory.newStateWith(ctx, gcpNfsInstanceWithoutStatus, "")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	err, _ = validateAlways(ctx, testState.State)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, len(testState.ObjAsNfsInstance().Status.Conditions))
}

func (s *validateAlwaysSuite) TestValidateAlwaysInvalidCapacity() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// no-op
	}))
	gcpNfsInstanceWithoutStatus := getGcpNfsInstanceWithoutStatus()
	gcpNfsInstanceWithoutStatus.Spec.Instance.Gcp.CapacityGb = 100
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstanceWithoutStatus)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	testState, err := factory.newStateWith(ctx, gcpNfsInstanceWithoutStatus, "")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	err, _ = validateAlways(ctx, testState.State)
	assert.Error(s.T(), err)
	assert.Len(s.T(), testState.ObjAsNfsInstance().Status.Conditions, 1)
	assert.Equal(s.T(), v1beta1.ConditionTypeError, testState.ObjAsNfsInstance().Status.Conditions[0].Type)
	assert.Equal(s.T(), metav1.ConditionTrue, testState.ObjAsNfsInstance().Status.Conditions[0].Status)
	assert.Equal(s.T(), v1beta1.ReasonValidationFailed, testState.ObjAsNfsInstance().Status.Conditions[0].Reason)
	assert.Equal(s.T(), 1, len(testState.validations))
}

func TestValidateAlways(t *testing.T) {
	suite.Run(t, new(validateAlwaysSuite))
}
