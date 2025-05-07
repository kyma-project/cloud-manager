package nuke

import (
	"context"
	"maps"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	azurenukeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/nuke/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadAzureBackups(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("LoadAzureProtectedItems")
	state.protectedItems = make(map[string]*armrecoveryservicesbackup.AzureFileshareProtectedItem)
	for _, vault := range state.recoveryVaults {
		items, err := state.azureClient.ListFileShareProtectedItems(ctx, vault)
		if err != nil {
			logger.Error(err, "Error listing AzureFileShareProtection")

			state.ObjAsNuke().Status.State = string(cloudcontrolv1beta1.StateError)

			return composed.PatchStatus(state.ObjAsNuke()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  "Error listing AzureFileShareProtection",
					Message: err.Error(),
				}).
				ErrorLogMessage("Error patching KCP Nuke status after list AzureBackups error").
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
				Run(ctx, state)
		}

		maps.Insert(state.protectedItems, maps.All(items))
	}

	azureBackups := make([]nuketypes.ProviderResourceObject, 0)
	for id, fileShare := range state.protectedItems {
		azureBackups = append(azureBackups, azureFileshare{fileShare, id})
	}

	state.ProviderResources = append(state.ProviderResources, &nuketypes.ProviderResourceKindState{
		Kind:     azurenukeclient.AzureFileShareProtection,
		Provider: cloudcontrolv1beta1.ProviderAzure,
		Objects:  azureBackups,
	})
	return nil, ctx
}
