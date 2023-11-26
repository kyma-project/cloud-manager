package genericActions

import (
	"context"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
)

func SaveObj(ctx context.Context, state composed.State) error {
	err := state.UpdateObj(ctx)
	return state.RequeueIfError(err, "error saving object")
}
