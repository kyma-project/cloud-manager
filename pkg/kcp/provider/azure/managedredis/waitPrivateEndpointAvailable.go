package managedredis

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"k8s.io/utils/ptr"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitPrivateEndpointAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.privateEndpoint == nil {
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	if state.privateEndpoint.Properties != nil &&
		ptr.Deref(state.privateEndpoint.Properties.ProvisioningState, "") == armnetwork.ProvisioningStateSucceeded {
		return nil, ctx
	}

	var provisioningState string
	if state.privateEndpoint.Properties != nil && state.privateEndpoint.Properties.ProvisioningState != nil {
		provisioningState = string(*state.privateEndpoint.Properties.ProvisioningState)
	}
	composed.LoggerFromCtx(ctx).Info("Azure Managed Redis Private Endpoint is not ready yet",
		"provisioningState", provisioningState)

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
