package azurerwxpv

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func loadAzureProtectedItem(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("Load Azure ProtectedItems")

	//If not able to load the fileShare, just continue.
	if state.fileShare == nil {
		return nil, ctx
	}

	for _, vault := range state.recoveryVaults {
		items, err := state.client.ListFileShareProtectedItems(ctx, vault)
		if err != nil {
			return composed.LogErrorAndReturn(err, "error loading AzureFileShareProtection", err, ctx)
		}

		for id, item := range items {
			if *item.FriendlyName == state.fileShareName {
				state.protectedId = id
				state.protectedItem = item

				logger.Info(fmt.Sprintf("Loaded protected item : %v", state.protectedId))
				return nil, ctx
			}
		}
	}

	return nil, ctx
}
