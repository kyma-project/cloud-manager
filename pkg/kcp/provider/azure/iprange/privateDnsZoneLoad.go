package iprange

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func privateDnsZoneLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	if state.privateDnsZone != nil {
		logger.Info("Azure Private DnsZone already loaded")
		return nil, nil
	}
	logger.Info("Loading Azure Private DnsZone")
	resourceGroupName := state.resourceGroupName
	privateDnsZoneName := azureutil.NewPrivateDnsZoneName(state.ObjAsIpRange().Name)
	privateDnsZoneInstance, err := state.azureClient.GetPrivateDnsZone(ctx, resourceGroupName, privateDnsZoneName)
	if err != nil {
		if azuremeta.IsNotFound(err) {
			logger.Info("Azure Private DnsZone instance not found")
			return nil, nil
		}
		logger.Error(err, "Error loading Azure Private DnsZone")
		meta.SetStatusCondition(state.ObjAsIpRange().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed loading AzureIpRange: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating IpRangeInstance status due failed azure Private DnsZone loading",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}
	state.privateDnsZone = privateDnsZoneInstance
	return nil, nil
}
