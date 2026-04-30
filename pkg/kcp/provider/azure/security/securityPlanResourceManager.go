package security

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func securityPlanResourceManager(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	planName := "Arm"
	targetTier := armsecurity.PricingTierFree
	if state.SecurityServiceEnabledOnSubscription() {
		targetTier = armsecurity.PricingTierStandard
	}

	var currentPricing *armsecurity.Pricing
	for _, p := range state.loadedSecurityPricing {
		if p.Name != nil && *p.Name == planName {
			currentPricing = p
			break
		}
	}

	if currentPricing != nil &&
		currentPricing.Properties != nil &&
		currentPricing.Properties.PricingTier != nil &&
		*currentPricing.Properties.PricingTier == targetTier {
		return nil, ctx
	}

	scopeId := azureutil.NewSubscriptionResourceId(
		state.Subscription().Status.SubscriptionInfo.Azure.SubscriptionId,
	).String()

	pricing := armsecurity.Pricing{
		Properties: &armsecurity.PricingProperties{
			PricingTier: &targetTier,
		},
	}

	_, err := state.azureClient.UpdateSecurityPricing(ctx, scopeId, planName, pricing, nil)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating Arm security pricing", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	logger.Info("Updated Arm security pricing", "tier", targetTier)
	return nil, ctx
}
