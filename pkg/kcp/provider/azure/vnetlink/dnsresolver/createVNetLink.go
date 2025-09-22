package dnsresolver

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
)

func createVNetLink(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vnetLink != nil {
		return nil, ctx
	}

	kymaNetworkName := state.Scope().Spec.Scope.Azure.VpcNetwork
	kymaResourceGroup := state.Scope().Spec.Scope.Azure.VpcNetwork
	vnetId := azureutil.NewVirtualNetworkResourceId(state.Scope().Spec.Scope.Azure.SubscriptionId,
		kymaResourceGroup, kymaNetworkName).String()

	err := state.remoteClient.CreateDnsResolverVNetLink(ctx,
		state.rulesetId.ResourceGroup,
		state.rulesetId.ResourceName,
		state.ObjAsAzureVNetLink().Spec.RemoteVNetLinkName,
		vnetId)

	if err == nil {
		logger.Info("DNS resolver VirtualNetworkLink created")
		return nil, ctx
	}

	logger.Error(err, "Error creating DNS resolver VirtualNetworkLink")

	return azuremeta.HandleError(err, state.ObjAsAzureVNetLink()).
		WithDefaultReason(cloudcontrolv1beta1.ReasonFailedCreatingVirtualNetworkLink).
		WithDefaultMessage("Failed creating DNS resolver VirtualNetworkLink").
		WithTooManyRequestsMessage("Too many requests on creating DNS resolver VirtualNetworkLink").
		WithUpdateStatusMessage("Error updating KCP AzureVNetLink status on failed creating of DNS resolver VirtualNetworkLink").
		Run(ctx, state)

}
