package v2

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type updateStateSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *updateStateSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *updateStateSuite) TestStateChangeToReady() {
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
	state.curState = v1beta1.ReadyState

	//Invoke the function under test
	err, _ = updateState(ctx, state)
	assert.Equal(suite.T(), composed.StopAndForget, err)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	// check ready condition in status
	assert.Equal(suite.T(), v1beta1.ReadyState, ipRange.Status.State)
	//assert.Equal(suite.T(), gcpIpRange.Spec.Cidr, ipRange.Status.Cidr) // Cidr is not getting set in updateState action for V2
	assert.Len(suite.T(), ipRange.Status.Conditions, 1)
	assert.Equal(suite.T(), v1beta1.ConditionTypeReady, ipRange.Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, ipRange.Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonReady, ipRange.Status.Conditions[0].Reason)
}

func (suite *updateStateSuite) TestStateChangeToOther() {
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
	state.curState = client.SyncPsaConnection

	//Invoke the function under test
	err, _ = updateState(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeue, err)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	// check state in status
	assert.Equal(suite.T(), client.SyncPsaConnection, ipRange.Status.State)
	assert.Equal(suite.T(), "", ipRange.Status.Cidr)
	assert.Len(suite.T(), ipRange.Status.Conditions, 0)
}

func TestUpdateState(t *testing.T) {
	suite.Run(t, new(updateStateSuite))
}
