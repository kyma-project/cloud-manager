package scope

import (
	"context"
	"errors"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func createScopeAzure(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	subscriptionID, ok := state.credentialData["subscriptionID"]
	if !ok {
		err := errors.New("gardener credential for azure missing subscriptionID key")
		logger.Error(err, "error defining Azure scope")
		return composed.StopAndForget, nil // no requeue
	}

	tenantID, ok := state.credentialData["tenantID"]
	if !ok {
		err := errors.New("gardener credential for azure missing tenantID key")
		logger.Error(err, "error defining Azure scope")
		return composed.StopAndForget, nil // no requeue
	}

	// just create the scope with Azure specifics, the ensureScopeCommonFields will set common values
	scope := &cloudcontrolv1beta1.Scope{
		Spec: cloudcontrolv1beta1.ScopeSpec{
			Scope: cloudcontrolv1beta1.ScopeInfo{
				Azure: &cloudcontrolv1beta1.AzureScope{
					TenantId:       tenantID,
					SubscriptionId: subscriptionID,
					VpcNetwork:     commonVpcName(state.shootNamespace, state.shootName),
				},
			},
		},
	}

	state.SetObj(scope)

	return nil, nil
}
