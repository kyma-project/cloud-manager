package scope

import (
	"context"
	"errors"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
)

func createScopeAzure(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(State)

	subscriptionID, ok := state.CredentialData()["subscriptionID"]
	if !ok {
		err := errors.New("gardener credential for azure missing subscriptionID key")
		logger.Error(err, "error defining Azure scope")
		return composed.StopAndForget, nil // no requeue
	}

	tenantID, ok := state.CredentialData()["tenantID"]
	if !ok {
		err := errors.New("gardener credential for azure missing tenantID key")
		logger.Error(err, "error defining Azure scope")
		return composed.StopAndForget, nil // no requeue
	}

	// just create the scope with Azure specifics, the ensureScopeCommonFields will set common values
	scope := &cloudresourcesv1beta1.Scope{
		Spec: cloudresourcesv1beta1.ScopeSpec{
			Scope: cloudresourcesv1beta1.ScopeInfo{
				Azure: &cloudresourcesv1beta1.AzureScope{
					TenantId:       tenantID,
					SubscriptionId: subscriptionID,
					VpcNetwork:     fmt.Sprintf("shoot--%s--%s", state.ShootNamespace(), state.ShootName()),
				},
			},
		},
	}

	state.SetScope(scope)

	return nil, nil
}
