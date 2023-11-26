package genericActions

import (
	"context"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
)

func SaveStatus(ctx context.Context, state composed.State) error {
	err := state.UpdateObjStatus(ctx)
	return state.RequeueIfError(err, "error saving status")
}
