package alicloudrediscluster

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/util"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"github.com/kyma-project/cloud-manager/pkg/skr/common/defaultiprange"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
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
		return ctrl.Result{}, fmt.Errorf("error creating AlicloudRedisCluster state: %w", err)
	}
	action := r.newAction()

	return composed.Handling().
		WithMetrics("alicloudrediscluster", util.RequestObjToString(request)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"alicloudRedisCluster",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.AlicloudRedisCluster{}),
		composed.LoadObj,

		defaultiprange.New(),

		updateId,
		loadKcpRedisCluster,
		loadAuthSecret,

		composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
			composed.ComposeActions(
				"alicloudRedisCluster-create",
				actions.AddCommonFinalizer(),
				createKcpRedisCluster,
				waitKcpStatusUpdate,
				updateStatus,
				waitSkrStatusReady,
				modifyKcpRedisCluster,
				createAuthSecret,
				loadAuthSecret,
				modifyAuthSecret,
			),
			composed.ComposeActions(
				"alicloudRedisCluster-delete",
				removeAuthSecretFinalizer,
				deleteAuthSecret,
				waitAuthSecretDeleted,
				deleteKcpRedisCluster,
				waitKcpRedisClusterDeleted,
				actions.RemoveCommonFinalizer(),
				composed.StopAndForgetAction,
			),
		),

		composed.StopAndForgetAction,
	)
}
