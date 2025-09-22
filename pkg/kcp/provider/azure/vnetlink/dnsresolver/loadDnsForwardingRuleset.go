package dnsresolver

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"k8s.io/utils/ptr"
)

func loadDnsForwardingRuleset(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ruleset, err := state.remoteClient.Get(ctx,
		state.rulesetId.ResourceGroup,
		state.rulesetId.ResourceName)

	if err == nil {
		ctx = composed.LoggerIntoCtx(ctx, logger.WithValues("dnsForwardingRulesetId", ptr.Deref(ruleset.ID, "")))
		state.ruleset = ruleset
		return nil, ctx
	}

	return azuremeta.HandleError(err, state.ObjAsAzureVNetLink()).
		WithDefaultReason(cloudcontrolv1beta1.ReasonFailedLoadingPrivateDnzZone).
		WithDefaultMessage("Failed loading DnsForwardingRuleset").
		WithTooManyRequestsMessage("Too many requests on loading DnsForwardingRuleset").
		WithUpdateStatusMessage("Error updating KCP AzureVNetLink status on failed loading of DnsForwardingRuleset").
		WithNotFoundMessage("DnsForwardingRuleset not found").
		Run(ctx, state)
}
