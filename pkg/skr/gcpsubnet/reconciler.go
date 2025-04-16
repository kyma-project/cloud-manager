package gcpsubnet

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/util"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
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
		WithMetrics("skrgcpsubnet", util.RequestObjToString(request)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"gcpSubnet",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.GcpSubnet{}),
		composed.LoadObj,

		updateId,
		loadKcpGcpSubnet,

		composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
			composed.ComposeActions(
				"gcpSubnet-create",
				createKcpGcpSubnet,
				waitKcpStatusUpdate,
				updateStatus,
				actions.AddCommonFinalizer(),
			),
			composed.ComposeActions(
				"gcpSubnet-delete",
				// preventDeleteOnGcpRedisClusterUsage,
				deleteKcpGcpSubnet,
				waitKcpGcpSubnetDeleted,
				actions.RemoveCommonFinalizer(),
				composed.StopAndForgetAction,
			),
		),

		composed.StopAndForgetAction,
	)
}
