package genericActions

import (
	"context"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
)

func Stop(ctx context.Context, state composed.State) error {
	return state.Stop(nil)
}
