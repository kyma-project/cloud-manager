package v2

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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
	obj := gcpIpRange.DeepCopy()
	obj.Status.Cidr = obj.Spec.Cidr
	state, err := factory.newStateWith(ctx, obj)
	assert.Nil(suite.T(), err)

	//Invoke validateCidr
	err, _ctx := validateCidr(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)

	//Validate state object
	assert.Equal(suite.T(), ipAddr, state.ipAddress)
	assert.Equal(suite.T(), prefix, state.prefix)
}
