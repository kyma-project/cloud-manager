package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
)

func checkUpdateMask(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	nfsInstance := state.ObjAsNfsInstance()
	logger.WithValues("NfsInstance :", nfsInstance.Name).Info("Checking for Update Mask")

	//If the operation is not modify, continue.
	if state.operation != client.MODIFY {
		return nil, nil
	}

	//If capacity is increased, add it to updateMask
	if nfsInstance.Spec.Instance.Gcp.CapacityGb > int(state.fsInstance.FileShares[0].CapacityGb) {
		state.updateMask = append(state.updateMask, "FileShares")
	}
	return nil, nil
}
