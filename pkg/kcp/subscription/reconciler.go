package subscription

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	subscriptionclient "github.com/kyma-project/cloud-manager/pkg/kcp/subscription/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SubscriptionReconciler interface {
	reconcile.Reconciler
}

func New(
	mgr manager.Manager,
	awsStsClientProvider awsclient.GardenClientProvider[subscriptionclient.AwsStsClient],
) SubscriptionReconciler {
	return &subscriptionReconciler{
		stateFactory: NewStateFactory(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(mgr)),
			awsStsClientProvider,
		),
	}
}

type subscriptionReconciler struct {
	stateFactory StateFactory
}

func (r *subscriptionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if Ignore != nil && Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}

	state := r.newState(req)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("subscription", util.RequestObjToString(req)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *subscriptionReconciler) newState(req ctrl.Request) *State {
	return r.stateFactory.NewState(req)
}

func (r *subscriptionReconciler) newAction() composed.Action {
	return composed.ComposeActionsNoName(
		composed.LoadObj,
		composed.If(
			composed.MarkedForDeletionPredicate,
			// being deleted
			statusDeleting,
			resourcesLoad,
			statusSaveOnDelete,
			actions.PatchRemoveCommonFinalizer(),
			composed.StopAndForgetAction,
		),
		composed.If(
			// no update, once Subscription gets Ready condition, do not reconcile
			func(ctx context.Context, st composed.State) bool {
				state := st.(*State)
				if composed.IsMarkedForDeletion(state.Obj()) {
					return false
				}
				if meta.IsStatusConditionTrue(state.ObjAsSubscription().Status.Conditions, cloudcontrolv1beta1.ConditionTypeSubscription) {
					return false
				}
				return true
			},
			// create
			actions.PatchAddCommonFinalizer(),
			statusInitial,
			composed.IfElse(
				isGardenerSubscription,
				composed.ComposeActionsNoName(
					checkIfGardenSubscriptionType,
					gardenerClientCreate,
					gardenerCredentialsRead,
					statusSaveOnCreate,
					composed.StopAndForgetAction,
				),
				composed.ComposeActionsNoName(
					handleNonGardenerSubscriptionType,
					composed.StopAndForgetAction,
				),
			),
		),
	)
}

func isGardenerSubscription(ctx context.Context, st composed.State) bool {
	state := st.(*State)
	return state.ObjAsSubscription().Spec.Details.Garden != nil
}
