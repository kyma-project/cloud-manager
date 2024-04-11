package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func loadKyma(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	kymaUnstructured := util.NewKymaUnstructured()
	err := state.Cluster().K8sClient().Get(ctx, state.Name(), kymaUnstructured)

	if apierrors.IsNotFound(err) {
		logger.Info("Kyma CR does not exist")
		return composed.StopAndForget, nil
	}

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading Kyma CR", composed.StopWithRequeue, ctx)
	}

	state.kyma = kymaUnstructured

	return nil, nil
}
