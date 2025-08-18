package defaultgcpsubnet

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func findDefaultSkrGcpSubnet(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)
	logger := composed.LoggerFromCtx(ctx)

	if state.GetSkrGcpSubnet() != nil {
		return nil, ctx
	}

	skrGcpSubnet := &cloudresourcesv1beta1.GcpSubnet{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Name: "default",
	}, skrGcpSubnet)
	if apierrors.IsNotFound(err) {
		logger.Info("Default SKR GcpSubnet does not exist")
		return nil, ctx
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error getting default SKR GcpSubnet", composed.StopWithRequeue, ctx)
	}

	logger.Info("Loaded default SKR GcpSubnet")
	state.SetSkrGcpSubnet(skrGcpSubnet)

	return nil, ctx
}
