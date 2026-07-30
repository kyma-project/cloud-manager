package redisinstance

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

// createRedis provisions a new r-kvstore instance if one does not yet exist.
// The password is generated here and stored immediately on Status.AuthString
// because AliCloud never returns it after CreateInstance (design decision 6).
func createRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.instance != nil {
		return nil, ctx
	}

	kcp := state.ObjAsRedisInstance()
	if kcp.Spec.Instance.Alicloud == nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("spec.instance.alicloud is nil"),
			"AliCloud redisinstance without alicloud provider spec",
			composed.StopAndForget, ctx)
	}
	if state.IpRange() == nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("ipRange is nil"),
			"AliCloud redisinstance requires resolved IpRange",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	// Collect candidate vSwitch IDs from all IpRange subnets.
	var vSwitchIds []string
	for _, sn := range state.IpRange().Status.Subnets {
		if sn.Id != "" {
			vSwitchIds = append(vSwitchIds, sn.Id)
		}
	}
	if len(vSwitchIds) == 0 {
		return composed.LogErrorAndReturn(
			fmt.Errorf("no vSwitch found in IpRange subnets"),
			"AliCloud redisinstance IpRange has no vSwitch",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	// Generate password before CreateInstance - AliCloud never returns it after.
	// Persist it before calling CreateInstance so a crash after the API call but
	// before status write does not lose the password on the next retry (the
	// idempotency Token returns the same instance; we must not regenerate).
	password := kcp.Status.AuthString
	if password == "" {
		password = alicloud.GeneratePassword()
		kcp.Status.AuthString = password
		if err := state.UpdateObjStatus(ctx); err != nil {
			return composed.LogErrorAndReturn(err,
				"Error persisting AliCloud r-kvstore instance auth string before create",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
	}

	// Try each vSwitch in turn. Some instance classes are only available in
	// specific zones; AliCloud returns InvalidvSwitchId when the zone does not
	// support the requested class. Iterating all subnets lets the reconciler
	// find a compatible zone without requiring the user to specify one.
	var instanceId string
	var lastErr error
	allZonesFailed := true
	for _, vSwitchId := range vSwitchIds {
		// "v2" suffix rotates tokens away from v1 tokens that omitted ReadOnlyCount.
		// Different ReadOnlyCount values must not share a token — AliCloud would
		// return the existing instance without applying the new replica count.
		tokenInput := fmt.Sprintf("%s%s%s%s%dv2",
			string(kcp.UID), password,
			kcp.Spec.Instance.Alicloud.InstanceClass, vSwitchId,
			kcp.Spec.Instance.Alicloud.ReadOnlyCount,
		)
		tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(tokenInput)))[:32]

		opts := alicloudclient.CreateInstanceOptions{
			InstanceName:  kcp.Name,
			InstanceClass: kcp.Spec.Instance.Alicloud.InstanceClass,
			EngineVersion: kcp.Spec.Instance.Alicloud.EngineVersion,
			VpcId:         state.IpRange().Status.VpcId,
			VSwitchId:     vSwitchId,
			Password:      password,
			ReadOnlyCount: kcp.Spec.Instance.Alicloud.ReadOnlyCount,
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
			logger.Info("AliCloud r-kvstore: vSwitch zone not supported for instance class, trying next", "vSwitchId", vSwitchId, "instanceClass", kcp.Spec.Instance.Alicloud.InstanceClass)
			continue
		}
		// Non-zone error - stop iterating and handle below.
		allZonesFailed = false
		break
	}

	if lastErr != nil {
		err := lastErr
		logger.Error(err, "Error creating AliCloud r-kvstore instance")
		meta.SetStatusCondition(kcp.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonFailedCreatingRedisInstance,
			Message: fmt.Sprintf("Failed creating AlicloudRedis: %s", err),
		})
		if updErr := state.UpdateObjStatus(ctx); updErr != nil {
			return composed.LogErrorAndReturn(updErr,
				"Error updating RedisInstance status after failed CreateInstance",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
		// When every zone rejected the instance class, don't give up permanently -
		// the user may add subnets in a compatible zone later.
		if allZonesFailed {
			return composed.StopWithRequeueDelay(util.Timing.T300000ms()), ctx
		}
		if alicloudclient.IsPermanentError(err) {
			if alicloudclient.IsPasswordErr(err) {
				// Clear authString so the next reconcile generates a fresh password.
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
			"Error persisting new AliCloud r-kvstore instance ID",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
}
