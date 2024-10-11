package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func loadRemoteNetwork(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	remoteNetwork := &cloudcontrolv1beta1.Network{}
	namespace := state.ObjAsVpcPeering().Spec.Details.RemoteNetwork.Namespace
	if namespace == "" {
		namespace = state.ObjAsVpcPeering().Namespace
	}

	err := state.Cluster().K8sClient().Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      state.ObjAsVpcPeering().Spec.Details.RemoteNetwork.Name,
	}, remoteNetwork)

	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "[KCP GCP VPCPeering loadRemoteNetwork] Error loading remote KCP Network for GCP VpcPeering", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	if apierrors.IsNotFound(err) {
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
			SuccessLogMsg("KCP VpcPeering remote KCP Network was not found").
			Run(ctx, state)
	}

	if remoteNetwork.Status.State != string(cloudcontrolv1beta1.ReadyState) {
		if !remoteNetwork.ObjectMeta.CreationTimestamp.After(state.ObjAsVpcPeering().ObjectMeta.CreationTimestamp.Time.Add(10 * time.Minute)) {
			return composed.LogErrorAndReturn(err, "[KCP GCP VPCPeering loadRemoteNetwork] KCP Remote Network is not ready yet, requeuing", composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil)
		}

		logger.Info("[KCP GCP VPCPeering loadRemoteNetwork] GCP KCP VpcPeering Remote Network didn't reach ready state after 10 minutes, setting error state")
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
			SuccessLogMsg("KCP VpcPeering remote KCP Network is not ready").
			Run(ctx, state)
	}

	state.remoteNetwork = remoteNetwork

	logger.Info("[KCP GCP VPCPeering loadRemoteNetwork] GCP KCP VpcPeering Remote Network loaded successfully")

	return nil, ctx
}
