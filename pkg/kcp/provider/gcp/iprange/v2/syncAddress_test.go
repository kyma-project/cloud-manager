package v2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type syncAddressSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *syncAddressSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *syncAddressSuite) TestUpdate() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)
	state.addressOp = client.MODIFY

	//Invoke the function under test
	err, resCtx := syncAddress(ctx, state)
	assert.Equal(suite.T(), composed.StopAndForget, err)
	assert.Nil(suite.T(), resCtx)
}

func TestSyncAddress(t *testing.T) {
	suite.Run(t, new(syncAddressSuite))
}
