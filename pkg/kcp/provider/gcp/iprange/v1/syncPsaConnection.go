package v1

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
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
		operation, err = state.serviceNetworkingClient.PatchServiceConnection(ctx, project, vpc, state.ipRanges)
	case client.DELETE:
		operation, err = state.serviceNetworkingClient.DeleteServiceConnection(ctx, project, vpc)
	}

	if err != nil {
		return composed.UpdateStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg(fmt.Sprintf("Error creating/deleting Service Connection object in GCP :%s", err)).
			Run(ctx, state)
	}
	if operation != nil {
		ipRange.Status.OpIdentifier = operation.Name
		return composed.UpdateStatus(ipRange).
			SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpOperationWaitTime)).
			Run(ctx, state)
	}
	return composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpOperationWaitTime), nil
}
