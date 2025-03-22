package nuke

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadAzureBackups(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	for _, vault := range state.recoveryVaults {
		items, err := state.azureClient.ListFileShareProtectedItems(ctx, vault)
		if err != nil {
			logger.Error(err, "Error listing Azure Recovery Points")

			state.ObjAsNuke().Status.State = string(cloudcontrolv1beta1.StateError)

			return composed.PatchStatus(state.ObjAsNuke()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  "ErrorListingAzureRecoveryPoints",
					Message: err.Error(),
				}).
				ErrorLogMessage("Error patching KCP Nuke status after list AzureBackups error").
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
				Run(ctx, state)
		}

		state.protectedItems = append(state.protectedItems, items...)
	}

	azureBackups := make([]nuketypes.ProviderResourceObject, len(state.protectedItems))
	for i, backup := range state.protectedItems {
		azureBackups[i] = azureProtectedItem{backup}
	}

	state.ProviderResources = append(state.ProviderResources, &nuketypes.ProviderResourceKindState{
		Kind:     "AzureRwxVolumeBackup",
		Provider: cloudcontrolv1beta1.ProviderAzure,
		Objects:  azureBackups,
	})
	return nil, ctx
}
