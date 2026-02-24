package v2

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func checkBackupOperation(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	backup := state.ObjAsGcpNfsVolumeBackup()
	opName := backup.Status.OpIdentifier
	logger.Info("Checking GCP Backup Operation Status")

	if opName == "" {
		return nil, nil
	}

	op, err := state.fileBackupClient.GetBackupLROperation(ctx, opName)
	if err != nil {
		// If the operation is not found, reset the OpIdentifier and let retry or updateStatus to update the status.
		if gcpmeta.IsNotFound(err) {
			backup.Status.OpIdentifier = ""
		}
		backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupError
		logger.Error(err, "Error getting Filestore backup Operation from GCP.")
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	if !op.Done {
		return composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), nil
	}

	// Operation is done, reset OpIdentifier
	backup.Status.OpIdentifier = ""

	if op.GetError() != nil {
		opErr := status.FromProto(op.GetError())
		backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupError
		logger.Error(opErr.Err(), "GCP Filestore backup operation failed", "code", opErr.Code(), "message", opErr.Message())
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonGcpError,
				Message: "Operation failed",
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	if backup.Status.State == cloudresourcesv1beta1.GcpNfsBackupDeleting {
		backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupDeleted
		state.fileBackup = nil
		return composed.PatchStatus(backup).
			SetExclusiveConditions().
			SuccessErrorNil().
			Run(ctx, state)
	}

	backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupReady
	return composed.PatchStatus(backup).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonReady,
			Message: fmt.Sprintf("Backup operation finished successfully: %s", opName),
		}).
		OnUpdateSuccess(func(ctx context.Context) (error, context.Context) {
			return nil, nil
		}). // proceed in case deletion is in progress
		SuccessLogMsg("GcpNfsVolumeBackup status got updated with Ready condition and Done state.").
		Run(ctx, state)
}
