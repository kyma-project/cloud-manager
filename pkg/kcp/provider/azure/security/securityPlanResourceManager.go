package security

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
)

var resourceManagerPlan = PlanSpec{
	PlanName: "Arm",
}

func securityPlanResourceManager(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	targetTier := armsecurity.PricingTierFree
	if state.SecurityServiceEnabledOnSubscription() {
		targetTier = armsecurity.PricingTierStandard
	}

	var currentPricing *armsecurity.Pricing
	for _, p := range state.loadedSecurityPricing {
		if p.Name != nil && *p.Name == resourceManagerPlan.PlanName {
			currentPricing = p
			break
		}
	}

	if !resourceManagerPlan.NeedsUpdate(currentPricing, targetTier) {
		return nil, ctx
	}

	scopeId := azureutil.NewSubscriptionResourceId(
		state.Subscription().Status.SubscriptionInfo.Azure.SubscriptionId,
	).String()

	pricing := resourceManagerPlan.BuildPricing(targetTier)

	logger.Info("Updating Arm security pricing", "currentPricing", currentPricing, "updatePricing", pricing)

	resp, err := state.azureClient.UpdateSecurityPricing(ctx, scopeId, resourceManagerPlan.PlanName, pricing, nil)
	if err != nil {
		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error updating defender security pricing ARM: %s", err.Error()), state.ObjAsRuntime().Generation)
		return composed.LogErrorAndReturn(err, "Error updating Arm security pricing", composed.StopWithRequeueDelay(rate.Slow1s.When(state.ObjAsRuntime())), ctx)
	}
	if extErr := resourceManagerPlan.extensionErrors(resp); extErr != nil {
		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error enabling defender security pricing ARM extensions: %s", extErr.Error()), state.ObjAsRuntime().Generation)
		return composed.LogErrorAndReturn(extErr, "Error enabling Arm security pricing extensions", composed.StopWithRequeueDelay(rate.Slow1s.When(state.ObjAsRuntime())), ctx)
	}

	return nil, ctx
}
