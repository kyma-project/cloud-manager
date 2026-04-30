package security

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func securityPricingLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	scopeId := azureutil.NewSubscriptionResourceId(
		state.Subscription().Status.SubscriptionInfo.Azure.SubscriptionId,
	).String()

	resp, err := state.azureClient.ListSecurityPricings(
		ctx,
		scopeId,
		&armsecurity.PricingsClientListOptions{},
	)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Err loading Azure security pricings", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	state.loadedSecurityPricing = resp.Value

	return nil, ctx
}
