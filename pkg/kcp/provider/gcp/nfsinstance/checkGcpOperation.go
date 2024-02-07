package nfsinstance

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

	nfsInstance := state.ObjAsNfsInstance()
	opName := nfsInstance.Status.OpIdentifier
	logger.WithValues("NfsInstance :", nfsInstance.Name).Info("Checking GCP Operation Status")

	//If no OpIdentifier, then continue to next action.
	if opName == "" {
		return nil, nil
	}

	project := state.Scope().Spec.Scope.Gcp.Project
	op, err := state.filestoreClient.GetFilestoreOperation(ctx, project, opName)
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
		nfsInstance.Status.OpIdentifier = ""
	}

	//If the operation failed, update the error status on the object.
	if op != nil && op.Error != nil {
		state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, errors.New(op.Error.Message))
	}

	return nil, nil
}
