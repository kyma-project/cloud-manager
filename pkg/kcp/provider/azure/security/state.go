package security

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azuresecurityclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/security/client"
	runtimetypes "github.com/kyma-project/cloud-manager/pkg/kcp/runtime/types"
)

func NewStateFactory(azureClientProvider azureclient.ClientProvider[azuresecurityclient.Client]) StateFactory {
	return &stateFactory{
		azureClientProvider: azureClientProvider,
	}
}

type StateFactory interface {
	NewState(ctx context.Context, runtimeState runtimetypes.State) (context.Context, composed.State, error)
}

type stateFactory struct {
	azureClientProvider azureclient.ClientProvider[azuresecurityclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, runtimeState runtimetypes.State) (context.Context, composed.State, error) {
	clientId := azureconfig.AzureConfig.DefaultCreds.ClientId
	clientSecret := azureconfig.AzureConfig.DefaultCreds.ClientSecret
	if runtimeState.Subscription().Status.Provider != cloudcontrolv1beta1.ProviderAzure {
		return ctx, nil, fmt.Errorf("subscription for Runtime must be of provider Azure, but subscription %q is of provider %q", runtimeState.Subscription().Name, runtimeState.Subscription().Status.Provider)
	}
	tenantId := runtimeState.Subscription().Status.SubscriptionInfo.Azure.TenantId
	subscriptionId := runtimeState.Subscription().Status.SubscriptionInfo.Azure.SubscriptionId

	c, err := f.azureClientProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)
	if err != nil {
		return ctx, nil, fmt.Errorf("error creating azure client: %w", err)
	}

	logger := composed.LoggerFromCtx(ctx).
		WithValues(
			"azureSubscriptionId", subscriptionId,
			"azureTenantId", tenantId,
		)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	return ctx, newState(runtimeState, c), nil
}

func newState(runtimeState runtimetypes.State, azureClient azuresecurityclient.Client) *State {
	return &State{
		State:       runtimeState,
		azureClient: azureClient,
	}
}

type State struct {
	runtimetypes.State

	azureClient azuresecurityclient.Client

	// resourceGroupData one per runtime, with name pattern "kyma-security-{shootName}" and is used for:
	// * storage account
	// * log analytics workspace
	// * flow log
	resourceGroupData *armresources.ResourceGroup

	// resourceGroupWatcher has fixed name "NetworkWatcherRG" and is used for:
	// * watcher
	resourceGroupWatcher *armresources.ResourceGroup

	// watcher from resourceGroupWatcher, one per location with name pattern "NetworkWatcher_{location}", ie NetworkWatcher_westeurope
	// This network watcher is only created when missing, but is not deleted when security is disabled
	watcher *armnetwork.Watcher

	// storageAccount from RG resourceGroupData, one per runtime, with name pattern "sapkymasec{shootName}{optional-sufix}"
	// Since Azure storage account must be globally unique it might happen prefer name is not available, thus optional
	// sufix is used to achieve global uniqueness. To load runtime related storage account, all from the RG have to be
	// listed and one matching the tag tagKymaRuntimeId = runtime.name picked and set to state
	storageAccount *armstorage.Account

	// flowLog under NetworkWatcherRG/NetworkWatcher_{location}, name = vpcNetwork.Status.Identifiers.Name
	flowLog *armnetwork.FlowLog

	// loadedSecurityPricing Defender for Cloud plans used to determine service on/off action
	loadedSecurityPricing []*armsecurity.Pricing
}

const (
	tagKymaRuntimeId = "kyma.runtime-id"
	tagKymaShootName = "kyma.shoot-name"

	flowLogRetentionDays = 30
)

func (s *State) ObjAsRuntime() *infrastructuremanagerv1.Runtime {
	return s.Obj().(*infrastructuremanagerv1.Runtime)
}
