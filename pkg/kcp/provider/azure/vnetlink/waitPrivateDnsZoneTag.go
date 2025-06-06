package vnetlink

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	peeringconfig "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func waitPrivateDnsZoneTag(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// If PrivateDnsZone is found but tags don't match, user can recover by adding tag to remote PrivateDnsZone so, we are
	// adding stop with requeue delay of one minute.

	_, hasShootTag := state.privateDnzZone.Tags[peeringconfig.VpcPeeringConfig.NetworkTag]
	if !hasShootTag {
		_, hasShootTag = state.privateDnzZone.Tags[state.Scope().Spec.ShootName]
	}

	if hasShootTag {
		logger.Info("Matching tag found for loaded PrivateDnsZone")
		return nil, ctx
	}

	var kv []any

	for k, v := range state.privateDnzZone.Tags {
		kv = append(kv, k, v)
	}

	logger.Info("Loaded remote PrivateDnsZone has no matching tags", kv...)

	return composed.UpdateStatus(state.ObjAsAzureVNetLink()).
		SetCondition(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonFailedLoadingPrivateDnzZone,
			Message: "Loaded remote PrivateDnsZone has no matching tags",
		}).
		ErrorLogMessage("Error updating KCP AzureVNetLink status due to remote PrivateDnsZone tag mismatch").
		FailedError(composed.StopWithRequeue).
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
		Run(ctx, state)

}
