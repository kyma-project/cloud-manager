package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/util"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func loadKyma(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	kymaUnstructured := util.NewKymaUnstructured()
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Name:      state.Obj().GetName(),
		Namespace: state.Obj().GetNamespace(),
	}, kymaUnstructured)

	if apierrors.IsNotFound(err) {
		logger.Info("Kyma CR does not exist")
		return composed.StopAndForget, nil
	}

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading Kyma CR", composed.StopWithRequeue, nil)
	}

	state.kyma = kymaUnstructured

	logger.Info("Kyma CR loaded")

	return nil, nil
}
