package v2

import (
	"context"
	"fmt"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	v2client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v2/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// deleteInstance deletes a Filestore instance from GCP.
func deleteInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	instance := state.GetInstance()

	// Skip if instance doesn't exist
	if instance == nil {
		return nil, ctx
	}

	// Skip if already deleting
	if instance.State == filestorepb.Instance_DELETING {
		return nil, ctx
	}

	nfsInstance := state.ObjAsNfsInstance()
	logger.Info("Deleting GCP Filestore Instance")

	// Get GCP details
	project := state.GetGcpProjectId()
	location := state.GetGcpLocation()
	name := v2client.GetFilestoreInstanceId(nfsInstance.Name)

	// Delete instance
	operationName, err := state.GetFilestoreClient().DeleteInstance(ctx, project, location, name)
	if err != nil {
		logger.Error(err, "Error deleting Filestore Instance in GCP")
		return composed.UpdateStatus(nfsInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: fmt.Sprintf("Error deleting Filestore Instance: %s", err.Error()),
			}).
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg("Error deleting Filestore Instance in GCP").
			Run(ctx, state)
	}

	// Store operation for polling
	if operationName != "" {
		nfsInstance.Status.OpIdentifier = operationName

		return composed.UpdateStatus(nfsInstance).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	return nil, ctx
}
