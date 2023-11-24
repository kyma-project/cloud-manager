package aggregation

import (
	"context"
	"fmt"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SaveCloudResourcesAggregations(ctx context.Context, state1 composed.LoggableState) (*ctrl.Result, error) {
	state := state1.(*ReconcilingState)
	err := state.Client.Patch(ctx, state.CloudResources, client.Apply, &client.PatchOptions{
		Force:        pointer.Bool(true),
		FieldManager: "cloud-resources-manager",
	})
	if err != nil {
		return &ctrl.Result{Requeue: true}, fmt.Errorf("error saving CloudResources: %w", err)
	}
	return nil, nil
}
