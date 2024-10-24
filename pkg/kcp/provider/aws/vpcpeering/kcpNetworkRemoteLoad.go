package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func kcpNetworkRemoteLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	net := &cloudcontrolv1beta1.Network{}
	namespace := state.ObjAsVpcPeering().Spec.Details.RemoteNetwork.Namespace
	if namespace == "" {
		namespace = state.ObjAsVpcPeering().Namespace
	}

	err := state.Cluster().K8sClient().Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      state.ObjAsVpcPeering().Spec.Details.RemoteNetwork.Name,
	}, net)

	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading remote KCP Network for KCP VpcPeering", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	if apierrors.IsNotFound(err) {
		if composed.IsMarkedForDeletion(state.Obj()) {
			return composed.LogErrorAndReturn(err, "KCP VpcPeering marked for deletion but, remote KCP Network not found", nil, ctx)
		}

		state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.ErrorState)
		return composed.PatchStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonMissingDependency,
				Message: "Remote network not found",
			}).
			ErrorLogMessage("Error patching KCP VpcPeering status with missing remote network dependency").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			SuccessLogMsg("KCP VpcPeering remote KCP Network not found").
			Run(ctx, state)
	}

	if net.Status.State != string(cloudcontrolv1beta1.ReadyState) {
		state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.ErrorState)
		return composed.PatchStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonWaitingDependency,
				Message: "Remote network not ready",
			}).
			ErrorLogMessage("Error patching KCP VpcPeering status with remote network not ready").
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg("KCP VpcPeering remote KCP Network not ready").
			Run(ctx, state)
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.WithValues(
		"remoteNetwork", net.Name,
	)
	ctx = composed.LoggerIntoCtx(ctx, logger)
	state.remoteNetwork = net

	logger.Info("KCP VpcPeeing remote network loaded")

	return nil, ctx
}
