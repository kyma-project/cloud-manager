package managedredis

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/common/statewithscope"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	azuremanagedredis "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/managedredis"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

type ManagedRedisReconciler interface {
	reconcile.Reconciler
}

type managedRedisReconciler struct {
	composedStateFactory composed.StateFactory
	focalStateFactory    focal.StateFactory

	azureStateFactory azuremanagedredis.StateFactory
}

func NewManagedRedisReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	azureStateFactory azuremanagedredis.StateFactory,
) ManagedRedisReconciler {
	return &managedRedisReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
		azureStateFactory:    azureStateFactory,
	}
}

func (r *managedRedisReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}

	state := r.newFocalState(req.NamespacedName)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("managedredis", util.RequestObjToString(req)).
		Handle(action(ctx, state))
}

func (r *managedRedisReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"main",
		feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.AzureManagedRedis{}),
		focal.New(),
		func(ctx context.Context, st composed.State) (error, context.Context) {
			return composed.ComposeActions(
				"managedRedisCommon",
				loadIpRange,
				composed.BuildSwitchAction(
					"providerSwitch",
					nil,
					composed.NewCase(statewithscope.AzureProviderPredicate, azuremanagedredis.New(r.azureStateFactory)),
				),
			)(ctx, newState(st.(focal.State)))
		},
	)
}

func (r *managedRedisReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.AzureManagedRedis{}),
	)
}
