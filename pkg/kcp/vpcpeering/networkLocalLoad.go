package vpcpeering

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func kcpNetworkLocalLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	net := &cloudcontrolv1beta1.Network{}

	namespace := state.ObjAsVpcPeering().Namespace

	logger = logger.
		WithValues("localKcpNetwork", fmt.Sprintf("%s/%s", namespace, state.ObjAsVpcPeering().Spec.Details.LocalNetwork.Name))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	err := state.Cluster().K8sClient().Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      state.ObjAsVpcPeering().Spec.Details.LocalNetwork.Name,
	}, net)

	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading local KCP Network for KCP VpcPeering", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	if apierrors.IsNotFound(err) {
		changed := false

		if meta.RemoveStatusCondition(state.ObjAsVpcPeering().Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
			changed = true
		}

		if meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonMissingDependency,
			Message: "Local network not found",
		}) {
			changed = true
		}

		if state.ObjAsVpcPeering().Status.State != string(cloudcontrolv1beta1.StateError) {
			state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateError)
			changed = true
		}

		if !changed {
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
		}

		return composed.PatchStatus(state.ObjAsVpcPeering()).
			ErrorLogMessage("Error patching KCP VpcPeering status with local network not found").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			SuccessLogMsg("KCP VpcPeering local KCP Network not found").
			Run(ctx, state)
	}

	state.localNetwork = net

	return nil, ctx
}
