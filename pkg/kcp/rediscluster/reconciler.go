package rediscluster

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

type RedisClusterReconciler interface {
	reconcile.Reconciler
}

type redisClusterReconciler struct {
	composedStateFactory composed.StateFactory
	focalStateFactory    focal.StateFactory
}

func NewRedisClusterReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,

) RedisClusterReconciler {
	return &redisClusterReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
	}
}

func (r *redisClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}

	state := r.newFocalState(req.NamespacedName)
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *redisClusterReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"main",
		focal.New(),
		func(ctx context.Context, st composed.State) (error, context.Context) {
			return composed.ComposeActions(
				"redisClusterCommon",
				loadIpRange,
			)(ctx, newState(st.(focal.State)))
		},
	)
}

func (r *redisClusterReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.RedisCluster{}),
	)
}
