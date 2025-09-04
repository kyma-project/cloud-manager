package v2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/servicenetworking/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type compareStatesSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *compareStatesSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *compareStatesSuite) TestWhenDeletingAndNoAddressOrConnectionExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()
	err = factory.kcpCluster.K8sClient().Delete(ctx, ipRange)
	assert.Nil(s.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(s.T(), err)

	////Invoke the function under test
	err, resCtx := compareStates(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)

	//Validate state attributes
	assert.Equal(s.T(), client.Deleted, state.curState)
	assert.Equal(s.T(), client.NONE, state.connectionOp)
	assert.Equal(s.T(), client.NONE, state.addressOp)
	assert.True(s.T(), state.inSync)
}

func (s *compareStatesSuite) TestWhenDeletingAndAddressExistsAndNoConnectionExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()
	err = factory.kcpCluster.K8sClient().Delete(ctx, ipRange)
	assert.Nil(s.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(s.T(), err)

	//Set Address
	address := &compute.Address{
		Address: ipAddr,
		Name:    GetIpRangeName(gcpIpRange.Name),
	}
	state.address = address

	////Invoke the function under test
	err, resCtx := compareStates(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)

	//Validate state attributes
	assert.Equal(s.T(), client.DeleteAddress, state.curState)
	assert.Equal(s.T(), client.NONE, state.connectionOp)
	assert.Equal(s.T(), client.DELETE, state.addressOp)
	assert.False(s.T(), state.inSync)
}

func (s *compareStatesSuite) TestWhenDeletingAndNoAddressExistsAndConnectionExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()
	err = factory.kcpCluster.K8sClient().Delete(ctx, ipRange)
	assert.Nil(s.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(s.T(), err)
	ipRange = state.ObjAsIpRange()

	//Set Connection
	connection := &servicenetworking.Connection{
		ReservedPeeringRanges: []string{GetIpRangeName(ipRange.Name)},
	}
	state.serviceConnection = connection

	////Invoke the function under test
	err, resCtx := compareStates(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)

	//Validate state attributes
	assert.Equal(s.T(), client.Deleted, state.curState)
	assert.Equal(s.T(), client.NONE, state.connectionOp)
	assert.Equal(s.T(), client.NONE, state.addressOp)
	assert.True(s.T(), state.inSync)
}

func (s *compareStatesSuite) TestWhenDeletingAndBothAddressAndConnectionExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()
	err = factory.kcpCluster.K8sClient().Delete(ctx, ipRange)
	assert.Nil(s.T(), err)

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(s.T(), err)
	ipRange = state.ObjAsIpRange()

	//Set Address
	address := &compute.Address{
		Address: ipAddr,
		Name:    GetIpRangeName(gcpIpRange.Name),
	}
	state.address = address

	//Set Connection
	connection := &servicenetworking.Connection{
		ReservedPeeringRanges: []string{GetIpRangeName(ipRange.Name), "test"},
	}
	state.serviceConnection = connection

	////Invoke the function under test
	err, resCtx := compareStates(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)

	//Validate state attributes
	assert.Equal(s.T(), client.DeletePsaConnection, state.curState)
	assert.Equal(s.T(), client.MODIFY, state.connectionOp)
	assert.Equal(s.T(), client.DELETE, state.addressOp)
	assert.False(s.T(), state.inSync)
}

func (s *compareStatesSuite) TestWhenNotDeleting_NoAddressOrConnectionExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(s.T(), err)

	////Invoke the function under test
	err, resCtx := compareStates(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)

	//Validate state attributes
	assert.Equal(s.T(), client.SyncAddress, state.curState)
	assert.Equal(s.T(), client.NONE, state.connectionOp)
	assert.Equal(s.T(), client.ADD, state.addressOp)
	assert.False(s.T(), state.inSync)
}

func (s *compareStatesSuite) TestWhenNotDeleting_AddressExistsAndNoConnectionExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(s.T(), err)

	//Set Address
	address := &compute.Address{
		Address:      ipAddr,
		PrefixLength: int64(prefix),
		Name:         GetIpRangeName(gcpIpRange.Name),
		Network:      state.Scope().Spec.Scope.Gcp.VpcNetwork,
	}
	state.address = address
	state.ipAddress = ipAddr
	state.prefix = prefix

	////Invoke the function under test
	err, resCtx := compareStates(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)

	//Validate state attributes
	assert.Equal(s.T(), client.SyncPsaConnection, state.curState)
	assert.Equal(s.T(), client.ADD, state.connectionOp)
	assert.Equal(s.T(), client.NONE, state.addressOp)
	assert.False(s.T(), state.inSync)
}

func (s *compareStatesSuite) TestWhenNotDeleting_AddressNotMatches() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(s.T(), err)

	//Set Address
	address := &compute.Address{
		Address:      ipAddr,
		PrefixLength: int64(prefix),
		Name:         GetIpRangeName(gcpIpRange.Name),
	}
	state.address = address
	state.ipAddress = ipAddr
	state.prefix = prefix + 1

	////Invoke the function under test
	err, resCtx := compareStates(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)

	//Validate state attributes
	assert.Equal(s.T(), client.SyncAddress, state.curState)
	assert.Equal(s.T(), client.ADD, state.connectionOp)
	assert.Equal(s.T(), client.MODIFY, state.addressOp)
	assert.False(s.T(), state.inSync)
}

func (s *compareStatesSuite) TestWhenNotDeleting_AddressExistsAndConnectionNotInclusive() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(s.T(), err)

	//Set Address
	address := &compute.Address{
		Address:      ipAddr,
		PrefixLength: int64(prefix),
		Name:         GetIpRangeName(gcpIpRange.Name),
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
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)

	//Validate state attributes
	assert.Equal(s.T(), client.SyncPsaConnection, state.curState)
	assert.Equal(s.T(), client.MODIFY, state.connectionOp)
	assert.Equal(s.T(), client.NONE, state.addressOp)
	assert.False(s.T(), state.inSync)
}

func (s *compareStatesSuite) TestWhenNotDeleting_BothAddressAndConnectionExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Get the updated object
	err = state.LoadObj(ctx)
	assert.Nil(s.T(), err)

	//Set Address
	address := &compute.Address{
		Address:      ipAddr,
		PrefixLength: int64(prefix),
		Name:         GetIpRangeName(gcpIpRange.Name),
		Network:      state.Scope().Spec.Scope.Gcp.VpcNetwork,
	}
	state.address = address
	state.ipAddress = ipAddr
	state.prefix = prefix

	//Set Connection
	connection := &servicenetworking.Connection{
		ReservedPeeringRanges: []string{GetIpRangeName(gcpIpRange.Name)},
	}
	state.serviceConnection = connection

	////Invoke the function under test
	err, resCtx := compareStates(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)

	//Validate state attributes
	assert.Equal(s.T(), cloudcontrolv1beta1.StateReady, state.curState)
	assert.Equal(s.T(), client.NONE, state.connectionOp)
	assert.Equal(s.T(), client.NONE, state.addressOp)
	assert.True(s.T(), state.inSync)
}

func TestCompareStates(t *testing.T) {
	suite.Run(t, new(compareStatesSuite))
}
