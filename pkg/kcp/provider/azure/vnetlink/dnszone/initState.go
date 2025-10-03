package dnszone

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func initState(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	resourceId, err := azureutil.ParseResourceID(state.ObjAsAzureVNetLink().Spec.RemotePrivateDnsZone)

	if err == nil {
		state.remotePrivateDnsZoneId = resourceId
		return nil, ctx
	}

	logger.Error(err, "Error parsing RemotePrivateDnsZone")

	return composed.PatchStatus(state.ObjAsAzureVNetLink()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonValidationFailed,
			Message: "Error parsing RemotePrivateDnsZone",
		}).
		ErrorLogMessage("Error patching KCP AzureVNetLink with error state after parsing RemotePrivateDnsZone failed").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)

}
