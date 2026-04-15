package v2

import (
	"context"
	"fmt"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func checkRestoreOperation(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	restore := state.ObjAsGcpNfsVolumeRestore()
	opName := restore.Status.OpIdentifier
	logger.WithValues("nfsRestoreSource", restore.Spec.Source.Backup.ToNamespacedName(state.Obj().GetNamespace()),
		"destination", restore.Spec.Destination.Volume.ToNamespacedName(state.Obj().GetNamespace())).Info("Checking GCP Restore Operation Status")

	if opName == "" {
		return nil, nil
	}

	op, err := state.fileRestoreClient.GetFilestoreOperation(ctx, &longrunningpb.GetOperationRequest{
		Name: opName,
	})
	if err != nil {
		if gcpmeta.IsNotFound(err) {
			restore.Status.OpIdentifier = ""
		}
		restore.Status.State = v1beta1.JobStateError
		logger.Error(err, "Error getting Filestore restore Operation from GCP.")
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ConditionReasonNfsRestoreFailed,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	if !op.Done {
		return composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), nil
	}

	restore.Status.OpIdentifier = ""

	//If the operation failed, update the error status on the object.
	if op.GetError() != nil {
		opErr := status.FromProto(op.GetError())
		restore.Status.State = v1beta1.JobStateFailed
		logger.Error(opErr.Err(), "GCP Filestore restore operation failed", "code", opErr.Code(), "message", opErr.Message())
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonGcpError,
				Message: op.GetError().GetMessage(),
			}).
			OnUpdateSuccess(func(ctx context.Context) (error, context.Context) {
				return nil, nil
			}). //proceed to release the lease
			SuccessLogMsg(fmt.Sprintf("Filestore Operation error : %s", op.GetError().GetMessage())).
			Run(ctx, state)
	}

	//Done Successfully
	restore.Status.State = v1beta1.JobStateDone
	return composed.PatchStatus(restore).
		SetExclusiveConditions(metav1.Condition{
			Type:    v1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  v1beta1.ConditionReasonReady,
			Message: fmt.Sprintf("Restore operation finished successfully: %s", opName),
		}).
		OnUpdateSuccess(func(ctx context.Context) (error, context.Context) {
			return nil, nil
		}). //proceed to release the lease
		SuccessLogMsg("GcpNfsVolumeRestore status got updated with Ready condition and Done state.").
		Run(ctx, state)
}
