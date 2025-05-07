package azurerwxvolumebackup

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type protectedFileShare struct {
	Id         *string
	Name       *string
	Properties *armrecoveryservicesbackup.AzureFileshareProtectedItem
}

func matchProtectedItems(protectedItems []*armrecoveryservicesbackup.ProtectedItemResource, fileShareName string) []protectedFileShare {

	var matchingFileshares []protectedFileShare

	for _, item := range protectedItems {
		iProperties, ok := item.Properties.(*armrecoveryservicesbackup.AzureFileshareProtectedItem)
		if !ok {
			continue
		}

		if iProperties.FriendlyName != nil && *iProperties.FriendlyName == fileShareName {
			matchingFileshares = append(matchingFileshares, protectedFileShare{
				Id:         item.ID,
				Name:       item.Name,
				Properties: iProperties,
			})
		}

	}

	return matchingFileshares

}

func getProtectedResourceName(ctx context.Context, st composed.State) (error, context.Context) {

	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAzureRwxVolumeBackup()
	logger.WithValues("RwxBackup", backup.Name).Info("Getting ProtectedResourceName")

	vaultName := state.vaultName
	resourceGroupName := state.resourceGroupName
	fileShareName := state.fileShareName

	protectedItems, err := state.client.ListProtectedItems(ctx, vaultName, resourceGroupName)
	if err != nil {
		backup.Status.State = cloudresourcesv1beta1.AzureRwxBackupError
		return composed.PatchStatus(backup).
			FailedError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	protectedItemsMatching := matchProtectedItems(protectedItems, fileShareName)

	// assign and proceed
	if len(protectedItemsMatching) == 1 {
		state.protectedResourceName = *protectedItemsMatching[0].Name
		return nil, ctx
	}

	if len(protectedItemsMatching) > 1 {
		backup.Status.State = cloudresourcesv1beta1.AzureRwxBackupFailed
		return composed.PatchStatus(backup).
			SetExclusiveConditions(
				metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  "AzureError", // TODO: create constant
					Message: "More than 1 matching protectedItem",
				},
			).
			FailedError(composed.StopAndForget).
			Run(ctx, state)

	}

	// Case: 0 matching; means we need to go protect; go in without setting
	return nil, ctx

}
