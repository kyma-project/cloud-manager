package genericActions

import (
	"context"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
)

func StopWithRequeue(_ context.Context, state composed.State) error {
	return state.StopWithRequeue()
}
