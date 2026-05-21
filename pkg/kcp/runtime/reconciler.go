package runtime

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	awssecurity "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/security"
	azuresecurity "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/security"
	gcpsecurity "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/security"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type RuntimeReconciler interface {
	reconcile.Reconciler
}

type runtimeReconciler struct {
	composedStateFactory composed.StateFactory
	awsStateFactory      awssecurity.StateFactory
	azureStateFactory    azuresecurity.StateFactory
	gcpStateFactory      gcpsecurity.StateFactory
}

var _ reconcile.Reconciler = &runtimeReconciler{}

func NewRuntimeReconciler(
	composedStateFactory composed.StateFactory,
	awsStateFactory awssecurity.StateFactory,
	azureStateFactory azuresecurity.StateFactory,
	gcpStateFactory gcpsecurity.StateFactory,
) RuntimeReconciler {
	return &runtimeReconciler{
		composedStateFactory: composedStateFactory,
		awsStateFactory:      awsStateFactory,
		azureStateFactory:    azureStateFactory,
		gcpStateFactory:      gcpStateFactory,
	}
}

func (r *runtimeReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	if Ignore != nil && Ignore.ShouldIgnoreKey(request) {
		return ctrl.Result{}, nil
	}

	state := r.newState(request.NamespacedName)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("runtime", util.RequestObjToString(request)).
		WithTracker(Tracker).
		Handle(action(ctx, state))
}

func (r *runtimeReconciler) newState(ns types.NamespacedName) *State {
	return newState(
		r.composedStateFactory.NewState(ns, &infrastructuremanagerv1.Runtime{}),
	)
}

func (r *runtimeReconciler) newAction() composed.Action {
	// provider flow:
	// - guarded by a gate that keeps track of the last desired state and prevents repeated runs for the same
	// - branches into specific providers that turn security on/off for runtime and subscription
	// - on success
	//   - record successful cloud reconciliation for current desired state with the gate
	//   - sets runtime status annotations
	providerFlow := composed.If(
		defaultSecurityGate.ShouldRunPredicate,
		composed.ComposeActionsNoName(
			composed.Switch(
				nil,
				composed.NewCase(awsProviderPredicate, awssecurity.New(r.awsStateFactory)),
				composed.NewCase(azureProviderPredicate, azuresecurity.New(r.azureStateFactory)),
				composed.NewCase(gcpProviderPredicate, gcpsecurity.New(r.gcpStateFactory)),
			),
			statusReady,
		),
	)

	return composed.ComposeActionsNoName(
		feature.LoadFeatureContextFromObj(&infrastructuremanagerv1.Runtime{}),
		composed.LoadObj,
		subscriptionLoad,
		vpcNetworkLoad,
		securityEnabledDetermine,
		composed.If(
			// delete =======================================
			composed.MarkedForDeletionPredicate,
			providerFlow,
			vpcNetworkDelete,
			composed.StopAndForgetAction,
		),
		composed.If(
			// create/update =======================================
			composed.NotMarkedForDeletionPredicate,
			subscriptionCreate,
			vpcNetworkCreate,
			subscriptionWaitReady,
			vpcNetworkWaitReady,
			providerFlow,
			composed.StopAndForgetAction,
		),
	)
}
