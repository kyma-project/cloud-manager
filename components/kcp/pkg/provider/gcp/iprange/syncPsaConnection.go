package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func syncPsaConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	logger.WithValues("ipRange :", ipRange.Name).Info("Saving GCP PSA Connection")

	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork

	switch state.connectionOp {
	case focal.ADD:
		_, err := state.serviceNetworkingClient.CreateServiceConnection(ctx, project, vpc, state.ipRanges)
		if err != nil {
			state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, err)
			return composed.LogErrorAndReturn(err, "Error creating Service Connections in GCP", composed.StopWithRequeue, nil)
		}
	case focal.MODIFY:
		_, err := state.serviceNetworkingClient.PatchServiceConnection(ctx, project, vpc, state.ipRanges)
		if err != nil {
			state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, err)
			return composed.LogErrorAndReturn(err, "Error patching Service Connections in GCP", composed.StopWithRequeue, nil)
		}
	case focal.DELETE:
		_, err := state.serviceNetworkingClient.DeleteServiceConnection(ctx, project, vpc)
		if err != nil {
			state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, err)
			return composed.LogErrorAndReturn(err, "Error deleting Service Connections in GCP", composed.StopWithRequeue, nil)
		}
	}

	return composed.StopWithRequeue, nil
}
