package v2

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/api"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"

	"net/http/httptest"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/servicenetworking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newFakeComputeClientProvider(fakeHttpServer *httptest.Server) client.ClientProvider[gcpiprangeclient.ComputeClient] {
	return client.NewCachedClientProvider(
		func(ctx context.Context, saJsonKeyPath string) (gcpiprangeclient.ComputeClient, error) {
			computeClient, err := compute.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(fakeHttpServer.URL))
			if err != nil {
				return nil, err
			}
			return gcpiprangeclient.NewComputeClientForService(computeClient), nil
		},
	)
}

func newFakeServiceNetworkingProvider(fakeHttpServer *httptest.Server) client.ClientProvider[gcpiprangeclient.ServiceNetworkingClient] {
	return client.NewCachedClientProvider(
		func(ctx context.Context, saJsonKeyPath string) (gcpiprangeclient.ServiceNetworkingClient, error) {
			svcNwClient, err := servicenetworking.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(fakeHttpServer.URL))
			if err != nil {
				return nil, err
			}
			crmService, err := cloudresourcemanager.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(fakeHttpServer.URL))
			if err != nil {
				return nil, err
			}
			return gcpiprangeclient.NewServiceNetworkingClientForService(svcNwClient, crmService), nil
		},
	)
}

type testStateFactory struct {
	factory                         StateFactory
	kcpCluster                      composed.StateCluster
	computeClientProvider           client.ClientProvider[gcpiprangeclient.ComputeClient]
	serviceNetworkingClientProvider client.ClientProvider[gcpiprangeclient.ServiceNetworkingClient]
	fakeHttpServer                  *httptest.Server
}

func newTestStateFactory(fakeHttpServer *httptest.Server) (*testStateFactory, error) {
	kcpClient := fake.NewClientBuilder().
		WithScheme(commonscheme.KcpScheme).
		WithObjects(&gcpIpRange).
		WithStatusSubresource(&gcpIpRange).
		Build()
	kcpCluster := composed.NewStateCluster(kcpClient, kcpClient, nil, commonscheme.KcpScheme)

	computeClientProvider := newFakeComputeClientProvider(fakeHttpServer)
	svcNwClientProvider := newFakeServiceNetworkingProvider(fakeHttpServer)
	env := abstractions.NewMockedEnvironment(map[string]string{
		"GCP_SA_JSON_KEY_PATH":        "test",
		"GCP_RETRY_WAIT_DURATION":     "100ms",
		"GCP_OPERATION_WAIT_DURATION": "100ms",
		"GCP_API_TIMEOUT_DURATION":    "100ms"})

	factory := NewStateFactory(svcNwClientProvider, computeClientProvider, env)

	return &testStateFactory{
		factory:                         factory,
		kcpCluster:                      kcpCluster,
		computeClientProvider:           computeClientProvider,
		serviceNetworkingClientProvider: svcNwClientProvider,
		fakeHttpServer:                  fakeHttpServer,
	}, nil

}

func (f *testStateFactory) newStateWith(ctx context.Context, ipRange *cloudcontrolv1beta1.IpRange) (*State, error) {
	return f.newStateWithScope(ctx, ipRange, gcpScope)
}

func (f *testStateFactory) newStateWithScope(ctx context.Context, ipRange *cloudcontrolv1beta1.IpRange, scope *cloudcontrolv1beta1.Scope) (*State, error) {
	snc, _ := f.serviceNetworkingClientProvider(ctx, "test")
	cc, _ := f.computeClientProvider(ctx, "test")

	focalState := focal.NewStateFactory().NewState(
		composed.NewStateFactory(f.kcpCluster).NewState(
			types.NamespacedName{
				Name:      ipRange.Name,
				Namespace: ipRange.Namespace,
			},
			ipRange))

	focalState.SetScope(scope)

	return newState(newTypesState(focalState), snc, cc), nil

}

var _ iprangetypes.State = &typesState{}

type typesState struct {
	focal.State
}

func (s *typesState) ObjAsIpRange() *cloudcontrolv1beta1.IpRange {
	return s.Obj().(*cloudcontrolv1beta1.IpRange)
}

func (s *typesState) Network() *cloudcontrolv1beta1.Network {
	return nil
}

func (s *typesState) ExistingCidrRanges() []string {
	return nil
}

func (s *typesState) SetExistingCidrRanges(v []string) {}

func newTypesState(focalState focal.State) iprangetypes.State {
	return &typesState{State: focalState}
}

// **** Global variables ****
var kymaRef = klog.ObjectRef{
	Name:      "skr",
	Namespace: "test",
}

var ipAddr = "10.20.30.0"
var prefix = 24
var gcpIpRange = cloudcontrolv1beta1.IpRange{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-ip-range",
		Namespace: kymaRef.Namespace,
		Labels: map[string]string{
			cloudcontrolv1beta1.LabelKymaName:   kymaRef.Name,
			cloudcontrolv1beta1.LabelRemoteName: "test-gcp-ip-range",
		},
		Finalizers: []string{api.CommonFinalizerDeletionHook},
	},
	Spec: cloudcontrolv1beta1.IpRangeSpec{
		RemoteRef: cloudcontrolv1beta1.RemoteRef{
			Name: "test-gcp-ip-range",
		},
		Scope: cloudcontrolv1beta1.ScopeRef{
			Name: kymaRef.Name,
		},
		Cidr: fmt.Sprintf("%s/%d", ipAddr, prefix),
		Options: cloudcontrolv1beta1.IpRangeOptions{
			Gcp: &cloudcontrolv1beta1.IpRangeGcp{
				Purpose: cloudcontrolv1beta1.GcpPurposePSA,
			},
		},
	},
	Status: cloudcontrolv1beta1.IpRangeStatus{
		Id: "cm-test-ip-range",
	},
}

var gcpScope = &cloudcontrolv1beta1.Scope{
	ObjectMeta: metav1.ObjectMeta{
		Name:      kymaRef.Name,
		Namespace: kymaRef.Namespace,
	},
	Spec: cloudcontrolv1beta1.ScopeSpec{
		Region: "us-west1",
		Scope: cloudcontrolv1beta1.ScopeInfo{
			Gcp: &cloudcontrolv1beta1.GcpScope{
				Project:    "test-project",
				VpcNetwork: "test-vpc",
			},
		},
	},
}

var opIdentifier = "/projects/test-project/locations/us-west1/operations/create-operation"
var urlGlobalAddress = "/projects/test-project/global/addresses"

var urlSvcNetworking = "services/servicenetworking.googleapis.com/connections"
