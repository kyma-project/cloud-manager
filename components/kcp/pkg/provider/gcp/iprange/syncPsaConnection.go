package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
	"google.golang.org/api/servicenetworking/v1"
)

func syncPsaConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	logger.WithValues("ipRange :", ipRange.Name).Info("Saving GCP PSA Connection")

	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork

	var operation *servicenetworking.Operation
	var err error

	switch state.connectionOp {
	case client.ADD:
		operation, err = state.serviceNetworkingClient.CreateServiceConnection(ctx, project, vpc, state.ipRanges)
	case client.MODIFY:
		operation, err = state.serviceNetworkingClient.PatchServiceConnection(ctx, project, vpc, state.ipRanges)
	case client.DELETE:
		operation, err = state.serviceNetworkingClient.DeleteServiceConnection(ctx, state.projectNumber, vpc)
	}

	if err != nil {
		state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, err)
		return composed.LogErrorAndReturn(err, "Error syncronizing Service Connections in GCP", composed.StopWithRequeueDelay(client.GcpRetryWaitTime), nil)
	}
	if operation != nil {
		ipRange.Status.OpIdentifier = operation.Name
		state.UpdateObjStatus(ctx)
	}
	return composed.StopWithRequeueDelay(client.GcpOperationWaitTime), nil
}
