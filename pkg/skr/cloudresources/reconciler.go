package cloudresources

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewReconcilerFactory() skrruntime.ReconcilerFactory {
	return &reconcilerFactory{}
}

type reconcilerFactory struct {
}

func (f *reconcilerFactory) New(args skrruntime.ReconcilerArguments) reconcile.Reconciler {
	return &reconciler{
		factory: newStateFactory(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(args.SkrCluster)),
			args.KymaRef,
			composed.NewStateClusterFromCluster(args.KcpCluster),
			args.Provider,
		),
	}
}

type reconciler struct {
	factory *stateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	state := r.factory.NewState(request)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("cloudresources", util.RequestObjToString(request)).
		WithNoLog().
		Handle(action(ctx, state))
}

// When module is deleted from the SKR Kyma spec
// Then module CR will get deletionTimestamp

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"cloudResources-main",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.CloudResources{}),
		composed.LoadObj,

		composed.BuildBranchingAction(
			"cloudResources-if-delete",
			composed.MarkedForDeletionPredicate,
			composed.ComposeActions(
				"cloudResources-delete",

				checkIfResourcesExist,
				deleteCrds,
				removeFinalizer,

				composed.StopAndForgetAction,
			),
			nil,
		),
		handleServed,
		addFinalizer,
		statusReady,

		composed.StopAndForgetAction,
	)
}
