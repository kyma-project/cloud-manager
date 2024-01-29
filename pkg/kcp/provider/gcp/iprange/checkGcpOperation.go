package iprange

import (
	"context"
	"errors"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func checkGcpOperation(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	opName := ipRange.Status.OpIdentifier
	logger.WithValues("ipRange :", ipRange.Name).Info("Checking Operation")

	//If no OpIdentifier, then continue to next action.
	if opName == "" {
		return nil, nil
	}

	//Check SyncPsa Operation..
	if ipRange.Status.State == client.SyncPsaConnection ||
		ipRange.Status.State == client.DeletePsaConnection {
		op, err := state.serviceNetworkingClient.GetOperation(ctx, opName)
		if err != nil {
			state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, err)
			return composed.LogErrorAndReturn(err, "Error getting operation from GCP", composed.StopWithRequeue, nil)
		}

		//Operation not completed yet.. requeue again.
		if op != nil && !op.Done {
			return composed.StopWithRequeueDelay(client.GcpRetryWaitTime), nil
		}

		//If not able to find the operation or it is completed, reset OpIdentifier.
		if op == nil || op.Done {
			ipRange.Status.OpIdentifier = ""
		}

		//If the operation failed, update the error status on the object.
		if op != nil && op.Error != nil {
			state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, errors.New(op.Error.Message))
		}
	} else if ipRange.Status.State == client.SyncAddress ||
		ipRange.Status.State == client.DeleteAddress {
		project := state.Scope().Spec.Scope.Gcp.Project
		op, err := state.computeClient.GetGlobalOperation(ctx, project, opName)
		if err != nil {
			state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, err)
			return composed.LogErrorAndReturn(err, "Error getting operation from GCP", composed.StopWithRequeue, nil)
		}

		//Operation not completed yet.. requeue again.
		if op != nil && op.Status != "DONE" {
			return composed.StopWithRequeueDelay(client.GcpRetryWaitTime), nil
		}

		//If not able to find the operation or it is completed, reset OpIdentifier.
		if op == nil || op.Status == "DONE" {
			ipRange.Status.OpIdentifier = ""
		}

		//If the operation failed, update the error status on the object.
		if op != nil && op.Error != nil {
			state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, errors.New(op.StatusMessage))
		}

	}

	return nil, nil
}
