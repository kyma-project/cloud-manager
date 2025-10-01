package dnszone

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"k8s.io/utils/ptr"
)

func loadPrivateDnsZone(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	privateDnsZone, err := state.remoteClient.GetPrivateDnsZone(ctx,
		state.remotePrivateDnsZoneId.ResourceGroup,
		state.remotePrivateDnsZoneId.ResourceName)

	if err == nil {
		ctx = composed.LoggerIntoCtx(ctx, logger.WithValues("privateDnsZoneId", ptr.Deref(privateDnsZone.ID, "")))
		state.privateDnzZone = privateDnsZone
		return nil, ctx
	}

	return azuremeta.HandleError(err, state.ObjAsAzureVNetLink()).
		WithDefaultReason(cloudcontrolv1beta1.ReasonFailedLoadingPrivateDnzZone).
		WithDefaultMessage("Failed loading PrivateDnsZone").
		WithTooManyRequestsMessage("Too many requests on loading PrivateDnsZone").
		WithUpdateStatusMessage("Error updating KCP AzureVNetLink status on failed loading of PrivateDnsZone").
		WithNotFoundMessage("PrivateDnsZone not found").
		Run(ctx, state)
}
