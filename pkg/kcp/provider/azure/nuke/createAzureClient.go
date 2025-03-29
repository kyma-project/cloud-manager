package nuke

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func createAzureClient(ctx context.Context, st composed.State) (error, context.Context) {

	state := st.(*State)
	clientId := azureconfig.AzureConfig.DefaultCreds.ClientId
	clientSecret := azureconfig.AzureConfig.DefaultCreds.ClientSecret
	subscriptionId := state.Scope().Spec.Scope.Azure.SubscriptionId
	tenantId := state.Scope().Spec.Scope.Azure.TenantId

	cli, err := state.azureClientProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)
	if err != nil {
		return composed.LogErrorAndReturn(err, "error creating azure client", composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx)
	}

	state.azureClient = cli

	return nil, nil
}
