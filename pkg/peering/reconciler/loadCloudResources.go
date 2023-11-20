package reconciler

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	ctrl "sigs.k8s.io/controller-runtime"
)

func loadCloudResources(ctx context.Context, state1 composed.LoggableState) (*ctrl.Result, error) {
	state := state1.(*ReconcilingState)
	list := &v1beta1.CloudResourcesList{}
	err := state.List(ctx, list)
	if err != nil {
		return &ctrl.Result{Requeue: true},
			fmt.Errorf("error listing CloudResources: %w", err)
	}

	for _, cr := range list.Items {
		if cr.Status.Served == v1beta1.ServedTrue {
			state.CloudResources = &cr
			return nil, nil
		}
	}

	return &ctrl.Result{Requeue: true},
		errors.New("no served CloudResources found")
}
