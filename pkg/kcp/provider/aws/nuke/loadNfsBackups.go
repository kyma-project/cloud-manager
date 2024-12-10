package nuke

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadNfsBackups(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	backups, err := state.awsClient.ListRecoveryPointsForVault(ctx, state.GetAccountId(), state.GetVaultName())
	if err != nil {
		logger.Error(err, "Error listing Aws Recovery Points")

		state.ObjAsNuke().Status.State = string(cloudcontrolv1beta1.ErrorState)

		return composed.PatchStatus(state.ObjAsNuke()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  "ErrorListingAwsRecoveryPoints",
				Message: err.Error(),
			}).
			ErrorLogMessage("Error patching KCP Nuke status after list AwsNfsVolume Backups error").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			Run(ctx, state)
	}
	awsBackups := make([]nuketypes.ProviderResourceObject, len(backups))
	for i, backup := range backups {
		awsBackups[i] = AwsBackup{&backup}
	}

	state.ProviderResources = append(state.ProviderResources, &nuketypes.ProviderResourceKindState{
		Kind:     "AwsNfsVolumeBackup",
		Provider: cloudcontrolv1beta1.ProviderAws,
		Objects:  awsBackups,
	})
	return nil, ctx
}
