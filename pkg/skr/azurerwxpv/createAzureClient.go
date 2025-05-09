package azurerwxpv

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
)

func createAzureClient(ctx context.Context, st composed.State) (error, context.Context) {

	state := st.(*State)
	clientId := azureconfig.AzureConfig.DefaultCreds.ClientId
	clientSecret := azureconfig.AzureConfig.DefaultCreds.ClientSecret
	subscriptionId := state.Scope().Spec.Scope.Azure.SubscriptionId
	tenantId := state.Scope().Spec.Scope.Azure.TenantId

	cli, err := state.clientProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)
	if err != nil {
		return composed.LogErrorAndReturn(err, "error creating azure client", err, ctx)
	}

	state.client = cli

	return nil, nil
}
