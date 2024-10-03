package iprange

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func privateDnsZoneDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.privateDnsZone == nil {
		return nil, ctx
	}

	logger.Info("Deleting Azure KCP IpRange dnsZone")

	resourceGroupName := state.resourceGroupName
	privateDnsZoneName := state.privateDnsZone.Name

	err := state.azureClient.DeletePrivateDnsZone(ctx, resourceGroupName, ptr.Deref(privateDnsZoneName, ""))
	if azuremeta.IsTooManyRequests(err) {
		return azuremeta.LogErrorAndReturn(err, "Azure KCP IpRange too many requests on delete dnsZone", ctx)
	}
	if err != nil {
		logger.Error(err, "Error deleting Azure KCP IpRange dnsZone")

		state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.ErrorState

		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Error deleting dnsZone",
			}).
			ErrorLogMessage("Error patching Azure KCP IpRange status after failed deleting dnsZone").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	return nil, nil
}
