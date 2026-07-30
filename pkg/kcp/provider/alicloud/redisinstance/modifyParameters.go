package redisinstance

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// modifyParameters calls ModifyInstanceConfig when the desired parameters in
// the KCP spec diverge from the current config persisted on the instance.
// The AliCloud API accepts the config as a JSON object string.
// When desired is empty and the cloud instance has custom config, sends "{}"
// to clear it.
func modifyParameters(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.instance == nil {
		return nil, ctx
	}
	kcp := state.ObjAsRedisInstance()
	if kcp.Spec.Instance.Alicloud == nil {
		return nil, ctx
	}
	desired := kcp.Spec.Instance.Alicloud.Parameters

	// Parse current config from instance. AliCloud returns the full config object
	// including system defaults.
	currentFull := map[string]string{}
	if state.instance.Config != "" {
		raw := map[string]interface{}{}
		if err := json.Unmarshal([]byte(state.instance.Config), &raw); err != nil {
			return composed.LogErrorAndReturn(err,
				"Error parsing AliCloud r-kvstore instance config JSON; will retry",
				composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
		}
		for k, v := range raw {
			currentFull[k] = fmt.Sprintf("%v", v)
		}
	}

	// Fast path: desired is empty — clear cloud config only if there is something to clear.
	if len(desired) == 0 {
		if len(currentFull) == 0 {
			return nil, ctx
		}
		if err := state.client.ModifyInstanceConfig(ctx, state.instance.InstanceId, "{}"); err != nil {
			return composed.LogErrorAndReturn(err,
				"Error clearing AliCloud r-kvstore instance config",
				composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
		}
		state.instance = nil
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	}

	// desired is non-empty: project current onto desired keys before comparing
	// to avoid a perpetual modify loop caused by system-default keys in currentFull.
	current := map[string]string{}
	for k := range desired {
		if v, ok := currentFull[k]; ok {
			current[k] = v
		}
	}
	if maps.Equal(desired, current) {
		return nil, ctx
	}

	configBytes, err := json.Marshal(desired)
	if err != nil {
		return composed.LogErrorAndReturn(err,
			"Error marshalling AliCloud r-kvstore instance parameters",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	if err := state.client.ModifyInstanceConfig(ctx, state.instance.InstanceId, string(configBytes)); err != nil {
		return composed.LogErrorAndReturn(err,
			"Error modifying AliCloud r-kvstore instance config",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	// Force re-load next reconcile to pick up the updated Config field.
	state.instance = nil
	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
