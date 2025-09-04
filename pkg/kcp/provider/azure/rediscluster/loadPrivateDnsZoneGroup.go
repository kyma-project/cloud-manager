package rediscluster

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func loadPrivateDnsZoneGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	if state.privateDnsZoneGroup != nil {
		logger.Info("Azure Private DnsZone Group already loaded")
		return nil, ctx
	}
	if state.privateEndPoint == nil {
		logger.Info("Skipping Azure Private DnsZone Group loading, Private EndPoint is not loaded")
		return nil, ctx
	}
	logger.Info("Loading Azure Private DnsZone Group")
	resourceGroupName := state.resourceGroupName
	privateEndPointName := ptr.Deref(state.privateEndPoint.Name, "")
	privateDnsZoneGroupName := state.ObjAsRedisCluster().Name
	privateDnsZoneGroupCluster, err := state.client.GetPrivateDnsZoneGroup(ctx, resourceGroupName, privateEndPointName, privateDnsZoneGroupName)
	if err != nil {
		if azuremeta.IsNotFound(err) {
			logger.Info("Azure Private DnsZone Group Cluster not found")
			return nil, ctx
		}
		logger.Error(err, "Error loading Azure Private DnsZone Group")
		meta.SetStatusCondition(state.ObjAsRedisCluster().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed loading AzureRedis: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisCluster status due failed azure Private DnsZone Group loading",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}
	state.privateDnsZoneGroup = privateDnsZoneGroupCluster
	return nil, ctx
}
