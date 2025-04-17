package rediscluster

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	ctrl "sigs.k8s.io/controller-runtime"
)

type GcpRedisClusterReconciler interface {
	reconcile.Reconciler
}

type gcpRedisClusterReconciler struct {
	composedStateFactory composed.StateFactory
	focalStateFactory    focal.StateFactory

	stateFactory StateFactory
}

func NewGcpRedisClusterReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	stateFactory StateFactory,
) GcpRedisClusterReconciler {
	return &gcpRedisClusterReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
		stateFactory:         stateFactory,
	}
}

func (r *gcpRedisClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}

	state := r.newFocalState(req.NamespacedName)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("kcpgcprediscluster", util.RequestObjToString(req)).
		Handle(action(ctx, state))
}

func (r *gcpRedisClusterReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"main",
		feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.GcpRedisCluster{}),
		focal.New(),
		r.newFlow(),
	)
}

func (r *gcpRedisClusterReconciler) newFlow() composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state, err := r.stateFactory.NewState(ctx, st.(focal.State))
		if err != nil {
			composed.LoggerFromCtx(ctx).Error(err, "Failed to bootstrap GCP RedisCluster state")
			redisCluster := st.Obj().(*cloudcontrolv1beta1.GcpRedisCluster)
			redisCluster.Status.State = cloudcontrolv1beta1.StateError
			return composed.UpdateStatus(redisCluster).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
					Message: "Failed to create GCP RedisCluster state",
				}).
				SuccessError(composed.StopAndForget).
				SuccessLogMsg(fmt.Sprintf("Error creating new GCP RedisCluster state: %s", err)).
				Run(ctx, st)
		}

		return composed.ComposeActions(
			"gcpRedisCluster",
			loadSubnet,
			actions.AddCommonFinalizer(),
			loadRedis,
			composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"gcpRedisCluster-create",
					createRedis,
					updateStatusId,
					addUpdatingCondition,
					waitRedisAvailable,
					loadCertificates,
					modifyNodeType,
					modifyShardCount,
					modifyShardCount,
					modifyRedisConfigs,
					updateRedis,
					updateStatus,
				),
				composed.ComposeActions(
					"gcpRedisCluster-delete",
					removeReadyCondition,
					deleteRedis,
					waitRedisDeleted,
					actions.RemoveCommonFinalizer(),
					composed.StopAndForgetAction,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}

func (r *gcpRedisClusterReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.GcpRedisCluster{}),
	)
}
