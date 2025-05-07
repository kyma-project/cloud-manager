package gcprediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/util"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"github.com/kyma-project/cloud-manager/pkg/skr/common/defaultgcpsubnet"

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
		WithMetrics("gcprediscluster", util.RequestObjToString(request)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"gcpRedisCluster",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.GcpRedisCluster{}),
		composed.LoadObj,

		defaultgcpsubnet.New(),

		updateId,
		loadKcpGcpRedisCluster,
		loadAuthSecret,

		composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
			composed.ComposeActions(
				"gcpRedisCluster-create",
				actions.AddCommonFinalizer(),
				createKcpGcpRedisCluster,
				modifyKcpGcpRedisCluster,
				waitKcpStatusUpdate,
				updateStatus,
				waitSkrStatusReady,
				createAuthSecret,
				modifyAuthSecret,
			),
			composed.ComposeActions(
				"gcpRedisCluster-delete",
				removeAuthSecretFinalizer,
				deleteAuthSecret,
				waitAuthSecretDeleted,
				deleteKcpGcpRedisCluster,
				waitKcpGcpRedisClusterDeleted,
				actions.RemoveCommonFinalizer(),
				composed.StopAndForgetAction,
			),
		),

		composed.StopAndForgetAction,
	)
}
