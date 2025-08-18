package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/redisinstance/client"
)

func modifyMaintenancePolicy(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisInstance := state.ObjAsRedisInstance()

	if state.gcpRedisInstance == nil {
		return composed.StopWithRequeue, nil
	}

	currentPolicy := state.gcpRedisInstance.MaintenancePolicy
	desiredPolicy := client.ToMaintenancePolicy(redisInstance.Spec.Instance.Gcp.MaintenancePolicy)

	if AreEqualPolicies(currentPolicy, desiredPolicy) {
		return nil, ctx
	}

	state.UpdateMaintenancePolicy(desiredPolicy)

	return nil, ctx
}
