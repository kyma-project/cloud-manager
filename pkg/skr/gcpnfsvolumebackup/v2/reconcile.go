package v2

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// New creates the v2 reconciliation action
func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Using GcpNfsVolumeBackup v2 implementation")

		state, err := stateFactory.NewState(ctx, st)
		if err != nil {
			logger.Error(err, "Error creating v2 state")
			backup := st.Obj().(*cloudresourcesv1beta1.GcpNfsVolumeBackup)
			return composed.PatchStatus(backup).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonError,
					Message: fmt.Sprintf("Failed to initialize GCP client: %s", err.Error()),
				}).
				SuccessError(composed.StopAndForget).
				Run(ctx, st)
		}

		return composeActions()(ctx, state)
	}
}

func composeActions() composed.Action {
	return composed.ComposeActions(
		"gcpNfsVolumeBackupV2",
		loadScope,
		shortCircuitCompleted,
		markFailed,
		actions.AddCommonFinalizer(),

		loadNfsBackup,
		loadGcpNfsVolume,

		checkBackupOperation,

		composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
			composed.ComposeActions(
				"gcpNfsVolumeBackupV2-create",
				createNfsBackup,
				waitBackupReady,
				addLabelsToNfsBackup,
				updateStatus,
			),
			composed.ComposeActions(
				"gcpNfsVolumeBackupV2-delete",
				deleteNfsBackup,
				waitBackupDeleted,
				actions.RemoveCommonFinalizer(),
			),
		),
		composed.StopAndForgetAction,
	)
}
