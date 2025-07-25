package subnet

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadNetwork(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	net := &cloudcontrolv1beta1.Network{}
	err := state.Cluster().K8sClient().Get(ctx, client.ObjectKey{
		Namespace: state.ObjAsGcpSubnet().Namespace,
		Name:      common.KcpNetworkKymaCommonName(state.Scope().Name),
	}, net)

	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading Kyma network", composed.StopWithRequeue, ctx)
	}

	if apierrors.IsNotFound(err) {
		logger.Info("Kyma network does not exist")
		state.ObjAsGcpSubnet().Status.State = cloudcontrolv1beta1.StateError
		return composed.UpdateStatus(state.ObjAsGcpSubnet()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Kyma network does not exist",
			}).
			ErrorLogMessage("Error updating KCP GcpSubnet status with kyma network does not exist").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	state.network = net

	return nil, nil
}
