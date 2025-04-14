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
	azurerwxvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
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
	cloudcontrolv1beta1.NukeStatus
}

func (s *State) GetSubscriptionId() string {
	return s.Scope().Spec.Scope.Azure.SubscriptionId
}

func (s *State) getContainerNames(vault *armrecoveryservices.Vault) ([]string, error) {
	containerNames := []string{}
	keys := make(map[string]string)
	nuke := s.ObjAsNuke()
	for _, rks := range nuke.Status.Resources {
		if rks.ResourceType == cloudcontrolv1beta1.ProviderResource && rks.Kind == "AzureRwxVolumeBackup" {
			for objId := range rks.Objects {
				_, _, vaultName, containerName, _, err := azurerwxvolumebackupclient.ParseProtectedItemId(objId)
				if err != nil {
					return containerNames, err
				}
				if _, exists := keys[containerName]; !exists {
					keys[containerName] = vaultName
					containerNames = append(containerNames, containerName)
				}
			}
		}
	}
	return containerNames, nil
}
