package subscription

import (
	"context"
	"fmt"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// resourcesLoad loads all resources having label cloudcontrolv1beta1.SubscriptionLabel matching
// this subscription. This is probably a tmp workaround or a final fallback. Known resources like VpcNetwork
// and Runtime rather can be loaded and their fields inspected
func resourcesLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	state.resources = map[schema.GroupVersionKind][]metav1.PartialObjectMetadata{}

	for gvk := range state.Cluster().Scheme().AllKnownTypes() {
		if gvk.Group != cloudcontrolv1beta1.GroupVersion.Group {
			continue
		}
		if gvk.Kind == "Subscription" {
			continue
		}
		if strings.HasSuffix(gvk.Kind, "List") {
			continue
		}
		list := &metav1.PartialObjectMetadataList{
			TypeMeta: metav1.TypeMeta{
				Kind:       gvk.Kind,
				APIVersion: gvk.GroupVersion().String(),
			},
		}

		err := state.Cluster().K8sClient().List(
			ctx,
			list,
			client.MatchingLabels{
				cloudcontrolv1beta1.SubscriptionLabel: state.Name().Name,
			},
			client.InNamespace(state.Name().Namespace),
		)
		if meta.IsNoMatchError(err) {
			// this CRD is not installed
			continue
		}
		if err != nil {
			logger.
				WithValues(
					"errorType", fmt.Sprintf("%T", err),
					"gvk", gvk.String(),
				).
				Error(err, "Error listing GVK for Subscription resource usage check on delete")
			continue
		}

		if len(list.Items) == 0 {
			continue
		}
		state.resources[gvk] = append(state.resources[gvk], list.Items...)
	}

	return nil, ctx
}
