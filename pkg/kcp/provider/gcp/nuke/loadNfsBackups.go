package nuke

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadNfsBackups(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	backups, err := state.fileBackupClient.ListFilesBackups(ctx, state.State.Scope().Spec.Scope.Gcp.Project, client.GetSkrBackupsFilter(state.State.Scope().Name))
	if err != nil {
		logger.Error(err, "Error listing Gcp Filestore Backups")

		state.ObjAsNuke().Status.State = string(cloudcontrolv1beta1.StateError)

		return composed.PatchStatus(state.ObjAsNuke()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  "ErrorListingGcpFilestoreBackups",
				Message: err.Error(),
			}).
			ErrorLogMessage("Error patching KCP Nuke status after list GCP Filestore Backups error").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			Run(ctx, state)
	}
	gcpBackups := make([]nuketypes.ProviderResourceObject, len(backups))
	for i, backup := range backups {
		gcpBackups[i] = GcpBackup{backup}
	}

	state.ProviderResources = append(state.ProviderResources, &nuketypes.ProviderResourceKindState{
		Kind:     "FilestoreBackup",
		Provider: cloudcontrolv1beta1.ProviderGCP,
		Objects:  gcpBackups,
	})
	return nil, ctx
}
