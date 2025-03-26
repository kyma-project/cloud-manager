package azurerwxvolumebackup

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func createClient(ctx context.Context, st composed.State) (error, context.Context) {

	state := st.(*State)
	clientId := azureconfig.AzureConfig.DefaultCreds.ClientId
	clientSecret := azureconfig.AzureConfig.DefaultCreds.ClientSecret
	subscriptionId := state.scope.Spec.Scope.Azure.SubscriptionId
	tenantId := state.scope.Spec.Scope.Azure.TenantId

	cli, err := state.clientProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)
	if err != nil {
		return composed.LogErrorAndReturn(err, "error creating azure client", composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx)
	}

	state.client = cli

	return nil, nil
}
