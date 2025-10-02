package subscription

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	subscriptionclient "github.com/kyma-project/cloud-manager/pkg/kcp/subscription/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	gardenerClient  client.Client
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
