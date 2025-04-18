package nuke

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	azurenukeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/nuke/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadAzureRecoveryVaults(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("Load AzureRecoveryVaults")
	vaults, err := state.azureClient.ListRwxVolumeBackupVaults(ctx)
	if err != nil {

		state.ObjAsNuke().Status.State = string(cloudcontrolv1beta1.StateError)
		return composed.PatchStatus(state.ObjAsNuke()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  "Error listing Azure Recovery Vaults",
				Message: err.Error(),
			}).
			ErrorLogMessage("Error patching KCP Nuke status after list AzureBackups error").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			Run(ctx, state)
	}

	//Store it in state
	state.recoveryVaults = vaults

	azureVaults := make([]nuketypes.ProviderResourceObject, len(state.recoveryVaults))
	for i, vault := range state.recoveryVaults {
		azureVaults[i] = azureVault{vault}
	}

	state.ProviderResources = append(state.ProviderResources, &nuketypes.ProviderResourceKindState{
		Kind:     azurenukeclient.AzureRecoveryVault,
		Provider: cloudcontrolv1beta1.ProviderAzure,
		Objects:  azureVaults,
	})
	return nil, ctx
}
