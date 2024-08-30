package v2

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/servicenetworking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
	"testing"
	"time"
)

type identifyPeeringIpRangesSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *identifyPeeringIpRangesSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *identifyPeeringIpRangesSuite) TestWhenAddressIsNil() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.Fail("unexpected request: " + r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()
	ipRange.Spec.Options.Gcp.Purpose = v1beta1.GcpPurposePSC

	state, err := factory.newStateWith(ctx, ipRange)
	suite.Nil(err)
	state.address = nil

	//Invoke the function under test
	err, resCtx := identifyPeeringIpRanges(ctx, state)
	suite.Nil(err)
	suite.Nil(resCtx)
}

func (suite *identifyPeeringIpRangesSuite) TestWhenDeletingAndSvcConnectionIsNil() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.Fail("unexpected request: " + r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()
	ipRange.SetDeletionTimestamp(&metav1.Time{Time: time.Now()})

	state, err := factory.newStateWith(ctx, ipRange)
	suite.Nil(err)

	state.address = &compute.Address{
		Name: "test-address",
	}
	state.serviceConnection = nil

	//Invoke the function under test
	err, resCtx := identifyPeeringIpRanges(ctx, state)
	suite.Nil(err)
	suite.Nil(resCtx)
	suite.Nil(state.ipRanges)
}

func (suite *identifyPeeringIpRangesSuite) TestWhenNotDeletingAndSvcConnectionIsNil() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.Fail("unexpected request: " + r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()

	state, err := factory.newStateWith(ctx, ipRange)
	suite.Nil(err)

	state.address = &compute.Address{
		Name: "test-address",
	}
	state.serviceConnection = nil

	//Invoke the function under test
	err, resCtx := identifyPeeringIpRanges(ctx, state)
	suite.Nil(err)
	suite.Nil(resCtx)
	suite.Equal(1, len(state.ipRanges))
	suite.Equal(state.address.Name, state.ipRanges[0])
}

func (suite *identifyPeeringIpRangesSuite) TestWhenErrorResponse() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, urlGlobalAddress) {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			} else {
				suite.Fail("unexpected request: " + r.URL.String())
			}
		default:
			suite.Fail("unexpected request: " + r.URL.String())
		}
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()

	state, err := factory.newStateWith(ctx, ipRange)
	suite.Nil(err)

	//set state attributes
	state.address = &compute.Address{
		Name: "test-address",
	}
	state.serviceConnection = &servicenetworking.Connection{
		Network: "test-network",
	}
	state.SetScope(gcpScope)

	//Invoke the function under test
	err, _ = identifyPeeringIpRanges(ctx, state)
	suite.Equal(composed.StopWithRequeue, err)

	//Load updated object
	err = state.LoadObj(ctx)
	suite.Nil(err)
	ipRange = state.ObjAsIpRange()

	// check error condition in status
	suite.Equal(v1beta1.ConditionTypeError, ipRange.Status.Conditions[0].Type)
	suite.Equal(metav1.ConditionTrue, ipRange.Status.Conditions[0].Status)
	suite.Equal(v1beta1.ReasonGcpError, ipRange.Status.Conditions[0].Reason)
}

func (suite *identifyPeeringIpRangesSuite) TestWhenAdding() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp *compute.AddressList

		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, urlGlobalAddress) {
				resp = &compute.AddressList{
					Items: []*compute.Address{
						{
							Name:    "test-address-dns",
							Purpose: string(v1beta1.GcpPurposeDNS),
						},
						{
							Name:    "test-address-0",
							Purpose: string(v1beta1.GcpPurposePSA),
						},
					},
				}
			} else {
				suite.Fail("unexpected request: " + r.URL.String())
			}
		default:
			suite.Fail("unexpected request: " + r.URL.String())
		}
		b, err := json.Marshal(resp)
		if err != nil {
			suite.Fail("unable to marshal request: " + err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			suite.Fail("unable to write to provided ResponseWriter: " + err.Error())
		}
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()

	state, err := factory.newStateWith(ctx, ipRange)
	suite.Nil(err)

	//set state attributes
	state.address = &compute.Address{
		Name: "test-address-1",
	}
	state.serviceConnection = &servicenetworking.Connection{
		Network:               "test-network",
		ReservedPeeringRanges: []string{"test-address-0"},
	}
	state.SetScope(gcpScope)

	//Invoke the function under test
	err, _ = identifyPeeringIpRanges(ctx, state)
	suite.Nil(err)

	suite.Equal(2, len(state.ipRanges))
	suite.Equal(state.address.Name, state.ipRanges[1])
}

func (suite *identifyPeeringIpRangesSuite) TestWhenUpdating() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp *compute.AddressList

		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, urlGlobalAddress) {
				resp = &compute.AddressList{
					Items: []*compute.Address{
						{
							Name:    "test-address-dns",
							Purpose: string(v1beta1.GcpPurposeDNS),
						},
						{
							Name:    "test-address-0",
							Purpose: string(v1beta1.GcpPurposePSA),
						},
					},
				}
			} else {
				suite.Fail("unexpected request: " + r.URL.String())
			}
		default:
			suite.Fail("unexpected request: " + r.URL.String())
		}
		b, err := json.Marshal(resp)
		if err != nil {
			suite.Fail("unable to marshal request: " + err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			suite.Fail("unable to write to provided ResponseWriter: " + err.Error())
		}
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()

	state, err := factory.newStateWith(ctx, ipRange)
	suite.Nil(err)

	//set state attributes
	state.address = &compute.Address{
		Name: "test-address-0",
	}
	state.serviceConnection = &servicenetworking.Connection{
		Network:               "test-network",
		ReservedPeeringRanges: []string{"test-address-0", "invalid-address"},
	}
	state.SetScope(gcpScope)

	//Invoke the function under test
	err, _ = identifyPeeringIpRanges(ctx, state)
	suite.Nil(err)

	suite.Equal(1, len(state.ipRanges))
	suite.Equal(state.address.Name, state.ipRanges[0])
}

func (suite *identifyPeeringIpRangesSuite) TestWhenDeleting() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp *compute.AddressList

		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, urlGlobalAddress) {
				resp = &compute.AddressList{
					Items: []*compute.Address{
						{
							Name:    "test-address-dns",
							Purpose: string(v1beta1.GcpPurposeDNS),
						},
						{
							Name:    "test-address-0",
							Purpose: string(v1beta1.GcpPurposePSA),
						},
					},
				}
			} else {
				suite.Fail("unexpected request: " + r.URL.String())
			}
		default:
			suite.Fail("unexpected request: " + r.URL.String())
		}
		b, err := json.Marshal(resp)
		if err != nil {
			suite.Fail("unable to marshal request: " + err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			suite.Fail("unable to write to provided ResponseWriter: " + err.Error())
		}
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory(fakeHttpServer)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with ipRange
	ipRange := gcpIpRange.DeepCopy()
	ipRange.SetDeletionTimestamp(&metav1.Time{Time: time.Now()})

	state, err := factory.newStateWith(ctx, ipRange)
	suite.Nil(err)

	//set state attributes
	state.address = &compute.Address{
		Name: "test-address-0",
	}
	state.serviceConnection = &servicenetworking.Connection{
		Network:               "test-network",
		ReservedPeeringRanges: []string{"test-address-0"},
	}
	state.SetScope(gcpScope)

	//Invoke the function under test
	err, _ = identifyPeeringIpRanges(ctx, state)
	suite.Nil(err)

	suite.Equal(0, len(state.ipRanges))
}

func TestIdentifyPeeringIpRanges(t *testing.T) {
	suite.Run(t, new(identifyPeeringIpRangesSuite))
}
