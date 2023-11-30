package reconcile

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/api/cloud-resources/v1beta1"
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
			"cloud-resources-manager",
			genericActions.LoadObj,
			genericActions.LoadCloudResources,
			handleServed,
		),
	}
}

type Reconciler struct {
	client        client.Client
	eventRecorder record.EventRecorder
	action        composed.Action
}

func (r *Reconciler) Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	state := genericActions.NewState(
		composed.NewState(r.client, r.eventRecorder, req.NamespacedName, &cloudresourcesv1beta1.CloudResources{}),
	)

	err := r.action(ctx, state)

	return state.Result(), err
}
