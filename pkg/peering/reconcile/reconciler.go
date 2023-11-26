package reconcile

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	"github.com/kyma-project/cloud-resources-manager/pkg/common/genericActions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

			composed.BuildBranchingAction(
				"ifBeingDeleted",
				genericActions.IsBeingDeleted,
				composed.ComposeActions(
					"whenBeingDeleted",
					composed.BuildBranchingAction(
						"ifHasDeleteOutcome",
						genericActions.HasOutcome(cloudresourcesv1beta1.OutcomeTypeDeleted),
						composed.ComposeActions(
							"whenBeingDeleted",
							genericActions.UpdateCondition(
								cloudresourcesv1beta1.ProcessingState,
								cloudresourcesv1beta1.ConditionTypeDeleted,
								cloudresourcesv1beta1.ConditionReasonProcessing,
								metav1.ConditionTrue,
								"Peering was deleted",
							),
							genericActions.SaveStatus,
							genericActions.FinalizerRemove(cloudresourcesv1beta1.Finalizer),
						),
						nil,
					),
					genericActions.Stop, // !!!important
				),
				nil,
			), // ifBeingDeleted

			composed.BuildBranchingAction(
				"ifHasNoFinalizer",
				composed.Not(genericActions.HasFinalizer),
				composed.ComposeActions(
					"whenNoFinalizer",
					genericActions.UpdateCondition(
						cloudresourcesv1beta1.ProcessingState,
						cloudresourcesv1beta1.ConditionTypeProcessing,
						cloudresourcesv1beta1.ConditionReasonProcessing,
						metav1.ConditionTrue,
						"Provisioning the resource",
					),
					genericActions.SaveStatus,
					genericActions.FinalizerAdd(cloudresourcesv1beta1.Finalizer),
					genericActions.SaveObj,
					genericActions.StopWithRequeue, // !!!important
				),
				nil,
			), // ifHasNoFinalizer

			genericActions.LoadCloudResources,
			genericActions.EnsureServedCloudResources,
			genericActions.Aggregate,
			genericActions.SaveServedCloudResourcesAggregations,

			genericActions.OutcomeErrorToCondition(),
			genericActions.OutcomeCreatedToCondition(),

			genericActions.SaveStatus,
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
