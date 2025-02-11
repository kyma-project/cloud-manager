package v2

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"google.golang.org/api/servicenetworking/v1"
)

func syncPsaConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	logger.WithValues("ipRange", ipRange.Name).Info("Saving GCP PSA Connection")

	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork

	var operation *servicenetworking.Operation
	var err error

	switch state.connectionOp {
	case client.ADD:
		operation, err = state.serviceNetworkingClient.CreateServiceConnection(ctx, project, vpc, state.ipRanges)
	case client.MODIFY:
		if len(state.ipRanges) > 0 {
			operation, err = state.serviceNetworkingClient.PatchServiceConnection(ctx, project, vpc, state.ipRanges)
		} else {
			operation, err = state.serviceNetworkingClient.DeleteServiceConnection(ctx, project, vpc)
		}
	case client.DELETE:
		operation, err = state.serviceNetworkingClient.DeleteServiceConnection(ctx, project, vpc)
	}

	if err != nil {
		logger.Error(err, "Error creating/deleting/patching Service Connection object in GCP")
		return composed.PatchStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime)).
			Run(ctx, state)
	}
	if operation != nil {
		ipRange.Status.OpIdentifier = operation.Name
		return composed.PatchStatus(ipRange).
			SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpOperationWaitTime)).
			Run(ctx, state)
	}
	return composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpOperationWaitTime), nil
}
