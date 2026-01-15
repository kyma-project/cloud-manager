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
)

// updateInstance updates an existing Filestore instance in GCP.
func updateInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	nfsInstance := state.ObjAsNfsInstance()
	instance := state.GetInstance()

	// Skip if instance doesn't exist
	if instance == nil {
		return nil, ctx
	}

	// Skip if instance is not ready
	if instance.State != filestorepb.Instance_READY {
		return nil, ctx
	}

	// Skip if no updates needed
	if state.DoesFilestoreMatch() {
		return nil, ctx
	}

	logger.Info("Updating GCP Filestore Instance")

	// Get GCP details
	project := state.GetGcpProjectId()
	location := state.GetGcpLocation()
	name := v2client.GetFilestoreInstanceId(nfsInstance.Name)

	// Build instance and calculate update mask
	gcpInstance := state.ToGcpInstance()
	var updateMask []string

	// Check if capacity changed
	if len(instance.FileShares) > 0 {
		desiredCapacity := int64(nfsInstance.Spec.Instance.Gcp.CapacityGb)
		actualCapacity := instance.FileShares[0].CapacityGb

		if desiredCapacity != actualCapacity {
			updateMask = append(updateMask, "file_shares")
		}
	}

	// If no changes, skip update
	if len(updateMask) == 0 {
		return nil, ctx
	}

	// Update instance
	operationName, err := state.GetFilestoreClient().UpdateInstance(ctx, project, location, name, gcpInstance, updateMask)
	if err != nil {
		logger.Error(err, "Error updating Filestore Instance in GCP")
		return composed.UpdateStatus(nfsInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: fmt.Sprintf("Error updating Filestore Instance: %s", err.Error()),
			}).
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg("Error updating Filestore Instance in GCP").
			Run(ctx, state)
	}

	// Store operation for polling
	if operationName != "" {
		nfsInstance.Status.OpIdentifier = operationName

		return composed.UpdateStatus(nfsInstance).
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpOperationWaitTime)).
			Run(ctx, state)
	}

	return nil, ctx
}
