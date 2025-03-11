package rediscluster

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureMeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func deletePrivateDnsZoneGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	if state.privateDnsZoneGroup == nil {
		return nil, nil
	}
	if *state.privateDnsZoneGroup.Properties.ProvisioningState == "Deleting" {
		return nil, nil
	}
	logger.Info("Deleting Azure PrivateDnsZoneGroup")
	resourceGroupName := state.resourceGroupName
	privateEndPointName := ptr.Deref(state.privateEndPoint.Name, "")
	privateDnsZoneGroupName := ptr.Deref(state.privateDnsZoneGroup.Name, "")
	err := state.client.DeletePrivateDnsZoneGroup(ctx, resourceGroupName, privateEndPointName, privateDnsZoneGroupName)
	if err != nil {
		if azureMeta.IsNotFound(err) {
			return nil, nil
		}
		logger.Error(err, "Error deleting Azure PrivateDnsZoneGroup")
		meta.SetStatusCondition(state.ObjAsRedisCluster().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed deleting Azure PrivateDnsZoneGroup: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisCluster status due failed azure PrivateDnsZoneGroup deleting",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}
	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
