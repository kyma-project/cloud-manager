package managedredis

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redisenterprise/armredisenterprise/v3"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitManagedRedisAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.managedRedis == nil {
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	if state.managedRedis.Properties == nil ||
		state.managedRedis.Properties.ProvisioningState == nil ||
		*state.managedRedis.Properties.ProvisioningState != armredisenterprise.ProvisioningStateSucceeded {
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return nil, ctx
}
