package v2

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// modifyCapacityGb checks if capacity needs to be updated and adds to updateMask.
// Follows the RedisInstance pattern for granular field modification tracking.
func modifyCapacityGb(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	nfsInstance := state.ObjAsNfsInstance()

	if state.GetInstance() == nil {
		return composed.StopWithRequeue, nil
	}

	if len(state.GetInstance().FileShares) == 0 {
		return nil, ctx
	}

	currentCapacityGb := state.GetInstance().FileShares[0].CapacityGb
	desiredCapacityGb := int64(nfsInstance.Spec.Instance.Gcp.CapacityGb)

	if currentCapacityGb == desiredCapacityGb {
		return nil, ctx
	}

	state.UpdateCapacityGb(desiredCapacityGb)

	return nil, ctx
}
