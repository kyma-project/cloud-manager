package managedredis

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	kcpcommonaction "github.com/kyma-project/cloud-manager/pkg/kcp/commonAction"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azuremanagedredisclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/managedredis/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

type ManagedRedisReconciler interface {
	reconcile.Reconciler
}

type managedRedisReconciler struct {
	composedStateFactory  composed.StateFactory
	kcpCommonStateFactory kcpcommonaction.StateFactory
	clientProvider        azureclient.ClientProvider[azuremanagedredisclient.Client]
}

func NewManagedRedisReconciler(
	composedStateFactory composed.StateFactory,
	kcpCommonStateFactory kcpcommonaction.StateFactory,
	clientProvider azureclient.ClientProvider[azuremanagedredisclient.Client],
) ManagedRedisReconciler {
	return &managedRedisReconciler{
		composedStateFactory:  composedStateFactory,
		kcpCommonStateFactory: kcpCommonStateFactory,
		clientProvider:        clientProvider,
	}
}

func (r *managedRedisReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}

	kcpCommonState := r.newKcpCommonState(req.NamespacedName)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("azuremanagedredis", util.RequestObjToString(req)).
		Handle(action(ctx, kcpCommonState))
}

func (r *managedRedisReconciler) newKcpCommonState(name types.NamespacedName) kcpcommonaction.State {
	return r.kcpCommonStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.AzureManagedRedis{}),
	)
}

func (r *managedRedisReconciler) newAction() composed.Action {
	return composed.ComposeActionsNoName(
		feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.AzureManagedRedis{}),
		kcpcommonaction.New(),
		func(ctx context.Context, st composed.State) (error, context.Context) {
			return composed.ComposeActionsNoName(
				actions.AddCommonFinalizer(),
				initAzureClient(r.clientProvider),
				loadManagedRedis,
				loadDatabase,
				loadPrivateEndPoint,
				loadPrivateDnsZoneGroup,
				loadPrivateDnsZone,
				loadVirtualNetworkLink,
				composed.If(composed.NotMarkedForDeletionPredicate,
					composed.ComposeActionsNoName(
						createManagedRedis,
						waitManagedRedisAvailable,
						updateStatusId,
						createDatabase,
						waitDatabaseAvailable,
						createPrivateEndPoint,
						waitPrivateEndPointAvailable,
						createPrivateDnsZone,
						waitPrivateDnsZoneAvailable,
						createVirtualNetworkLink,
						waitVirtualNetworkLinkAvailable,
						createPrivateDnsZoneGroup,
						updateStatus,
						composed.StopAndForgetAction,
					),
				),
				composed.If(composed.MarkedForDeletionPredicate,
					composed.ComposeActionsNoName(
						deletePrivateDnsZoneGroup,
						waitPrivateDnsZoneGroupDeleted,
						deletePrivateEndPoint,
						waitPrivateEndPointDeleted,
						deleteDatabase,
						waitDatabaseDeleted,
						deleteManagedRedis,
						waitManagedRedisDeleted,
						actions.RemoveCommonFinalizer(),
						composed.StopAndForgetAction,
					),
				),
			)(ctx, newState(st.(kcpcommonaction.State)))
		},
	)
}
