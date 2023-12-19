package scope

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateScopeRef(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(State)

	state.CommonObj().SetScopeRef(&cloudresourcesv1beta1.ScopeRef{
		Name: state.Scope().Name,
	})

	meta := state.CommonObj().GetObjectMeta()
	metav1.SetMetaDataLabel(meta, cloudresourcesv1beta1.KymaLabel, state.CommonObj().KymaName())
	metav1.SetMetaDataLabel(meta, cloudresourcesv1beta1.ScopeLabel, state.Scope().Name)
	metav1.SetMetaDataLabel(meta, cloudresourcesv1beta1.ProviderTypeLabel, string(state.Provider()))

	err := state.UpdateObj(ctx)
	if err != nil {
		err = fmt.Errorf("error updating object scope ref: %w", err)
		logger.Error(err, "error saving object with Gcp scope ref")
		return composed.StopWithRequeue, nil // will requeue
	}

	return nil, nil
}
