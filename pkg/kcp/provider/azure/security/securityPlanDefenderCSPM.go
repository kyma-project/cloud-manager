package security

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"gopkg.in/yaml.v3"
)

var defenderCSPMPlan = PlanSpec{
	PlanName: "CloudPosture",
	RequiredExtensions: []ExtensionSpec{
		{Name: "SensitiveDataDiscovery"},
		{
			Name:                          "AgentlessVmScanning",
			AdditionalExtensionProperties: map[string]any{"ExclusionTags": "[]"},
		},
		{Name: "EntraPermissionsManagement"},
		{Name: "ApiPosture"},
	},
	DisabledExtensions: []string{
		//"ContainerRegistriesVulnerabilityAssessments",
		//"AgentlessDiscoveryForKubernetes",
		//"AgentlessServerlessPosture",
		//"DatabricksSecurityPosture",
	},
}

func securityPlanDefenderCSPM(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	targetTier := armsecurity.PricingTierFree
	if state.SecurityServiceEnabledOnSubscription() {
		targetTier = armsecurity.PricingTierStandard
	}

	var currentPricing *armsecurity.Pricing
	for _, p := range state.loadedSecurityPricing {
		if p.Name != nil && *p.Name == defenderCSPMPlan.PlanName {
			currentPricing = p
			break
		}
	}

	if !defenderCSPMPlan.NeedsUpdate(currentPricing, targetTier) {
		return nil, ctx
	}

	scopeId := azureutil.NewSubscriptionResourceId(
		state.Subscription().Status.SubscriptionInfo.Azure.SubscriptionId,
	).String()

	pricing := defenderCSPMPlan.BuildPricing(targetTier)

	logger.Info("Updating CloudPosture security pricing", "currentPricing", currentPricing, "updatePricing", pricing)

	resp, err := state.azureClient.UpdateSecurityPricing(ctx, scopeId, defenderCSPMPlan.PlanName, pricing, nil)
	if err != nil {
		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error updating defender security pricing CSPM: %s", err.Error()), state.ObjAsRuntime().Generation)
		txtCurrent, _ := yaml.Marshal(currentPricing)
		txtNew, _ := yaml.Marshal(pricing)
		logger.Error(err, "Error updating CloudPosture security pricing", "currentPricing", string(txtCurrent), "updatePricing", string(txtNew))
		return composed.StopWithRequeueDelay(rate.Slow1s.When(state.ObjAsRuntime())), ctx
	}
	if extErr := defenderCSPMPlan.extensionErrors(resp); extErr != nil {
		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error enabling defender security pricing CSPM extensions: %s", extErr.Error()), state.ObjAsRuntime().Generation)
		txtCurrent, _ := yaml.Marshal(currentPricing)
		txtNew, _ := yaml.Marshal(pricing)
		logger.Error(err, "Error enabling CloudPosture security pricing extensions", "currentPricing", string(txtCurrent), "updatePricing", string(txtNew))
		return composed.StopWithRequeueDelay(rate.Slow1s.When(state.ObjAsRuntime())), ctx
	}

	return nil, ctx
}
