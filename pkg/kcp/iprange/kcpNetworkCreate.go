package iprange

import (
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func kcpNetworkCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.network != nil {
		return nil, ctx
	}

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, ctx
	}

	if state.isKymaNetwork {
		logger.Error(errors.New("network not found"), "Error loading Kyma KCP Network that should exist!!!")

		state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError

		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Kyma network not found",
			}).
			ErrorLogMessage("Error patching KCP IpRange status with specified Kyma network not found").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
			Run(ctx, state)
	}

	if !state.isCloudManagerNetwork {
		logger.Error(errors.New("network not found"), "Error loading non-CM KCP Network")

		state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError

		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Network not found",
			}).
			ErrorLogMessage("Error patching KCP IpRange status with specified non-CM network not found").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
			Run(ctx, state)
	}

	// cloud manager network is specified in the IpRange, but it does not exist, so it should be created
	net := &cloudcontrolv1beta1.Network{}
	net.Name = state.networkKey.Name
	net.Namespace = state.networkKey.Namespace
	net.Spec = cloudcontrolv1beta1.NetworkSpec{
		Scope: cloudcontrolv1beta1.ScopeRef{
			Name: state.Scope().Name,
		},
		Network: cloudcontrolv1beta1.NetworkInfo{
			Managed: &cloudcontrolv1beta1.ManagedNetwork{
				Cidr:     state.ObjAsIpRange().Status.Cidr,
				Location: state.Scope().Spec.Region,
			},
		},
		Type: cloudcontrolv1beta1.NetworkTypeCloudResources,
	}

	logger.Info("Creating KCP Network for IpRange")

	err := state.Cluster().K8sClient().Create(ctx, net)
	if apierrors.IsAlreadyExists(err) {
		return composed.LogErrorAndReturn(err, "Error creating CM KCP Network - already exists", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	if err != nil {
		logger.Error(err, "Error creating CM KCP Network")
		state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Network not found",
			}).
			ErrorLogMessage("Error patching KCP IpRange status failed creating CM network").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	state.network = net

	return nil, nil
}
