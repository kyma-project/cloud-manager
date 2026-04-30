package azurerwxpv

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"github.com/kyma-project/cloud-manager/pkg/skr/azurerwxpv/client"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State struct {
	commonscope.State
	client         client.Client
	clientProvider azureclient.ClientProvider[client.Client]
	fileShare      *armstorage.FileShareItem
	fileShareName  string

	recoveryVaults []*armrecoveryservices.Vault
	protectedId    string
	protectedItem  *armrecoveryservicesbackup.AzureFileshareProtectedItem
}

func (s *State) ObjAsPV() *corev1.PersistentVolume {
	return s.Obj().(*corev1.PersistentVolume)
}

type stateFactory struct {
	baseStateFactory        composed.StateFactory
	commonScopeStateFactory commonscope.StateFactory
	clientProvider          azureclient.ClientProvider[client.Client]
}

func (f *stateFactory) NewState(ctx context.Context, req ctrl.Request) (*State, error) {
	scopeState, err := f.commonScopeStateFactory.NewState(
		ctx,
		req.NamespacedName,
		f.baseStateFactory.NewState(req.NamespacedName, &corev1.PersistentVolume{}),
	)
	if err != nil {
		return nil, err
	}
	return &State{
		State:          scopeState,
		clientProvider: f.clientProvider,
	}, nil
}

func newStateFactory(
	baseStateFactory composed.StateFactory,
	commonScopeStateFactory commonscope.StateFactory,
	clientProvider azureclient.ClientProvider[client.Client],
) *stateFactory {
	return &stateFactory{
		baseStateFactory:        baseStateFactory,
		commonScopeStateFactory: commonScopeStateFactory,
		clientProvider:          clientProvider,
	}
}
