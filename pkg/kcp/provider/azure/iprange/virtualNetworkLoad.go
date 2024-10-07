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

func virtualNetworkLinkLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	if state.virtualNetworkLink != nil {
		logger.Info("Azure virtual network link already loaded")
		return nil, nil
	}
	logger.Info("Loading Azure virtual network link")
	resourceGroupName := state.resourceGroupName
	privateDnsZoneName := azureutil.NewPrivateDnsZoneName(state.ObjAsIpRange().Name)
	virtualNetworkLinkName := state.ObjAsIpRange().Name
	virtualNetworkLinkInstance, err := state.azureClient.GetVirtualNetworkLink(ctx, resourceGroupName, privateDnsZoneName, virtualNetworkLinkName)
	if err != nil {
		if azuremeta.IsNotFound(err) {
			logger.Info("Azure virtual network link instance not found")
			return nil, nil
		}
		logger.Error(err, "Error loading Azure virtual network link")
		meta.SetStatusCondition(state.ObjAsIpRange().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed loading AzureIpRange: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating IpRangeInstance status due failed azure virtual network link loading",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}
	state.virtualNetworkLink = virtualNetworkLinkInstance
	return nil, nil
}
