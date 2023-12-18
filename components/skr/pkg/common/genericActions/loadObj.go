package genericActions

import (
	"context"
	"fmt"
	composed "github.com/kyma-project/cloud-resources/components/skr/pkg/common/composedAction"
)

func LoadObj(ctx context.Context, state composed.State) error {
	err := state.LoadObj(ctx)
	return state.RequeueIfError(err, fmt.Sprintf("error getting object %s", state.Name()))
}
