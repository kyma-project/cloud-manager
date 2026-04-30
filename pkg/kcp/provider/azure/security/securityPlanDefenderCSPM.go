package security

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

var defenderCSPMRequiredExtensions = []armsecurity.Extension{
	{
		Name:      new("SensitiveDataDiscovery"),
		IsEnabled: new(armsecurity.IsEnabledTrue),
	},
	{
		Name:      new("ContainerRegistriesVulnerabilityAssessments"),
		IsEnabled: new(armsecurity.IsEnabledTrue),
	},
	{
		Name:      new("AgentlessDiscoveryForKubernetes"),
		IsEnabled: new(armsecurity.IsEnabledTrue),
	},
	{
		Name:      new("AgentlessVmScanning"),
		IsEnabled: new(armsecurity.IsEnabledTrue),
	},
	{
		Name:      new("EntraPermissionsManagement"),
		IsEnabled: new(armsecurity.IsEnabledTrue),
	},
	{
		Name:      new("ApiPosture"),
		IsEnabled: new(armsecurity.IsEnabledTrue),
	},
	{
		Name:      new("AgentlessServerlessPosture"),
		IsEnabled: new(armsecurity.IsEnabledTrue),
	},
}

func securityPlanDefenderCSPM(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	planName := "CloudPosture"
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
		if hasAllExtensionsEnabled(currentPricing.Properties.Extensions, defenderCSPMRequiredExtensions) {
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
		pricing.Properties.Extensions = toExtensionPointers(defenderCSPMRequiredExtensions)
	}

	_, err := state.azureClient.UpdateSecurityPricing(ctx, scopeId, planName, pricing, nil)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating CloudPosture security pricing", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	logger.Info("Updated CloudPosture security pricing", "tier", targetTier)
	return nil, ctx
}
