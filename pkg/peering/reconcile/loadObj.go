package reconcile

import (
	"context"
	"fmt"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	ctrl "sigs.k8s.io/controller-runtime"
)

func loadObj(ctx context.Context, state1 composed.LoggableState) (*ctrl.Result, error) {
	state := state1.(*ReconcilingState)
	err := state.Client.Get(ctx, state.NamespacedName, state.Obj)
	if err != nil {
		return &ctrl.Result{Requeue: true}, fmt.Errorf("error getting object %s: %w", state.NamespacedName, err)
	}
	return nil, nil
}
