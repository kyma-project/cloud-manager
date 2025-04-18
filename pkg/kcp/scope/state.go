package scope

import (
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerclient "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	scopeclient "github.com/kyma-project/cloud-manager/pkg/kcp/scope/client"
	scopetypes "github.com/kyma-project/cloud-manager/pkg/kcp/scope/types"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kubernetesclient "k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

type StateFactory interface {
	NewState(req ctrl.Request) *State
}

func NewStateFactory(
	baseStateFactory composed.StateFactory,
	activeSkrCollection skrruntime.ActiveSkrCollection,
	awsStsClientProvider awsclient.GardenClientProvider[scopeclient.AwsStsClient],
	gcpServiceUsageClientProvider gcpclient.ClientProvider[gcpclient.ServiceUsageClient],
) StateFactory {
	return &stateFactory{
		baseStateFactory:              baseStateFactory,
		activeSkrCollection:           activeSkrCollection,
		awsStsClientProvider:          awsStsClientProvider,
		gcpServiceUsageClientProvider: gcpServiceUsageClientProvider,
	}
}

type stateFactory struct {
	baseStateFactory              composed.StateFactory
	activeSkrCollection           skrruntime.ActiveSkrCollection
	awsStsClientProvider          awsclient.GardenClientProvider[scopeclient.AwsStsClient]
	gcpServiceUsageClientProvider gcpclient.ClientProvider[gcpclient.ServiceUsageClient]
}

func (f *stateFactory) NewState(req ctrl.Request) *State {
	baseState := f.baseStateFactory.NewState(req.NamespacedName, &cloudcontrolv1beta1.Scope{})

	return newState(
		baseState,
		f.activeSkrCollection,
		f.awsStsClientProvider,
		f.gcpServiceUsageClientProvider,
	)
}

// =====================================================================

func newState(
	baseState composed.State,
	activeSkrCollection skrruntime.ActiveSkrCollection,
	awsStsClientProvider awsclient.GardenClientProvider[scopeclient.AwsStsClient],
	gcpServiceUsageClientProvider gcpclient.ClientProvider[gcpclient.ServiceUsageClient],
) *State {
	return &State{
		State:                         baseState,
		activeSkrCollection:           activeSkrCollection,
		awsStsClientProvider:          awsStsClientProvider,
		gcpServiceUsageClientProvider: gcpServiceUsageClientProvider,
		credentialData:                map[string]string{},
	}
}

var _ scopetypes.State = &State{}

type State struct {
	composed.State

	activeSkrCollection skrruntime.ActiveSkrCollection

	gardenerCluster        *unstructured.Unstructured
	gardenerClusterSummary *util.GardenerClusterSummary

	shootName      string
	shootNamespace string

	gardenerClient  gardenerclient.CoreV1beta1Interface
	gardenK8sClient kubernetesclient.Interface

	provider       cloudcontrolv1beta1.ProviderType
	shoot          *gardenertypes.Shoot
	credentialData map[string]string

	kcpNetworkKyma *cloudcontrolv1beta1.Network

	awsStsClientProvider          awsclient.GardenClientProvider[scopeclient.AwsStsClient]
	gcpServiceUsageClientProvider gcpclient.ClientProvider[gcpclient.ServiceUsageClient]

	nuke *cloudcontrolv1beta1.Nuke
}

func (s *State) ObjAsScope() *cloudcontrolv1beta1.Scope {
	return s.Obj().(*cloudcontrolv1beta1.Scope)
}
