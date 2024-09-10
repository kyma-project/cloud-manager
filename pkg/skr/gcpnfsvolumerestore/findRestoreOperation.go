package gcpnfsvolumerestore

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func findRestoreOperation(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	//If deleting, continue with next steps.
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	restore := state.ObjAsGcpNfsVolumeRestore()
	opName := restore.Status.OpIdentifier
	logger.WithValues("nfsRestoreSource:", restore.Spec.Source.Backup.ToNamespacedName(state.Obj().GetNamespace()),
		"destination:", restore.Spec.Destination.Volume.ToNamespacedName(state.Obj().GetNamespace())).Info("Finding GCP Restore Operations")

	//If no OpIdentifier, then continue to next action.
	if opName != "" {
		return nil, nil
	}

	project := state.Scope.Spec.Scope.Gcp.Project
	dstLocation := state.GcpNfsVolume.Status.Location
	nfsInstanceName := fmt.Sprintf("cm-%.60s", state.GcpNfsVolume.Status.Id)
	op, err := state.fileRestoreClient.FindRestoreOperation(ctx, project, dstLocation, nfsInstanceName)
	if err != nil {
		if meta.IsNotFound(err) {
			return nil, nil
		}
		restore.Status.State = v1beta1.JobStateError
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ConditionReasonNfsRestoreFailed,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T100ms())).
			SuccessLogMsg("Error listing operations from GCP.").
			Run(ctx, state)
	}

	if op != nil {
		restore.Status.OpIdentifier = op.Name
		return composed.PatchStatus(restore).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg(fmt.Sprintf("Updated the opIdentifier with %s", op.Name)).
			Run(ctx, state)
	}
	return nil, nil
}
