package iprange

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func privateDnsZoneCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.privateDnsZone != nil {
		return nil, ctx
	}

	logger.Info("Creating Azure KCP IpRange privateDnsZone")

	privateDnsZoneName := azureutil.NewPrivateDnsZoneName()

	err := state.azureClient.CreatePrivateDnsZone(ctx, state.resourceGroupName, privateDnsZoneName, nil)

	if azuremeta.IsTooManyRequests(err) {
		return composed.LogErrorAndReturn(err,
			"Azure KCP IpRange too many requests on private dns zone create",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()),
			ctx,
		)
	}

	if err != nil {
		logger.Error(err, "Error creating Azure KCP IpRange privateDnsZone")

		state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError

		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Error creating Azure privateDnsZone",
			}).
			ErrorLogMessage("Error patching Azure KCP IpRange status after failed creating privateDnsZone").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
			Run(ctx, state)
	}

	return nil, nil
}
