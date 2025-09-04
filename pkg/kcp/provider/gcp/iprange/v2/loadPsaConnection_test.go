package v2

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/servicenetworking/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type loadPsaConnectionSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *loadPsaConnectionSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *loadPsaConnectionSuite) TestWhenIpRangePurposeIsNotPSA() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp *servicenetworking.ListConnectionsResponse

		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, urlSvcNetworking) {
				resp = &servicenetworking.ListConnectionsResponse{
					Connections: nil,
				}
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
		b, err := json.Marshal(resp)
		if err != nil {
			assert.Fail(s.T(), "unable to marshal request: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(s.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()
	ipRange.Spec.Options.Gcp.Purpose = v1beta1.GcpPurposePSC

	state, err := factory.newStateWith(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Invoke the function under test
	err, resCtx := loadPsaConnection(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(s.T(), err)
	ipRange = state.ObjAsIpRange()

	// check error condition in status
	assert.Len(s.T(), ipRange.Status.Conditions, 0)
}

func (s *loadPsaConnectionSuite) TestWhenSvcConnectionNotFound() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp *servicenetworking.ListConnectionsResponse

		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, urlSvcNetworking) {
				resp = &servicenetworking.ListConnectionsResponse{
					Connections: nil,
				}
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
		b, err := json.Marshal(resp)
		if err != nil {
			assert.Fail(s.T(), "unable to marshal request: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(s.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
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

	//Invoke the function under test
	err, resCtx := loadPsaConnection(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(s.T(), err)
	ipRange = state.ObjAsIpRange()

	// check error condition in status
	assert.Len(s.T(), ipRange.Status.Conditions, 0)
}

func (s *loadPsaConnectionSuite) TestWhenMatchingConnectionFound() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp *servicenetworking.ListConnectionsResponse

		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, urlSvcNetworking) {
				resp = &servicenetworking.ListConnectionsResponse{
					Connections: []*servicenetworking.Connection{
						{
							Network:               "test-vpc",
							Peering:               client.PsaPeeringName,
							ReservedPeeringRanges: nil,
							Service:               client.ServiceNetworkingServiceConnectionName,
						},
					},
				}
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
		b, err := json.Marshal(resp)
		if err != nil {
			assert.Fail(s.T(), "unable to marshal request: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(s.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
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

	//Invoke the function under test
	err, resCtx := loadPsaConnection(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)

	//Load updated object
	err = state.LoadObj(ctx)
	assert.Nil(s.T(), err)
	ipRange = state.ObjAsIpRange()

	// check error condition in status
	assert.NotNil(s.T(), state.serviceConnection)
	assert.Len(s.T(), ipRange.Status.Conditions, 0)
}

func TestLoadPsaConnection(t *testing.T) {
	suite.Run(t, new(loadPsaConnectionSuite))
}
