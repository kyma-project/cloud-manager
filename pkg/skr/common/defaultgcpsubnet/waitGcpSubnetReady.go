package defaultgcpsubnet

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
)

func waitGcpSubnetReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)

	isReady := meta.IsStatusConditionTrue(state.GetSkrGcpSubnet().Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	if isReady {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.
		WithValues("GcpSubnet", fmt.Sprintf("%s/%s", state.GetSkrGcpSubnet().Namespace, state.GetSkrGcpSubnet().Name)).
		Info("GcpSubnet is not ready, requeue delayed")

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
}
