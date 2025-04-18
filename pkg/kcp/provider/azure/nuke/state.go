package nuke

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/nuke/client"
	"k8s.io/utils/ptr"
)

type StateFactory interface {
	NewState(ctx context.Context, nukeState nuketypes.State) (focal.State, error)
}

func NewStateFactory(
	azureClientProvider azureclient.ClientProvider[client.NukeRwxBackupClient],
	env abstractions.Environment) StateFactory {
	return stateFactory{
		azureClientProvider: azureClientProvider,
		env:                 env,
	}
}

type stateFactory struct {
	azureClientProvider azureclient.ClientProvider[client.NukeRwxBackupClient]
	env                 abstractions.Environment
}

func (f stateFactory) NewState(ctx context.Context, nukeState nuketypes.State) (focal.State, error) {
	return &State{
		State:               nukeState,
		azureClientProvider: f.azureClientProvider,
		env:                 f.env,
	}, nil
}

type State struct {
	nuketypes.State
	ProviderResources []*nuketypes.ProviderResourceKindState

	azureClientProvider azureclient.ClientProvider[client.NukeRwxBackupClient]
	env                 abstractions.Environment
	azureClient         client.NukeRwxBackupClient

	recoveryVaults       []*armrecoveryservices.Vault
	protectionContainers map[string]*armrecoveryservicesbackup.AzureStorageContainer
	protectedItems       map[string]*armrecoveryservicesbackup.AzureFileshareProtectedItem
}

type azureFileshare struct {
	*armrecoveryservicesbackup.AzureFileshareProtectedItem
	id string
}

func (b azureFileshare) GetId() string {
	return b.id
}

func (b azureFileshare) GetObject() interface{} {
	return b
}

type azureVault struct {
	*armrecoveryservices.Vault
}

func (v azureVault) GetId() string {
	return ptr.Deref(v.ID, "")
}

func (v azureVault) GetObject() interface{} {
	return v
}

type azureContainer struct {
	*armrecoveryservicesbackup.AzureStorageContainer
	id string
}

func (v azureContainer) GetId() string {
	return v.id
}

func (v azureContainer) GetObject() interface{} {
	return v
}

type ProviderNukeStatus struct {
	cloudcontrolv1beta1.NukeStatus
}

func (s *State) GetSubscriptionId() string {
	return s.Scope().Spec.Scope.Azure.SubscriptionId
}
