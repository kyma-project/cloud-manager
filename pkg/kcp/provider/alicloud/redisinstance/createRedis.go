package redisinstance

import (
	"context"
	"crypto/sha256"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud"
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

	if state.instance != nil {
		return nil, ctx
	}

	kcp := state.ObjAsRedisInstance()

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

	meta.RemoveStatusCondition(kcp.Conditions(), cloudcontrolv1beta1.ConditionTypeError)

	instanceId, lastErr, allZonesFailed := tryCreateInstanceInVSwitches(ctx, state, vSwitchIds, password)

	if lastErr != nil {
		return handleInstanceCreateError(ctx, state, lastErr, allZonesFailed)
	}

	kcp.Status.Id = instanceId
	if err := state.UpdateObjStatus(ctx); err != nil {
		return composed.LogErrorAndReturn(err,
			"Error persisting new AliCloud r-kvstore instance ID",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
}

// tryCreateInstanceInVSwitches tries each vSwitch in turn and returns the new instance ID,
// the last error (nil on success), and whether every zone rejected the request.
func tryCreateInstanceInVSwitches(ctx context.Context, state *State, vSwitchIds []string, password string) (string, error, bool) {
	logger := composed.LoggerFromCtx(ctx)
	kcp := state.ObjAsRedisInstance()

	var instanceId string
	var lastErr error
	allZonesFailed := true

	for _, vSwitchId := range vSwitchIds {
		// "v3" suffix rotates tokens away from v2 tokens that included password.
		// Different ReadOnlyCount values must not share a token — AliCloud would
		// return the existing instance without applying the new replica count.
		tokenInput := fmt.Sprintf("%s%s%s%dv3",
			string(kcp.UID),
			kcp.Spec.Instance.Alicloud.InstanceClass, vSwitchId,
			kcp.Spec.Instance.Alicloud.ReadOnlyCount,
		)
		tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(tokenInput)))[:32] //nolint:gosec

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
			return instanceId, nil, false
		}
		lastErr = err
		if alicloudclient.IsVSwitchZoneErr(err) {
			logger.Info("AliCloud r-kvstore: vSwitch zone not supported for instance class, trying next",
				"vSwitchId", vSwitchId, "instanceClass", kcp.Spec.Instance.Alicloud.InstanceClass)
			continue
		}
		allZonesFailed = false
		break
	}

	return "", lastErr, allZonesFailed
}

// handleInstanceCreateError dispatches a CreateInstance error to the appropriate reconciler response.
func handleInstanceCreateError(ctx context.Context, state *State, err error, allZonesFailed bool) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	kcp := state.ObjAsRedisInstance()

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
	if allZonesFailed {
		// Don't give up permanently — the user may add subnets in a compatible zone later.
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
