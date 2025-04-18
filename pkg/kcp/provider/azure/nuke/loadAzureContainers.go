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

func loadAzureContainers(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("LoadAzureContainers")
	state.protectionContainers = make(map[string]*armrecoveryservicesbackup.AzureStorageContainer)
	for _, vault := range state.recoveryVaults {
		items, err := state.azureClient.ListStorageContainers(ctx, vault)
		if err != nil {
			logger.Error(err, "Error listing Azure Storage Containers")

			state.ObjAsNuke().Status.State = string(cloudcontrolv1beta1.StateError)

			return composed.PatchStatus(state.ObjAsNuke()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  "ErrorListingAzureContainers",
					Message: err.Error(),
				}).
				ErrorLogMessage("Error patching KCP Nuke status after list AzureContainers error").
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
				Run(ctx, state)
		}

		maps.Insert(state.protectionContainers, maps.All(items))
	}

	azureContainers := make([]nuketypes.ProviderResourceObject, 0)
	for id, container := range state.protectionContainers {
		azureContainers = append(azureContainers, azureContainer{container, id})
	}

	state.ProviderResources = append(state.ProviderResources, &nuketypes.ProviderResourceKindState{
		Kind:     azurenukeclient.AzureStorageContainer,
		Provider: cloudcontrolv1beta1.ProviderAzure,
		Objects:  azureContainers,
	})
	return nil, ctx
}
