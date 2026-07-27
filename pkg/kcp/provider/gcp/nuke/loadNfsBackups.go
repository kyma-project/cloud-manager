package nuke

import (
	"context"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsbackupclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadNfsBackups(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	project := state.Scope().Spec.Scope.Gcp.Project
	// Location "-" lists backups across all locations of the project, matching the prior
	// V1 all-locations listing; the SKR filter restricts the result to this Scope's backups.
	parent := gcpnfsbackupclientv2.GetFilestoreParentPath(project, "-")
	filter := client.GetSkrBackupsFilter(state.State.Scope().Name)

	iter := state.fileBackupClient.ListFilestoreBackups(ctx, &filestorepb.ListBackupsRequest{
		Parent: parent,
		Filter: filter,
	})

	var backups []*filestorepb.Backup
	for backup, err := range iter.All() {
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
		backups = append(backups, backup)
	}

	gcpBackups := make([]nuketypes.ProviderResourceObject, len(backups))
	for i, backup := range backups {
		gcpBackups[i] = GcpBackup{backup}
	}

	state.ProviderResources = append(state.ProviderResources, &nuketypes.ProviderResourceKindState{
		Kind:     kindFilestoreBackup,
		Provider: cloudcontrolv1beta1.ProviderGCP,
		Objects:  gcpBackups,
	})
	return nil, ctx
}
