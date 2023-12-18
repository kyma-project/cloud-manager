package scope

import (
	"context"
	"errors"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	scope := &cloudresourcesv1beta1.Scope{
		ObjectMeta: metav1.ObjectMeta{
			Name:      state.Obj().GetName(),
			Namespace: state.Obj().GetNamespace(),
			Labels: map[string]string{
				cloudresourcesv1beta1.ScopeKymaLabel: state.CommonObj().KymaName(),
			},
		},
		Spec: cloudresourcesv1beta1.ScopeSpec{
			Kyma:      "",
			ShootName: "",
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
