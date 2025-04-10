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
	"testing"
)

type validateAlwaysSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *validateAlwaysSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *validateAlwaysSuite) TestValidateAlwaysHappy() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// no-op
	}))
	gcpNfsInstanceWithoutStatus := getGcpNfsInstanceWithoutStatus()
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstanceWithoutStatus)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	testState, err := factory.newStateWith(ctx, gcpNfsInstanceWithoutStatus, "")
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, _ = validateAlways(ctx, testState.State)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(testState.ObjAsNfsInstance().Status.Conditions))
}

func (suite *validateAlwaysSuite) TestValidateAlwaysInvalidCapacity() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// no-op
	}))
	gcpNfsInstanceWithoutStatus := getGcpNfsInstanceWithoutStatus()
	gcpNfsInstanceWithoutStatus.Spec.Instance.Gcp.CapacityGb = 100
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstanceWithoutStatus)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	testState, err := factory.newStateWith(ctx, gcpNfsInstanceWithoutStatus, "")
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, _ = validateAlways(ctx, testState.State)
	assert.Error(suite.T(), err)
	assert.Len(suite.T(), testState.ObjAsNfsInstance().Status.Conditions, 1)
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, testState.ObjAsNfsInstance().Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, testState.ObjAsNfsInstance().Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonValidationFailed, testState.ObjAsNfsInstance().Status.Conditions[0].Reason)
	assert.Equal(suite.T(), 1, len(testState.validations))
}

func TestValidateAlways(t *testing.T) {
	suite.Run(t, new(validateAlwaysSuite))
}
