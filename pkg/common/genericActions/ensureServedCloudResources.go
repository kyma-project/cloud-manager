package genericActions

import (
	"context"
	"errors"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
)

func EnsureServedCloudResources(ctx context.Context, state composed.State) error {
	if state.(StateWithCloudResources).ServedCloudResources() == nil {
		return state.RequeueIfError(errors.New("no served CloudResources found"))
	}
	return nil
}
