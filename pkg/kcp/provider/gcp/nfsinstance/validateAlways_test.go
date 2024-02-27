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
		return // no-op
	}))

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	testState, err := factory.newStateWith(ctx, getGcpNfsInstanceWithoutStatus(), "")
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, _ = validateAlways(ctx, testState.State)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(testState.State.ObjAsNfsInstance().Status.Conditions))
}

func (suite *validateAlwaysSuite) TestValidateAlwaysInvalidCapacity() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		return // no-op
	}))
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	instance := getGcpNfsInstanceWithoutStatus()
	instance.Spec.Instance.Gcp.CapacityGb = 100
	testState, err := factory.newStateWith(ctx, instance, "")
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, _ = validateAlways(ctx, testState.State)
	assert.Error(suite.T(), err)
	assert.Len(suite.T(), testState.State.ObjAsNfsInstance().Status.Conditions, 1)
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, testState.State.ObjAsNfsInstance().Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, testState.State.ObjAsNfsInstance().Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonValidationFailed, testState.State.ObjAsNfsInstance().Status.Conditions[0].Reason)
	assert.Equal(suite.T(), 1, len(testState.State.validations))
}

func TestValidateAlways(t *testing.T) {
	suite.Run(t, new(validateAlwaysSuite))
}
