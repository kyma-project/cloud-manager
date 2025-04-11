package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/common/statewithscope"
	"github.com/kyma-project/cloud-manager/pkg/util"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsrediscluster "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/rediscluster"
	azurerediscluster "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/rediscluster"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

type RedisClusterReconciler interface {
	reconcile.Reconciler
}

type redisClusterReconciler struct {
	composedStateFactory composed.StateFactory
	focalStateFactory    focal.StateFactory

	awsStateFactory   awsrediscluster.StateFactory
	azureStateFactory azurerediscluster.StateFactory
}

func NewRedisClusterReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	awsStateFactory awsrediscluster.StateFactory,
	azureStateFactory azurerediscluster.StateFactory,
) RedisClusterReconciler {
	return &redisClusterReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
		awsStateFactory:      awsStateFactory,
		azureStateFactory:    azureStateFactory,
	}
}

func (r *redisClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}

	state := r.newFocalState(req.NamespacedName)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("rediscluster", util.RequestObjToString(req)).
		Handle(action(ctx, state))
}

func (r *redisClusterReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"main",
		focal.New(),
		func(ctx context.Context, st composed.State) (error, context.Context) {
			return composed.ComposeActions(
				"redisClusterCommon",
				loadIpRange,
				composed.BuildSwitchAction(
					"providerSwitch",
					nil,
					composed.NewCase(statewithscope.AwsProviderPredicate, awsrediscluster.New(r.awsStateFactory)),
					composed.NewCase(statewithscope.AzureProviderPredicate, azurerediscluster.New(r.azureStateFactory)),
				),
			)(ctx, newState(st.(focal.State)))
		},
	)
}

func (r *redisClusterReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.RedisCluster{}),
	)
}
