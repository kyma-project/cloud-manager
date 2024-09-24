package vpcpeering

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadLocalNetwork(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	net := &cloudcontrolv1beta1.Network{}
	namespace := state.ObjAsVpcPeering().Spec.Details.LocalNetwork.Namespace
	if namespace == "" {
		namespace = state.ObjAsVpcPeering().Namespace
	}

	logger = logger.
		WithValues("localKcpNetwork", fmt.Sprintf("%s/%s", namespace, state.ObjAsVpcPeering().Spec.Details.LocalNetwork.Name))

	err := state.Cluster().K8sClient().Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      state.ObjAsVpcPeering().Spec.Details.LocalNetwork.Name,
	}, net)

	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "[KCP GCP VPCPeering loadLocalNetwork] Error loading GCP KCP Local Network", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	if apierrors.IsNotFound(err) {
		state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.ErrorState)
		return composed.PatchStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonMissingDependency,
				Message: "Local network not found",
			}).
			ErrorLogMessage("Error patching KCP VpcPeering status with missing local network dependency").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			SuccessLogMsg("KCP VpcPeering local KCP Network not found").
			Run(ctx, state)
	}

	if net.Status.State != string(cloudcontrolv1beta1.ReadyState) {
		state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.ErrorState)
		return composed.PatchStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonWaitingDependency,
				Message: "Local network is not ready",
			}).
			ErrorLogMessage("Error patching KCP VpcPeering status with local network not ready").
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg("KCP VpcPeering local Network is not ready").
			Run(ctx, state)
	}

	ctx = composed.LoggerIntoCtx(ctx, logger)
	state.localNetwork = net

	logger.Info("[SKR GCP VPCPeering createKcpRemoteNetwork] GCP KCP VpcPeering local network loaded successfully")

	return nil, ctx
}
