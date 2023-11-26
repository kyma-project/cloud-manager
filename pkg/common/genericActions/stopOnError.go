package genericActions

import (
	"context"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
)

func StopOnError(a composed.Action, msg ...string) composed.Action {
	return func(ctx context.Context, state composed.State) error {
		err := a(ctx, state)
		if err != nil {
			return state.Stop(err, msg...)
		}
		return nil
	}
}
