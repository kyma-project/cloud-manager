package security

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
)

var serversPlan = PlanSpec{
	PlanName: "VirtualMachines",
	SubPlan:  "P2",
	RequiredExtensions: []ExtensionSpec{
		{
			Name:                          "AgentlessVmScanning",
			AdditionalExtensionProperties: map[string]any{"ExclusionTags": "[]"},
		},
	},
	DisabledExtensions: []string{
		//"MdeDesignatedSubscription",
		//"FileIntegrityMonitoring",
	},
}

func securityPlanServers(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	targetTier := armsecurity.PricingTierFree
	if state.SecurityServiceEnabledOnSubscription() {
		targetTier = armsecurity.PricingTierStandard
	}

	var currentPricing *armsecurity.Pricing
	for _, p := range state.loadedSecurityPricing {
		if p.Name != nil && *p.Name == serversPlan.PlanName {
			currentPricing = p
			break
		}
	}

	if !serversPlan.NeedsUpdate(currentPricing, targetTier) {
		return nil, ctx
	}

	scopeId := azureutil.NewSubscriptionResourceId(
		state.Subscription().Status.SubscriptionInfo.Azure.SubscriptionId,
	).String()

	pricing := serversPlan.BuildPricing(targetTier)

	logger.Info("Updating Servers security pricing", "currentPricing", currentPricing, "updatePricing", pricing)

	resp, err := state.azureClient.UpdateSecurityPricing(ctx, scopeId, serversPlan.PlanName, pricing, nil)
	if err != nil {
		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error updating defender security pricing Servers P2: %s", err.Error()), state.ObjAsRuntime().Generation)
		return composed.LogErrorAndReturn(err, "Error updating Servers security pricing", composed.StopWithRequeueDelay(rate.Slow1s.When(state.ObjAsRuntime())), ctx)
	}
	if extErr := serversPlan.extensionErrors(resp); extErr != nil {
		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error enabling defender security pricing Servers P2 extensions: %s", extErr.Error()), state.ObjAsRuntime().Generation)
		return composed.LogErrorAndReturn(extErr, "Error enabling Servers security pricing extensions", composed.StopWithRequeueDelay(rate.Slow1s.When(state.ObjAsRuntime())), ctx)
	}

	return nil, ctx
}
