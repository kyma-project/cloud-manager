package defaultiprange

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func findDefaultSkrIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)
	logger := composed.LoggerFromCtx(ctx)

	if state.GetSkrIpRange() != nil {
		return nil, ctx
	}

	skrIpRange := &cloudresourcesv1beta1.IpRange{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Name: "default",
	}, skrIpRange)
	if apierrors.IsNotFound(err) {
		logger.Info("Default SKR IpRange does not exist")
		return nil, ctx
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error getting default SKR IpRange", composed.StopWithRequeue, ctx)
	}

	state.SetSkrIpRange(skrIpRange)

	return nil, ctx
}
