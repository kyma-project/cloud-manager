package scope

import (
	gardenerTypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerClient "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/scope/client"
	awsClient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/client"
	skrruntime "github.com/kyma-project/cloud-manager/components/kcp/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/util"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kubernetesClient "k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

type StateFactory interface {
	NewState(req ctrl.Request) *State
}

func NewStateFactory(
	baseStateFactory composed.StateFactory,
	fileReader abstractions.FileReader,
	awsStsClientProvider awsClient.GardenClientProvider[client.AwsStsClient],
) StateFactory {
	return &stateFactory{
		baseStateFactory:     baseStateFactory,
		fileReader:           fileReader,
		awsStsClientProvider: awsStsClientProvider,
	}
}

type stateFactory struct {
	baseStateFactory     composed.StateFactory
	fileReader           abstractions.FileReader
	awsStsClientProvider awsClient.GardenClientProvider[client.AwsStsClient]
}

func (f *stateFactory) NewState(req ctrl.Request) *State {
	baseState := f.baseStateFactory.NewState(req.NamespacedName, &cloudcontrolv1beta1.Scope{})

	return newState(
		baseState,
		f.fileReader,
		f.awsStsClientProvider,
	)
}

// =====================================================================

func newState(baseState composed.State, fileReader abstractions.FileReader, awsStsClientProvider awsClient.GardenClientProvider[client.AwsStsClient]) *State {
	return &State{
		State:                baseState,
		fileReader:           fileReader,
		awsStsClientProvider: awsStsClientProvider,
		credentialData:       map[string]string{},
	}
}

type State struct {
	composed.State

	fileReader abstractions.FileReader

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

	awsStsClientProvider awsClient.GardenClientProvider[client.AwsStsClient]
}

func (s *State) ObjAsScope() *cloudcontrolv1beta1.Scope {
	return s.Obj().(*cloudcontrolv1beta1.Scope)
}
