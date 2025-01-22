package iprange

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func loadKcpIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	kcpIpRange := &cloudcontrolv1beta1.IpRange{}
	err := state.KcpCluster.K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.KymaRef.Namespace,
		Name:      state.ObjAsIpRange().Status.Id,
	}, kcpIpRange)
	if apierrors.IsNotFound(err) {
		logger.Info("KCP IpRange not found")
		return nil, nil
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP IpRange", composed.StopWithRequeue, ctx)
	}

	state.KcpIpRange = kcpIpRange
	return nil, nil
}
