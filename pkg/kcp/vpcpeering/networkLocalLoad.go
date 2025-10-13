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

	err := state.Cluster().K8sClient().Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      state.ObjAsVpcPeering().Spec.Details.LocalNetwork.Name,
	}, net)

	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading local KCP Network for KCP VpcPeering", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	if apierrors.IsNotFound(err) {
		if composed.IsMarkedForDeletion(state.Obj()) {
			return composed.LogErrorAndReturn(err, "KCP VpcPeering marked for deletion but, local KCP Network not found", nil, ctx)
		}

		// Patch status was triggered reconciliation immediately. Changing delay to 1 second to reduce wait time.
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
	}

	state.localNetwork = net

	// Ignore state check if marked for deletion
	if composed.IsMarkedForDeletion(state.Obj()) {
		return composed.LogErrorAndReturn(err, "KCP VpcPeering marked for deletion, continue", nil, ctx)
	}

	if net.Status.State == string(cloudcontrolv1beta1.StateError) {
		changed := false

		if meta.RemoveStatusCondition(state.ObjAsVpcPeering().Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
			changed = true
		}

		if meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonWaitingDependency,
			Message: "Local network not ready",
		}) {
			changed = true
		}

		if !changed {
			return composed.StopAndForget, ctx
		}

		return composed.PatchStatus(state.ObjAsVpcPeering()).
			ErrorLogMessage("Error patching KCP VpcPeering status with local network not ready").
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("KCP VpcPeering local KCP Network not ready").
			Run(ctx, state)
	}

	if net.Status.State != string(cloudcontrolv1beta1.StateReady) {
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
	}

	return nil, ctx
}
