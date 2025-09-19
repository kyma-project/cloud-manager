package dnszone

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"

	"k8s.io/utils/ptr"
)

func deleteVNetLink(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vnetLink == nil {
		logger.Info("AzureVNetLink deleted before VirtualNetworkLink is created")
		return nil, ctx
	}

	resourceId, err := azureutil.ParseResourceID(ptr.Deref(state.vnetLink.ID, ""))

	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Failed parsing vnetLink.ID while deleting AzureVNetLink", ctx)
	}

	logger.Info("Deleting VirtualNetworkLink")

	err = state.remoteClient.DeleteVirtualNetworkLink(
		ctx,
		resourceId.ResourceGroup,
		resourceId.ResourceName,
		resourceId.SubResourceName,
	)

	if err == nil {
		logger.Info("VirtualNetworkLink deleted")
		return nil, ctx
	}

	if azuremeta.IsTooManyRequests(err) {
		return composed.LogErrorAndReturn(err,
			"Too many requests on loading VirtualNetworkLink",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()),
			ctx,
		)
	}

	return azuremeta.LogErrorAndReturn(err, "Error deleting VirtualNetworkLink", ctx)

}
