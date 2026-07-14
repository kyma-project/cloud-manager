package security

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"k8s.io/utils/ptr"
)

func storageAccountLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	list, err := state.azureClient.ListStorageAccounts(ctx)
	if err != nil {
		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error loading network watcher: %s", err.Error()), state.ObjAsRuntime().Generation)
		return azuremeta.LogErrorAndReturn(err, "Error listing storage accounts", ctx)
	}

	runtimeId := state.ObjAsRuntime().Name
	for _, acc := range list {
		if acc.Tags != nil {
			if tagVal := ptr.Deref(acc.Tags[tagKymaRuntimeId], ""); tagVal == runtimeId {
				if tagVal := ptr.Deref(acc.Tags[tagKymaPurpose], ""); tagVal == tagValuePurposeNetworkFlowLogs {
					state.storageAccount = acc
					break
				}
			}
		}
	}

	return nil, ctx
}
