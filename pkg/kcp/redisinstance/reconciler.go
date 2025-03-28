package redisinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/util"

	awsRedisinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/redisinstance"
	azureRedisinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/redisinstance"
	gcpRedisinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/redisinstance"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

type RedisInstanceReconciler interface {
	reconcile.Reconciler
}

type redisInstanceReconciler struct {
	composedStateFactory composed.StateFactory
	focalStateFactory    focal.StateFactory

	gcpStateFactory   gcpRedisinstance.StateFactory
	azureStateFactory azureRedisinstance.StateFactory
	awsStateFactory   awsRedisinstance.StateFactory
}

func NewRedisInstanceReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	gcpStateFactory gcpRedisinstance.StateFactory,
	azureStateFactory azureRedisinstance.StateFactory,
	awsStateFactory awsRedisinstance.StateFactory,
) RedisInstanceReconciler {
	return &redisInstanceReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
		gcpStateFactory:      gcpStateFactory,
		azureStateFactory:    azureStateFactory,
		awsStateFactory:      awsStateFactory,
	}
}

func (r *redisInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}

	state := r.newFocalState(req.NamespacedName)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("redisinstance", util.RequestObjToString(req)).
		Handle(action(ctx, state))
}

func (r *redisInstanceReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"main",
		focal.New(),
		func(ctx context.Context, st composed.State) (error, context.Context) {
			return composed.ComposeActions(
				"redisInstanceCommon",
				loadIpRange,
				composed.BuildSwitchAction(
					"providerSwitch",
					nil,
					composed.NewCase(focal.GcpProviderPredicate, gcpRedisinstance.New(r.gcpStateFactory)),
					composed.NewCase(focal.AzureProviderPredicate, azureRedisinstance.New(r.azureStateFactory)),
					composed.NewCase(focal.AwsProviderPredicate, awsRedisinstance.New(r.awsStateFactory)),
				),
			)(ctx, newState(st.(focal.State)))
		},
	)
}

func (r *redisInstanceReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.RedisInstance{}),
	)
}
