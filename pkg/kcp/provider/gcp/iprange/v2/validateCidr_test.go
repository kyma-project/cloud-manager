package v2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type validateCidrSuite struct {
	suite.Suite
	ctx context.Context
}

func TestValidateCidr(t *testing.T) {
	suite.Run(t, new(validateCidrSuite))
}

func (s *validateCidrSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *validateCidrSuite) TestValidCidr() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object using factory
	obj := gcpIpRange.DeepCopy()
	obj.Status.Cidr = obj.Spec.Cidr
	state, err := factory.newStateWith(ctx, obj)
	assert.Nil(s.T(), err)

	//Invoke validateCidr
	err, _ctx := validateCidr(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)

	//Validate state object
	assert.Equal(s.T(), ipAddr, state.ipAddress)
	assert.Equal(s.T(), prefix, state.prefix)
}
