package iprange

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"cloud.google.com/go/compute/apiv1/computepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	spy "github.com/kyma-project/cloud-manager/pkg/testinfra/clientspy"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/googleapi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ iprangetypes.State = &testState{}

type testState struct {
	focal.State
}

func (s *testState) ObjAsIpRange() *cloudcontrolv1beta1.IpRange {
	return s.Obj().(*cloudcontrolv1beta1.IpRange)
}

func (s *testState) Network() *cloudcontrolv1beta1.Network {
	return nil
}

func (s *testState) SetNetwork(network *cloudcontrolv1beta1.Network) {}

func (s *testState) NetworkKey() client.ObjectKey {
	return client.ObjectKey{}
}

func (s *testState) SetNetworkKey(key client.ObjectKey) {}

func (s *testState) IsCloudManagerNetwork() bool {
	return false
}

func (s *testState) SetIsCloudManagerNetwork(v bool) {}

func (s *testState) IsKymaNetwork() bool {
	return false
}

func (s *testState) SetIsKymaNetwork(v bool) {}

func (s *testState) KymaNetwork() *cloudcontrolv1beta1.Network {
	return nil
}

func (s *testState) SetKymaNetwork(network *cloudcontrolv1beta1.Network) {}

func (s *testState) KymaPeering() *cloudcontrolv1beta1.VpcPeering {
	return nil
}

func (s *testState) SetKymaPeering(peering *cloudcontrolv1beta1.VpcPeering) {}

func (s *testState) ExistingCidrRanges() []string {
	return nil
}

func (s *testState) SetExistingCidrRanges(v []string) {}

type computeClientStubUtils interface {
	ReturnOnFirstCall(address *computepb.Address, err error)
	ReturnOnSecondCall(address *computepb.Address, err error)
	GetFirstCallName() string
	GetSecondCallName() string
	GetCallCount() int
}

type computeClientStub struct {
	firstCallAddress  *computepb.Address
	firstCallErr      error
	secondCallAddress *computepb.Address
	secondCallErr     error
	callCount         int
	mutex             *sync.Mutex
	firstCallName     string
	secondCallName    string
}

func (c *computeClientStub) ReturnOnFirstCall(address *computepb.Address, err error) {
	c.firstCallAddress = address
	c.firstCallErr = err
}

func (c *computeClientStub) ReturnOnSecondCall(address *computepb.Address, err error) {
	c.secondCallAddress = address
	c.secondCallErr = err
}

func (c *computeClientStub) GetFirstCallName() string {
	return c.firstCallName
}

func (c *computeClientStub) GetSecondCallName() string {
	return c.secondCallName
}

func (c *computeClientStub) GetCallCount() int {
	return c.callCount
}

func (c *computeClientStub) CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (string, error) {
	panic("unimplemented")
}

func (c *computeClientStub) DeleteIpRange(ctx context.Context, projectId, name string) (string, error) {
	panic("unimplemented")
}

func (c *computeClientStub) GetGlobalOperation(ctx context.Context, projectId, operationName string) (*computepb.Operation, error) {
	panic("unimplemented")
}

func (c *computeClientStub) WaitGlobalOperation(ctx context.Context, projectId, operationName string) error {
	panic("unimplemented")
}

func (c *computeClientStub) GetIpRange(ctx context.Context, projectId, name string) (*computepb.Address, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.callCount = c.callCount + 1
	if c.callCount == 1 {
		c.firstCallName = name
		return c.firstCallAddress, c.firstCallErr
	}
	if c.callCount == 2 {
		c.secondCallName = name
		return c.secondCallAddress, c.secondCallErr
	}

	panic("unexpected call")
}

func (c *computeClientStub) ListGlobalAddresses(ctx context.Context, projectId, vpc string) ([]*computepb.Address, error) {
	panic("unimplemented")
}

func newComputeClientStub() gcpiprangeclient.ComputeClient {
	return &computeClientStub{
		callCount: 0,
		mutex:     &sync.Mutex{},
	}
}

func TestLoadAddress(t *testing.T) {

	t.Run("loadAddress", func(t *testing.T) {

		notFoundError := &googleapi.Error{Code: 404}
		var ipRange *cloudcontrolv1beta1.IpRange
		var state *State
		var scope *cloudcontrolv1beta1.Scope
		var k8sClient client.WithWatch
		var address *computepb.Address
		var computeClient gcpiprangeclient.ComputeClient

		createEmptyGcpIpRangeState := func(k8sClient client.WithWatch, gcpIpRange *cloudcontrolv1beta1.IpRange) *State {
			cluster := composed.NewStateCluster(k8sClient, k8sClient, nil, k8sClient.Scheme())

			focalState := focal.NewStateFactory().NewState(
				composed.NewStateFactory(cluster).NewState(types.NamespacedName{}, gcpIpRange),
			)
			focalState.SetScope(scope)

			return &State{
				State:         &testState{State: focalState},
				computeClient: computeClient,
			}
		}

		setupTest := func() {
			scope = &cloudcontrolv1beta1.Scope{
				ObjectMeta: metav1.ObjectMeta{
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
				ObjectMeta: metav1.ObjectMeta{
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

			fakeClient := fake.NewClientBuilder().
				WithScheme(commonscheme.KcpScheme).
				Build()

			k8sClient = spy.NewClientSpy(fakeClient)

			computeClient = newComputeClientStub()

			state = createEmptyGcpIpRangeState(k8sClient, ipRange)
		}

		t.Run("Should: load IpRange (new name)", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			networkUrl := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork)
			address = &computepb.Address{
				Network: &networkUrl,
			}
			computeClient.(computeClientStubUtils).ReturnOnFirstCall(address, nil)

			err, res := loadAddress(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
			assert.Equal(t, address, state.address)
			assert.Equal(t, 1, computeClient.(computeClientStubUtils).GetCallCount())
			assert.Equal(t, GetIpRangeName(ipRange.Name), computeClient.(computeClientStubUtils).GetFirstCallName())
		})

		t.Run("Should: fallback and load IpRange with old name", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			networkUrl := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork)
			address = &computepb.Address{
				Network: &networkUrl,
			}
			computeClient.(computeClientStubUtils).ReturnOnFirstCall(nil, notFoundError)
			computeClient.(computeClientStubUtils).ReturnOnSecondCall(address, nil)

			err, res := loadAddress(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
			assert.Equal(t, address, state.address)
			assert.Equal(t, 2, computeClient.(computeClientStubUtils).GetCallCount())
			assert.Equal(t, GetIpRangeName(ipRange.Name), computeClient.(computeClientStubUtils).GetFirstCallName())
			assert.Equal(t, ipRange.Spec.RemoteRef.Name, computeClient.(computeClientStubUtils).GetSecondCallName())
		})

		t.Run("Should: skip loaded fallback IpRange if it belongs to other VPC", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			otherNetwork := "some-other-network"
			address = &computepb.Address{
				Network: &otherNetwork,
			}
			computeClient.(computeClientStubUtils).ReturnOnFirstCall(nil, notFoundError)
			computeClient.(computeClientStubUtils).ReturnOnSecondCall(address, nil)

			err, res := loadAddress(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
			assert.Nilf(t, state.address, "address should be unset")
			assert.Equal(t, 2, computeClient.(computeClientStubUtils).GetCallCount())
			assert.Equal(t, GetIpRangeName(ipRange.Name), computeClient.(computeClientStubUtils).GetFirstCallName())
			assert.Equal(t, ipRange.Spec.RemoteRef.Name, computeClient.(computeClientStubUtils).GetSecondCallName())
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
			assert.Equal(t, 2, computeClient.(computeClientStubUtils).GetCallCount())
			assert.Equal(t, GetIpRangeName(ipRange.Name), computeClient.(computeClientStubUtils).GetFirstCallName())
			assert.Equal(t, ipRange.Spec.RemoteRef.Name, computeClient.(computeClientStubUtils).GetSecondCallName())
		})

		t.Run("Should: set error if obtained IpRange (new name) belongs to another VPC", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// VPC name is different but has same suffix
			networkUrl := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/another-%s", scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork)
			address = &computepb.Address{
				Network: &networkUrl,
			}
			computeClient.(computeClientStubUtils).ReturnOnFirstCall(address, nil)

			err, res := loadAddress(ctx, state)

			assert.NotNil(t, res, "should return non-nil res")
			assert.NotNil(t, err, "should return non-nil err")
			assert.Nilf(t, state.address, "address should be unset")
			assert.Equal(t, 1, computeClient.(computeClientStubUtils).GetCallCount())
			assert.Equal(t, GetIpRangeName(ipRange.Name), computeClient.(computeClientStubUtils).GetFirstCallName())
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
			assert.Equal(t, 1, computeClient.(computeClientStubUtils).GetCallCount())
			assert.Equal(t, GetIpRangeName(ipRange.Name), computeClient.(computeClientStubUtils).GetFirstCallName())
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
			assert.Equal(t, 2, computeClient.(computeClientStubUtils).GetCallCount())
			assert.Equal(t, GetIpRangeName(ipRange.Name), computeClient.(computeClientStubUtils).GetFirstCallName())
			assert.Equal(t, ipRange.Spec.RemoteRef.Name, computeClient.(computeClientStubUtils).GetSecondCallName())
		})
	})
}
