package scope

import (
	gardenerTypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerClient "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
	scopeclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/scope/client"
	awsClient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/client"
	skrruntime "github.com/kyma-project/cloud-manager/components/kcp/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kubernetesClient "k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

type StateFactory interface {
	NewState(req ctrl.Request) *State
}

func NewStateFactory(
	baseStateFactory composed.StateFactory,
	awsStsClientProvider awsClient.GardenClientProvider[scopeclient.AwsStsClient],
	activeSkrCollection skrruntime.ActiveSkrCollection,
) StateFactory {
	return &stateFactory{
		baseStateFactory:     baseStateFactory,
		awsStsClientProvider: awsStsClientProvider,
		activeSkrCollection:  activeSkrCollection,
	}
}

type stateFactory struct {
	baseStateFactory     composed.StateFactory
	awsStsClientProvider awsClient.GardenClientProvider[scopeclient.AwsStsClient]
	activeSkrCollection  skrruntime.ActiveSkrCollection
}

func (f *stateFactory) NewState(req ctrl.Request) *State {
	baseState := f.baseStateFactory.NewState(req.NamespacedName, &cloudcontrolv1beta1.Scope{})

	return newState(
		baseState,
		f.awsStsClientProvider,
		f.activeSkrCollection,
	)
}

// =====================================================================

func newState(
	baseState composed.State,
	awsStsClientProvider awsClient.GardenClientProvider[scopeclient.AwsStsClient],
	activeSkrCollection skrruntime.ActiveSkrCollection,
) *State {
	return &State{
		State:                baseState,
		awsStsClientProvider: awsStsClientProvider,
		activeSkrCollection:  activeSkrCollection,
		credentialData:       map[string]string{},
	}
}

type State struct {
	composed.State

	kyma                *unstructured.Unstructured
	moduleState         util.KymaModuleState
	activeSkrCollection skrruntime.ActiveSkrCollection

	shootName      string
	shootNamespace string

	gardenerClient  gardenerClient.CoreV1beta1Interface
	gardenK8sClient kubernetesClient.Interface

	provider       cloudcontrolv1beta1.ProviderType
	shoot          *gardenerTypes.Shoot
	credentialData map[string]string

	awsStsClientProvider awsClient.GardenClientProvider[scopeclient.AwsStsClient]
}

func (s *State) ObjAsScope() *cloudcontrolv1beta1.Scope {
	return s.Obj().(*cloudcontrolv1beta1.Scope)
}
