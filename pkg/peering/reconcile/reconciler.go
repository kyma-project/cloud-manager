package reconcile

import (
	"context"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	"github.com/kyma-project/cloud-resources-manager/pkg/common/genericActions"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewReconciler(client client.Client, eventRecorder record.EventRecorder) *Reconciler {
	return &Reconciler{
		client:        client,
		eventRecorder: eventRecorder,
		action: composed.ComposeActions(
			"peeringLoop",

			genericActions.LoadObj,
			whenBeingDeleted,
			whenNoFinalizer,
			genericActions.LoadCloudResources,
			genericActions.EnsureServedCloudResources,
			genericActions.Aggregate,
			genericActions.SaveServedCloudResourcesAggregations,
			whenOutcomeError,
			whenOutcomeCreated,
		),
	}
}

type Reconciler struct {
	client        client.Client
	eventRecorder record.EventRecorder
	action        composed.Action
}

// Run runs the reconciliation actions.
// For ctx and req pass values from the kubebuilder's Reconcile() arguments.
// For obj argument pass empty instance of the object type you're reconciling
// in the controller, it will be loaded by executed actions and aggregated
// to the served CloudResources instance which then will be saved
func (r *Reconciler) Run(
	ctx context.Context,
	req ctrl.Request,
	obj client.Object,
) (ctrl.Result, error) {
	state := genericActions.NewState(
		composed.NewState(
			r.client,
			r.eventRecorder,
			req.NamespacedName,
			obj,
		),
	)

	err := r.action(ctx, state)

	return state.Result(), err
}
