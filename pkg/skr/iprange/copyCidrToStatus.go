package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func copyCidrToStatus(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)
	obj := state.ObjAsIpRange()
	if len(obj.Status.Cidr) > 0 {
		return nil, nil
	}

	logger.Info("Copying IpRange Cidr to status field")

	obj.Status.Cidr = obj.Spec.Cidr
	err := state.UpdateObjStatus(ctx)
	if err != nil {
		logger.Error(err, "Error updating IpRange status with cidr")
		return composed.StopWithRequeue, nil
	}

	return nil, nil
}
