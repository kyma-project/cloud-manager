package subscription

import (
	gardenerclient "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	subscriptionclient "github.com/kyma-project/cloud-manager/pkg/kcp/subscription/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubernetesclient "k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

type StateFactory interface {
	NewState(req ctrl.Request) *State
}

func NewStateFactory(
	baseStateFactory composed.StateFactory,
	awsStsClientProvider awsclient.GardenClientProvider[subscriptionclient.AwsStsClient],
) StateFactory {
	return &stateFactory{
		baseStateFactory:     baseStateFactory,
		awsStsClientProvider: awsStsClientProvider,
	}
}

type stateFactory struct {
	baseStateFactory     composed.StateFactory
	awsStsClientProvider awsclient.GardenClientProvider[subscriptionclient.AwsStsClient]
}

func (f *stateFactory) NewState(req ctrl.Request) *State {
	baseState := f.baseStateFactory.NewState(req.NamespacedName, &cloudcontrolv1beta1.Subscription{})

	return newState(
		baseState,
		f.awsStsClientProvider,
	)
}

// State ==================================

type State struct {
	composed.State

	awsStsClientProvider awsclient.GardenClientProvider[subscriptionclient.AwsStsClient]

	gardenNamespace string
	gardenerClient  gardenerclient.CoreV1beta1Interface
	gardenK8sClient kubernetesclient.Interface
	provider        cloudcontrolv1beta1.ProviderType
	credentialData  map[string]string

	resources map[schema.GroupVersionKind][]metav1.PartialObjectMetadata
}

func newState(
	baseState composed.State,
	awsStsClientProvider awsclient.GardenClientProvider[subscriptionclient.AwsStsClient],
) *State {
	return &State{
		State:                baseState,
		awsStsClientProvider: awsStsClientProvider,
	}
}

func (s *State) ObjAsSubscription() *cloudcontrolv1beta1.Subscription {
	return s.Obj().(*cloudcontrolv1beta1.Subscription)
}
