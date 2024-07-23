package v2

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	composed "github.com/kyma-project/cloud-manager/pkg/composed"
	ipRangeClient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	spy "github.com/kyma-project/cloud-manager/pkg/testinfra/clientspy"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type testState struct {
	focal.State
}

func (s *testState) ObjAsIpRange() *cloudcontrolv1beta1.IpRange {
	return s.Obj().(*cloudcontrolv1beta1.IpRange)
}

type computeClientStubUtils interface {
	ReturnOnFirstCall(address *compute.Address, err error)
	ReturnOnSecondCall(address *compute.Address, err error)
}

type computeClientStub struct {
	firstCallAddress  *compute.Address
	firstCallErr      error
	secondCallAddress *compute.Address
	secondCallErr     error
	call              int32
	mutex             *sync.Mutex
}

func (c *computeClientStub) ReturnOnFirstCall(address *compute.Address, err error) {
	c.firstCallAddress = address
	c.firstCallErr = err
}

func (c *computeClientStub) ReturnOnSecondCall(address *compute.Address, err error) {
	c.secondCallAddress = address
	c.secondCallErr = err
}

func (c *computeClientStub) CreatePscIpRange(ctx context.Context, projectId string, vpcName string, name string, description string, address string, prefixLength int64) (*compute.Operation, error) {
	panic("unimplemented")
}

func (c *computeClientStub) DeleteIpRange(ctx context.Context, projectId string, name string) (*compute.Operation, error) {
	panic("unimplemented")
}

func (c *computeClientStub) GetGlobalOperation(ctx context.Context, projectId string, operationName string) (*compute.Operation, error) {
	panic("unimplemented")
}

func (c *computeClientStub) GetIpRange(ctx context.Context, projectId string, name string) (*compute.Address, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.call = c.call + 1
	if c.call == 1 {
		return c.firstCallAddress, c.firstCallErr
	}
	if c.call == 2 {
		return c.secondCallAddress, c.secondCallErr
	}

	panic("unexpected call")
}

func (c *computeClientStub) ListGlobalAddresses(ctx context.Context, projectId string, vpc string) (*compute.AddressList, error) {
	panic("unimplemented")
}

func newComputeClientStub() ipRangeClient.ComputeClient {
	return &computeClientStub{
		call:  0,
		mutex: &sync.Mutex{},
	}
}

func TestLoadAddress(t *testing.T) {

	t.Run("loadAddress", func(t *testing.T) {

		notFoundError := &googleapi.Error{Code: 404}
		var ipRange *cloudcontrolv1beta1.IpRange
		var state *State
		var scope *cloudcontrolv1beta1.Scope
		var k8sClient client.WithWatch
		var address *compute.Address
		var computeClient ipRangeClient.ComputeClient

		createEmptyGcpIpRangeState := func(k8sClient client.WithWatch, gcpNfsVolume *cloudcontrolv1beta1.IpRange) *State {
			cluster := composed.NewStateCluster(k8sClient, k8sClient, nil, k8sClient.Scheme())

			focalState := focal.NewStateFactory().NewState(
				composed.NewStateFactory(cluster).NewState(types.NamespacedName{}, gcpNfsVolume),
			)
			focalState.SetScope(scope)

			return &State{
				State:         &testState{State: focalState},
				computeClient: computeClient,
			}
		}

		setupTest := func() {
			scope = &cloudcontrolv1beta1.Scope{
				ObjectMeta: v1.ObjectMeta{
					Name:      "gcp-test",
					Namespace: "kcp-system",
				},
				Spec: cloudcontrolv1beta1.ScopeSpec{
					KymaName:  "8ed7960c-7596-4039-8143-485df5312725",
					ShootName: "a34d2ac1-81f5-4dd1-9288-9cfac0151c4f",
					Region:    "us-east-1",
					Provider:  cloudcontrolv1beta1.ProviderGCP,
					Scope: cloudcontrolv1beta1.ScopeInfo{
						Gcp: &cloudcontrolv1beta1.GcpScope{
							Project:    "non-existant-test-proj",
							VpcNetwork: "test-vpc",
						},
					},
				},
			}

			ipRange = &cloudcontrolv1beta1.IpRange{
				ObjectMeta: v1.ObjectMeta{
					Name:      "3abd9b3c-1125-42e2-9157-72a3758b1b59",
					Namespace: "kcp-system",
				},
				Spec: cloudcontrolv1beta1.IpRangeSpec{
					RemoteRef: cloudcontrolv1beta1.RemoteRef{
						Name: "default",
					},
					Scope: cloudcontrolv1beta1.ScopeRef{
						Name: scope.GetName(),
					},
				},
			}

			scheme := runtime.NewScheme()
			utilruntime.Must(clientgoscheme.AddToScheme(scheme))
			utilruntime.Must(cloudcontrolv1beta1.AddToScheme(scheme))
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			k8sClient = spy.NewClientSpy(fakeClient)

			computeClient = newComputeClientStub()

			state = createEmptyGcpIpRangeState(k8sClient, ipRange)
		}

		t.Run("Should: load IpRange (new name)", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			address = &compute.Address{
				Network: fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork),
			}
			computeClient.(computeClientStubUtils).ReturnOnFirstCall(address, nil)

			err, res := loadAddress(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
			assert.Equal(t, state.address, address)
		})

		t.Run("Should: fallback and load IpRange with old name", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			address = &compute.Address{
				Network: fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork),
			}
			computeClient.(computeClientStubUtils).ReturnOnFirstCall(nil, notFoundError)
			computeClient.(computeClientStubUtils).ReturnOnSecondCall(address, nil)

			err, res := loadAddress(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
			assert.Equal(t, state.address, address)
		})

		t.Run("Should: skip loaded fallback IpRange if it belongs to other VPC", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			address = &compute.Address{
				Network: "some-other-network",
			}
			computeClient.(computeClientStubUtils).ReturnOnFirstCall(nil, notFoundError)
			computeClient.(computeClientStubUtils).ReturnOnSecondCall(address, nil)

			err, res := loadAddress(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
			assert.Nilf(t, state.address, "address should be unset")
		})

		t.Run("Should: do nothing if IpRange doesnt exist (new or old name)", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			computeClient.(computeClientStubUtils).ReturnOnFirstCall(nil, notFoundError)
			computeClient.(computeClientStubUtils).ReturnOnSecondCall(nil, notFoundError)

			err, res := loadAddress(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
			assert.Nilf(t, state.address, "address should be unset")
		})

		t.Run("Should: set error if obtained IpRange (new name) belongs to another VPC", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			address = &compute.Address{
				Network: fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/another-%s", scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork), // note: VPC name is different, but has same suffix
			}
			computeClient.(computeClientStubUtils).ReturnOnFirstCall(address, nil)

			err, res := loadAddress(ctx, state)

			assert.NotNil(t, res, "should return non-nil res")
			assert.NotNil(t, err, "should return non-nil err")
			assert.Nilf(t, state.address, "address should be unset")
		})

		t.Run("Should: set error if its unable to get IpRange (new name)", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			computeClient.(computeClientStubUtils).ReturnOnFirstCall(nil, errors.New("Random error"))

			err, res := loadAddress(ctx, state)

			assert.NotNil(t, res, "should return non-nil res")
			assert.NotNil(t, err, "should return non-nil err")
			assert.Nilf(t, state.address, "address should be unset")
		})

		t.Run("Should: set error if its unable to get IpRange (fallback name)", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			computeClient.(computeClientStubUtils).ReturnOnFirstCall(nil, notFoundError)
			computeClient.(computeClientStubUtils).ReturnOnSecondCall(nil, errors.New("Random error"))

			err, res := loadAddress(ctx, state)

			assert.NotNil(t, res, "should return non-nil res")
			assert.NotNil(t, err, "should return non-nil err")
			assert.Nilf(t, state.address, "address should be unset")
		})
	})
}
