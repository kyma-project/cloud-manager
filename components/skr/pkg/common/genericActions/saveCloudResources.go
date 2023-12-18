package genericActions

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/skr/api/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources/components/skr/pkg/common/composedAction"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SaveServedCloudResourcesAggregations(ctx context.Context, state composed.State) error {
	cr := state.(StateWithCloudResources).ServedCloudResources()
	obj := &cloudresourcesv1beta1.CloudResources{
		TypeMeta: cr.TypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: cloudresourcesv1beta1.CloudResourcesSpec{
			Aggregations: cr.Spec.Aggregations,
		},
	}
	err := state.Client().Patch(ctx, obj, client.Apply, &client.PatchOptions{
		Force:        pointer.Bool(true),
		FieldManager: "cloud-resources-manager",
	})

	return state.RequeueIfError(err, "error saving CloudResources")
}
