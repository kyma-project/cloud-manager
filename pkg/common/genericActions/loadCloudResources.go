package genericActions

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/api/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
)

func LoadCloudResources(ctx context.Context, state composed.State) error {
	list := &cloudresourcesv1beta1.CloudResourcesList{}
	state.(StateWithCloudResources).SetCloudResourcesList(list)
	err := state.Client().List(ctx, list)
	if err != nil {
		return fmt.Errorf("error listing CloudResources: %w", err)
	}

	for _, cr := range list.Items {
		if cr.Status.Served == cloudresourcesv1beta1.ServedTrue {
			ensureCloudResourcesDefaults(&cr)
			state.(StateWithCloudResources).SetServedCloudResources(&cr)
			return nil
		}
	}

	return nil
}

func ensureCloudResourcesDefaults(cr *cloudresourcesv1beta1.CloudResources) {
	if cr.Spec.Aggregations == nil {
		cr.Spec.Aggregations = &cloudresourcesv1beta1.CloudResourcesAggregation{}
	}
}
