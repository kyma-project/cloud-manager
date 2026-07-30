package rediscluster

import (
	"context"
	"crypto/sha256"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloud "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud"
	alicloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.instance != nil {
		return nil, ctx
	}

	kcp := state.ObjAsRedisCluster()
	if kcp.Spec.Instance.Alicloud == nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("spec.instance.alicloud is nil"),
			"AliCloud rediscluster without alicloud provider spec",
			composed.StopAndForget, ctx)
	}
	if state.IpRange() == nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("ipRange is nil"),
			"AliCloud rediscluster requires resolved IpRange",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	var vSwitchIds []string
	for _, sn := range state.IpRange().Status.Subnets {
		if sn.Id != "" {
			vSwitchIds = append(vSwitchIds, sn.Id)
		}
	}
	if len(vSwitchIds) == 0 {
		return composed.LogErrorAndReturn(
			fmt.Errorf("no vSwitch found in IpRange subnets"),
			"AliCloud rediscluster IpRange has no vSwitch",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	// Generate password before CreateInstance - AliCloud never returns it after.
	// Persist it before calling CreateInstance so a crash after the API call but
	// before status write does not lose the password on the next retry.
	password := kcp.Status.AuthString
	if password == "" {
		password = alicloud.GeneratePassword()
		kcp.Status.AuthString = password
		if err := state.UpdateObjStatus(ctx); err != nil {
			return composed.LogErrorAndReturn(err,
				"Error persisting AliCloud r-kvstore cluster auth string before create",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
	}

	// Try each vSwitch in turn. Some instance classes are only available in
	// specific zones; AliCloud returns InvalidvSwitchId when the zone does not
	// support the requested class.
	var instanceId string
	var lastErr error
	allZonesFailed := true
	for _, vSwitchId := range vSwitchIds {
		// "v4" suffix rotates tokens away from v3 tokens that omitted ShardCount
		// and ReplicasPerShard. Different shard/replica configs must not share a token.
		tokenInput := fmt.Sprintf("%s%s%s%s%d%dv4",
			string(kcp.UID), password,
			kcp.Spec.Instance.Alicloud.InstanceClass, vSwitchId,
			kcp.Spec.Instance.Alicloud.ShardCount,
			kcp.Spec.Instance.Alicloud.ReplicasPerShard,
		)
		tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(tokenInput)))[:32]

		opts := alicloudclient.CreateInstanceOptions{
			InstanceName:  kcp.Name,
			InstanceClass: kcp.Spec.Instance.Alicloud.InstanceClass,
			EngineVersion: kcp.Spec.Instance.Alicloud.EngineVersion,
			VpcId:         state.IpRange().Status.VpcId,
			VSwitchId:     vSwitchId,
			Password:      password,
			ShardCount:    kcp.Spec.Instance.Alicloud.ShardCount,
			ReadOnlyCount: kcp.Spec.Instance.Alicloud.ReplicasPerShard,
			Token:         tokenHash,
		}
		var err error
		instanceId, err = state.client.CreateInstance(ctx, opts)
		if err == nil {
			lastErr = nil
			allZonesFailed = false
			break
		}
		lastErr = err
		if alicloudclient.IsVSwitchZoneErr(err) {
			logger.Info("AliCloud r-kvstore cluster: vSwitch zone not supported, trying next", "vSwitchId", vSwitchId, "instanceClass", kcp.Spec.Instance.Alicloud.InstanceClass)
			continue
		}
		allZonesFailed = false
		break
	}

	if lastErr != nil {
		err := lastErr
		logger.Error(err, "Error creating AliCloud r-kvstore cluster instance")
		meta.SetStatusCondition(kcp.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonFailedCreatingRedisCluster,
			Message: fmt.Sprintf("Failed creating AlicloudRedisCluster: %s", err),
		})
		if updErr := state.UpdateObjStatus(ctx); updErr != nil {
			return composed.LogErrorAndReturn(updErr,
				"Error updating RedisCluster status after failed CreateInstance",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
		if allZonesFailed {
			return composed.StopWithRequeueDelay(util.Timing.T300000ms()), ctx
		}
		if alicloudclient.IsPermanentError(err) {
			if alicloudclient.IsPasswordErr(err) {
				kcp.Status.AuthString = ""
				if updErr := state.UpdateObjStatus(ctx); updErr != nil {
					logger.Error(updErr, "Error clearing invalid password from status")
				}
			}
			return composed.StopAndForget, ctx
		}
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	}

	kcp.Status.Id = instanceId
	if err := state.UpdateObjStatus(ctx); err != nil {
		return composed.LogErrorAndReturn(err,
			"Error persisting new AliCloud r-kvstore cluster instance ID",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
}
