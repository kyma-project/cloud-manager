package iprange

import (
	"context"
	"errors"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/provider/gcp/client"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
	"google.golang.org/api/compute/v1"
)

func syncAddress(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	logger.WithValues("ipRange :", ipRange.Name).Info("Saving GCP Address")

	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork
	name := ipRange.Spec.RemoteRef.Name

	var operation *compute.Operation
	var err error
	switch state.addressOp {
	case client.ADD:
		operation, err = state.computeClient.CreatePscIpRange(ctx, project, vpc, name, name, state.ipAddress, int64(state.prefix))
	case client.MODIFY:
		err = errors.New("IpRange update not supported.")
		state.AddErrorCondition(ctx, v1beta1.ReasonNotSupported, err)
		return composed.LogErrorAndReturn(err, "IpRange update not supported.", composed.StopAndForget, nil)
	case client.DELETE:
		operation, err = state.computeClient.DeleteIpRange(ctx, project, name)
	}

	if err != nil {
		state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, err)
		return composed.LogErrorAndReturn(err, "Error synchronizing Address object in GCP", composed.StopWithRequeueDelay(client.GcpRetryWaitTime), nil)
	}
	if operation != nil {
		ipRange.Status.OpIdentifier = operation.Name
		state.UpdateObjStatus(ctx)
	}
	return composed.StopWithRequeueDelay(client.GcpOperationWaitTime), nil
}
