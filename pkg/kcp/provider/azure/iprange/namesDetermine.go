package iprange

import (
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azurecommon "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func namesDetermine(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	switch state.Network().Spec.Type {
	case cloudcontrolv1beta1.NetworkTypeCloudResources:
		state.resourceGroupName = azurecommon.AzureCloudManagerResourceGroupName(state.Scope().Spec.Scope.Azure.VpcNetwork)
		state.virtualNetworkName = state.resourceGroupName
		state.securityGroupName = state.resourceGroupName
		state.subnetName = state.resourceGroupName
	default:
		logger.
			WithValues(
				"networkName", state.Network().Name,
				"networkType", state.Network().Spec.Type,
			).
			Error(errors.New("invalid network type"), "Azure IpRange can be created on CM network only")
		ctx = composed.LoggerIntoCtx(ctx, logger)

		state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError

		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Unsupported network type",
			}).
			ErrorLogMessage("Error patching Azure KCP IpRange with error status for unsupported network type").
			Run(ctx, state)
	}

	logger = logger.WithValues(
		"resourceGroupName", state.resourceGroupName,
		"virtualNetworkName", state.virtualNetworkName,
		"securityGroupName", state.securityGroupName,
		"subnetName", state.subnetName,
	)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	return nil, ctx
}
