package scope

import (
	gardenerTypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerClient "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsClient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	scopeclient "github.com/kyma-project/cloud-manager/pkg/kcp/scope/client"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kubernetesClient "k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

// NukeScopesWithoutKyma defines if during the Scope reconciliation a Nuke resource will be created
// if Kyma does not exist. For the normal behavior it should be true, so that Nuke could
// clean up the Scope's orphaned resources. But since the majority of tests creates Scope
// but doesn't create Kyma, this flag in test suite setup is set to `false`, to prevent deletion
// of all the Scope resources and allow tests to function
var NukeScopesWithoutKyma = true

type StateFactory interface {
	NewState(req ctrl.Request) *State
}

func NewStateFactory(
	baseStateFactory composed.StateFactory,
	awsStsClientProvider awsClient.GardenClientProvider[scopeclient.AwsStsClient],
	activeSkrCollection skrruntime.ActiveSkrCollection,
	gcpServiceUsageClientProvider gcpclient.ClientProvider[gcpclient.ServiceUsageClient],
) StateFactory {
	return &stateFactory{
		baseStateFactory:              baseStateFactory,
		awsStsClientProvider:          awsStsClientProvider,
		activeSkrCollection:           activeSkrCollection,
		gcpServiceUsageClientProvider: gcpServiceUsageClientProvider,
	}
}

type stateFactory struct {
	baseStateFactory              composed.StateFactory
	awsStsClientProvider          awsClient.GardenClientProvider[scopeclient.AwsStsClient]
	activeSkrCollection           skrruntime.ActiveSkrCollection
	gcpServiceUsageClientProvider gcpclient.ClientProvider[gcpclient.ServiceUsageClient]
}

func (f *stateFactory) NewState(req ctrl.Request) *State {
	baseState := f.baseStateFactory.NewState(req.NamespacedName, &cloudcontrolv1beta1.Scope{})

	return newState(
		baseState,
		f.awsStsClientProvider,
		f.activeSkrCollection,
		f.gcpServiceUsageClientProvider,
	)
}

// =====================================================================

func newState(
	baseState composed.State,
	awsStsClientProvider awsClient.GardenClientProvider[scopeclient.AwsStsClient],
	activeSkrCollection skrruntime.ActiveSkrCollection,
	gcpServiceUsageClientProvider gcpclient.ClientProvider[gcpclient.ServiceUsageClient],
) *State {
	return &State{
		State:                         baseState,
		awsStsClientProvider:          awsStsClientProvider,
		activeSkrCollection:           activeSkrCollection,
		gcpServiceUsageClientProvider: gcpServiceUsageClientProvider,
		credentialData:                map[string]string{},
	}
}

type State struct {
	composed.State

	kyma                *unstructured.Unstructured
	activeSkrCollection skrruntime.ActiveSkrCollection

	moduleState  util.KymaModuleState
	moduleInSpec bool
	skrActive    bool

	shootName      string
	shootNamespace string

	gardenerClient  gardenerClient.CoreV1beta1Interface
	gardenK8sClient kubernetesClient.Interface

	provider       cloudcontrolv1beta1.ProviderType
	shoot          *gardenerTypes.Shoot
	credentialData map[string]string

	kcpNetworkKyma *cloudcontrolv1beta1.Network
	kcpNetworkCm   *cloudcontrolv1beta1.Network

	awsStsClientProvider          awsClient.GardenClientProvider[scopeclient.AwsStsClient]
	gcpServiceUsageClientProvider gcpclient.ClientProvider[gcpclient.ServiceUsageClient]
}

func (s *State) ObjAsScope() *cloudcontrolv1beta1.Scope {
	return s.Obj().(*cloudcontrolv1beta1.Scope)
}
