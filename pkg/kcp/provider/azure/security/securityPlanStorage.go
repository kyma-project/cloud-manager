package security

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

var storageRequiredExtensions = []armsecurity.Extension{
	{
		Name:      new("OnUploadMalwareScanning"),
		IsEnabled: new(armsecurity.IsEnabledTrue),
		AdditionalExtensionProperties: map[string]any{
			"CapGBPerMonthPerStorageAccount": "10000",
		},
	},
	{
		Name:      new("SensitiveDataDiscovery"),
		IsEnabled: new(armsecurity.IsEnabledTrue),
	},
}

func securityPlanStorage(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	planName := "StorageAccounts"
	subPlan := "DefenderForStorageV2"
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
		if targetTier == armsecurity.PricingTierFree {
			return nil, ctx
		}
		if currentPricing.Properties.SubPlan != nil &&
			*currentPricing.Properties.SubPlan == subPlan &&
			hasAllExtensionsEnabled(currentPricing.Properties.Extensions, storageRequiredExtensions) {
			return nil, ctx
		}
	}

	scopeId := azureutil.NewSubscriptionResourceId(
		state.Subscription().Status.SubscriptionInfo.Azure.SubscriptionId,
	).String()

	pricing := armsecurity.Pricing{
		Properties: &armsecurity.PricingProperties{
			PricingTier: &targetTier,
		},
	}
	if targetTier == armsecurity.PricingTierStandard {
		pricing.Properties.SubPlan = ptr.To(subPlan)
		pricing.Properties.Extensions = toExtensionPointers(storageRequiredExtensions)
	}

	_, err := state.azureClient.UpdateSecurityPricing(ctx, scopeId, planName, pricing, nil)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating StorageAccounts security pricing", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	logger.Info("Updated StorageAccounts security pricing", "tier", targetTier, "subPlan", subPlan)
	return nil, ctx
}
