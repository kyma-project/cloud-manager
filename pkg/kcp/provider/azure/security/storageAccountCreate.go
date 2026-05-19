package security

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

const maxStorageAccountCreateAttempts = 10

func storageAccountCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.storageAccount != nil {
		return nil, ctx
	}
	if state.resourceGroupData == nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("resourceGroupData must exist before creating storage account"),
			"Cannot create storage account",
			composed.StopWithRequeue, ctx)
	}

	params := armstorage.AccountCreateParameters{
		Location: new(state.location()),
		Kind:     new(armstorage.KindStorageV2),
		SKU: &armstorage.SKU{
			Name: new(armstorage.SKUNameStandardLRS),
		},
		Tags: map[string]*string{
			tagKymaRuntimeId: new(state.ObjAsRuntime().Name),
			tagKymaShootName: new(state.shootName()),
		},
		Properties: &armstorage.AccountPropertiesCreateParameters{
			AllowBlobPublicAccess:  new(false),
			AllowSharedKeyAccess:   new(false),
			PublicNetworkAccess:    new(armstorage.PublicNetworkAccessDisabled),
			EnableHTTPSTrafficOnly: new(true),
			MinimumTLSVersion:      new(armstorage.MinimumTLSVersionTLS12),
			Encryption: &armstorage.Encryption{
				Services: &armstorage.EncryptionServices{
					File: &armstorage.EncryptionService{
						Enabled: new(true),
						KeyType: new(armstorage.KeyTypeAccount),
					},
					Blob: &armstorage.EncryptionService{
						Enabled: new(true),
						KeyType: new(armstorage.KeyTypeAccount),
					},
				},
				KeySource: new(armstorage.KeySourceMicrosoftStorage),
			},
		},
	}

	composed.LoggerFromCtx(ctx).Info("Creating storage account")

	for i := 0; i < maxStorageAccountCreateAttempts; i++ {
		accountName := state.storageAccountNameAttempt(i)

		resp, err := azureclient.PollUntilDone(state.azureClient.CreateStorageAccount(ctx,
			state.resourceGroupDataName(),
			accountName,
			params,
			nil))(ctx, nil)
		if err == nil {
			state.storageAccount = &resp.Account
			break
		}

		if azuremeta.IsStorageAccountNameConflict(err) {
			continue
		}

		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error creating storage account: %s", err.Error()), state.ObjAsRuntime().Generation)
		return azuremeta.LogErrorAndReturn(err, "Error creating storage account", ctx)
	}

	return nil, ctx
}
