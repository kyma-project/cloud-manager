package awsvpcpeering

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
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
			args.ScopeProvider,
			composed.NewStateClusterFromCluster(args.KcpCluster),
		),
	}
}

type reconciler struct {
	factory *stateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	state, err := r.factory.NewState(ctx, request)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating AwsVpcPeering state: %w", err)
	}
	action := r.newAction()

	return composed.Handling().
		WithMetrics("awsvpcpeering", util.RequestObjToString(request)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crAwsVpcPeeringMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.AwsVpcPeering{}),
		composed.LoadObj,
		addFinalizer,
		updateId,
		loadKcpRemoteNetwork,
		createKcpRemoteNetwork,
		waitNetworkReady,
		loadKcpAwsVpcPeering,
		createKcpVpcPeering,
		deleteKcpVpcPeering,
		waitKcpVpcPeeringDeleted,
		deleteKcpRemoteNetwork,
		removeFinalizer,
		updateStatus,
		waitStatusActive,
		composed.StopAndForgetAction,
	)
}
