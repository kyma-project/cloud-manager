package nuke

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	azureClient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/nuke/client"
	"k8s.io/utils/ptr"
)

type StateFactory interface {
	NewState(ctx context.Context, nukeState nuketypes.State) (focal.State, error)
}

func NewStateFactory(
	azureClientProvider azureClient.ClientProvider[client.NukeRwxBackupClient],
	env abstractions.Environment) StateFactory {
	return stateFactory{
		azureClientProvider: azureClientProvider,
		env:                 env,
	}
}

type stateFactory struct {
	azureClientProvider azureClient.ClientProvider[client.NukeRwxBackupClient]
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

	azureClientProvider azureClient.ClientProvider[client.NukeRwxBackupClient]
	env                 abstractions.Environment
	azureClient         client.NukeRwxBackupClient

	recoveryVaults []*armrecoveryservices.Vault
	protectedItems []*armrecoveryservicesbackup.ProtectedItemResource
}

type azureProtectedItem struct {
	*armrecoveryservicesbackup.ProtectedItemResource
}

func (b azureProtectedItem) GetId() string {
	return ptr.Deref(b.ID, "")
}

func (b azureProtectedItem) GetObject() interface{} {
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

type ProviderNukeStatus struct {
	v1beta1.NukeStatus
}

func (s *State) GetSubscriptionId() string {
	return s.Scope().Spec.Scope.Azure.SubscriptionId
}
