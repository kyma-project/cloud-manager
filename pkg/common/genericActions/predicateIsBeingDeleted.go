package genericActions

import (
	"context"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
)

func IsBeingDeleted(ctx context.Context, state composed.State) bool {
	return state.Obj().GetDeletionTimestamp().IsZero()
}
