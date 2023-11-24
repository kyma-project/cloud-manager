package reconcile

import (
	"context"
	"github.com/kyma-project/cloud-resources-manager/pkg/common/aggregation"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler interface {
	Run(ctx context.Context, req ctrl.Request, obj client.Object) (*ctrl.Result, error)
}

func NewReconciler(client client.Client, eventRecorder record.EventRecorder) Reconciler {
	return &reconciler{
		client:        client,
		eventRecorder: eventRecorder,
		action: composed.ComposeActions(
			"cloud-resources-manager",
			aggregation.LoadObj,
			aggregation.LoadCloudResources,
			aggregation.Aggregate,
			aggregation.SaveCloudResourcesAggregations,
		),
	}
}

type reconciler struct {
	client        client.Client
	eventRecorder record.EventRecorder
	action        composed.Action
}

// Run runs the reconciliation actions.
// For ctx and req pass values from the kubebuilder Reconcile arguments.
// For obj argument pass empty instance of the object type you're reconciling
// in the controller, it will be loaded by executed actions and aggregated
// to the served CloudResources instance which then will be saved
func (r *reconciler) Run(
	ctx context.Context,
	req ctrl.Request,
	obj client.Object,
) (*ctrl.Result, error) {
	state := &aggregation.ReconcilingState{
		NamespacedName: req.NamespacedName,
		Obj:            obj,
		BaseState: composed.BaseState{
			Client:        r.client,
			EventRecorder: r.eventRecorder,
		},
	}

	return r.action(ctx, state)
}
