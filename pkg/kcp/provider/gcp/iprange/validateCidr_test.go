package iprange

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type validateCidrSuite struct {
	suite.Suite
	ctx context.Context
}

func TestValidateCidr(t *testing.T) {
	suite.Run(t, new(validateCidrSuite))
}

func (suite *validateCidrSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *validateCidrSuite) TestValidCidr() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object using factory
	state, err := factory.newState(ctx)
	assert.Nil(suite.T(), err)

	//Invoke validateCidr
	err, _ctx := validateCidr(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)

	//Validate state object
	assert.Equal(suite.T(), ipAddr, state.ipAddress)
	assert.Equal(suite.T(), prefix, state.prefix)
}

func (suite *validateCidrSuite) TestInvalidCidr() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Update ipRange object with invalid CIDR
	ipRange := gcpIpRange.DeepCopy()
	ipRange.Spec.Cidr = fmt.Sprintf("%s/%d", ipAddr, prefix-8)
	err = factory.kcpCluster.K8sClient().Update(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Get state object using factory
	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Invoke validateCidr
	err, _ctx := validateCidr(ctx, state)
	assert.Equal(suite.T(), composed.StopAndForget, err)
	assert.Nil(suite.T(), _ctx)

	//Get the updated ipRange object
	ipRange = &cloudcontrolv1beta1.IpRange{}
	err = factory.kcpCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpIpRange.Name, Namespace: gcpIpRange.Namespace}, ipRange)
	assert.Nil(suite.T(), err)

	//Validate the ipRange status
	assert.Equal(suite.T(), 1, len(ipRange.Status.Conditions))
	assert.Equal(suite.T(), cloudcontrolv1beta1.ConditionTypeError, ipRange.Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, ipRange.Status.Conditions[0].Status)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ReasonInvalidCidr, ipRange.Status.Conditions[0].Reason)
}
