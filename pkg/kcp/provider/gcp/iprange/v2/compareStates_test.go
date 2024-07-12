package v2

import (
	"context"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/servicenetworking/v1"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type compareStatesSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *compareStatesSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *compareStatesSuite) TestWhenDeletingAndNoAddressOrConnectionExists() {
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
	err = factory.kcpCluster.K8sClient().Delete(ctx, ipRange)
	assert.Nil(suite.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)

	////Invoke the function under test
	err, resCtx := compareStates(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)

	//Validate state attributes
	assert.Equal(suite.T(), client.Deleted, state.curState)
	assert.Equal(suite.T(), client.NONE, state.connectionOp)
	assert.Equal(suite.T(), client.NONE, state.addressOp)
	assert.True(suite.T(), state.inSync)
}

func (suite *compareStatesSuite) TestWhenDeletingAndAddressExistsAndNoConnectionExists() {
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
	err = factory.kcpCluster.K8sClient().Delete(ctx, ipRange)
	assert.Nil(suite.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)

	//Set Address
	address := &compute.Address{
		Address: ipAddr,
		Name:    gcpIpRange.Spec.RemoteRef.Name,
	}
	state.address = address

	////Invoke the function under test
	err, resCtx := compareStates(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)

	//Validate state attributes
	assert.Equal(suite.T(), client.DeleteAddress, state.curState)
	assert.Equal(suite.T(), client.NONE, state.connectionOp)
	assert.Equal(suite.T(), client.DELETE, state.addressOp)
	assert.False(suite.T(), state.inSync)
}

func (suite *compareStatesSuite) TestWhenDeletingAndNoAddressExistsAndConnectionExists() {
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
	err = factory.kcpCluster.K8sClient().Delete(ctx, ipRange)
	assert.Nil(suite.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	//Set Connection
	connection := &servicenetworking.Connection{
		ReservedPeeringRanges: []string{ipRange.Spec.RemoteRef.Name},
	}
	state.serviceConnection = connection

	////Invoke the function under test
	err, resCtx := compareStates(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)

	//Validate state attributes
	assert.Equal(suite.T(), client.DeletePsaConnection, state.curState)
	assert.Equal(suite.T(), client.DELETE, state.connectionOp)
	assert.Equal(suite.T(), client.NONE, state.addressOp)
	assert.False(suite.T(), state.inSync)
}

func (suite *compareStatesSuite) TestWhenDeletingAndBothAddressAndConnectionExists() {
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
	err = factory.kcpCluster.K8sClient().Delete(ctx, ipRange)
	assert.Nil(suite.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(suite.T(), err)

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	//Set Address
	address := &compute.Address{
		Address: ipAddr,
		Name:    gcpIpRange.Spec.RemoteRef.Name,
	}
	state.address = address

	//Set Connection
	connection := &servicenetworking.Connection{
		ReservedPeeringRanges: []string{ipRange.Spec.RemoteRef.Name, "test"},
	}
	state.serviceConnection = connection

	////Invoke the function under test
	err, resCtx := compareStates(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)

	//Validate state attributes
	assert.Equal(suite.T(), client.DeletePsaConnection, state.curState)
	assert.Equal(suite.T(), client.MODIFY, state.connectionOp)
	assert.Equal(suite.T(), client.DELETE, state.addressOp)
	assert.False(suite.T(), state.inSync)
}

func (suite *compareStatesSuite) TestWhenNotDeleting_NoAddressOrConnectionExists() {
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

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)

	////Invoke the function under test
	err, resCtx := compareStates(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)

	//Validate state attributes
	assert.Equal(suite.T(), client.SyncAddress, state.curState)
	assert.Equal(suite.T(), client.ADD, state.connectionOp)
	assert.Equal(suite.T(), client.ADD, state.addressOp)
	assert.False(suite.T(), state.inSync)
}

func (suite *compareStatesSuite) TestWhenNotDeleting_AddressExistsAndNoConnectionExists() {
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

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	//Set Address
	address := &compute.Address{
		Address:      ipAddr,
		PrefixLength: int64(prefix),
		Name:         gcpIpRange.Spec.RemoteRef.Name,
		Network:      state.Scope().Spec.Scope.Gcp.VpcNetwork,
	}
	state.address = address
	state.ipAddress = ipAddr
	state.prefix = prefix

	////Invoke the function under test
	err, resCtx := compareStates(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)

	//Validate state attributes
	assert.Equal(suite.T(), client.SyncPsaConnection, state.curState)
	assert.Equal(suite.T(), client.ADD, state.connectionOp)
	assert.Equal(suite.T(), client.NONE, state.addressOp)
	assert.False(suite.T(), state.inSync)
}

func (suite *compareStatesSuite) TestWhenNotDeleting_AddressNotMatches() {
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

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	//Set Address
	address := &compute.Address{
		Address:      ipAddr,
		PrefixLength: int64(prefix),
		Name:         gcpIpRange.Spec.RemoteRef.Name,
	}
	state.address = address
	state.ipAddress = ipAddr
	state.prefix = prefix + 1

	////Invoke the function under test
	err, resCtx := compareStates(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)

	//Validate state attributes
	assert.Equal(suite.T(), client.SyncAddress, state.curState)
	assert.Equal(suite.T(), client.ADD, state.connectionOp)
	assert.Equal(suite.T(), client.MODIFY, state.addressOp)
	assert.False(suite.T(), state.inSync)
}

func (suite *compareStatesSuite) TestWhenNotDeleting_AddressExistsAndConnectionNotInclusive() {
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

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)
	ipRange = state.ObjAsIpRange()

	//Set Address
	address := &compute.Address{
		Address:      ipAddr,
		PrefixLength: int64(prefix),
		Name:         gcpIpRange.Spec.RemoteRef.Name,
		Network:      state.Scope().Spec.Scope.Gcp.VpcNetwork,
	}
	state.address = address
	state.ipAddress = ipAddr
	state.prefix = prefix

	//Set Connection
	connection := &servicenetworking.Connection{
		ReservedPeeringRanges: []string{"test"},
	}
	state.serviceConnection = connection

	//Invoke the function under test
	err, resCtx := compareStates(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)

	//Validate state attributes
	assert.Equal(suite.T(), client.SyncPsaConnection, state.curState)
	assert.Equal(suite.T(), client.MODIFY, state.connectionOp)
	assert.Equal(suite.T(), client.NONE, state.addressOp)
	assert.False(suite.T(), state.inSync)
}

func (suite *compareStatesSuite) TestWhenNotDeleting_BothAddressAndConnectionExists() {
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

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(suite.T(), err)

	//Set Address
	address := &compute.Address{
		Address:      ipAddr,
		PrefixLength: int64(prefix),
		Name:         gcpIpRange.Spec.RemoteRef.Name,
		Network:      state.Scope().Spec.Scope.Gcp.VpcNetwork,
	}
	state.address = address
	state.ipAddress = ipAddr
	state.prefix = prefix

	//Set Connection
	connection := &servicenetworking.Connection{
		ReservedPeeringRanges: []string{gcpIpRange.Spec.RemoteRef.Name},
	}
	state.serviceConnection = connection

	////Invoke the function under test
	err, resCtx := compareStates(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)

	//Validate state attributes
	assert.Equal(suite.T(), cloudcontrolv1beta1.ReadyState, state.curState)
	assert.Equal(suite.T(), client.NONE, state.connectionOp)
	assert.Equal(suite.T(), client.NONE, state.addressOp)
	assert.True(suite.T(), state.inSync)
}

func TestCompareStates(t *testing.T) {
	suite.Run(t, new(compareStatesSuite))
}
