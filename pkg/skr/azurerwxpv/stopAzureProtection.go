package azurerwxpv

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func stopAzureProtection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	//If no protectedItem found, return
	if state.protectedItem == nil {
		return nil, ctx
	}

	//If the protection is already stopped, then continue with deletion of file share
	if *state.protectedItem.ProtectionState == armrecoveryservicesbackup.ProtectionStateProtectionStopped {
		return nil, ctx
	}

	logger.Info("Stop Azure Protection")
	err := state.client.StopFileShareProtection(ctx, state.protectedId, state.protectedItem)
	if err != nil {
		return composed.LogErrorAndReturn(err, "error stopping Azure file share protection", err, ctx)
	}

	//Requeue to wait till the protection is stopped.
	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
