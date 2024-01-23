package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ScopeReconciler interface {
	reconcile.Reconciler
}

func NewScopeReconciler(stateFactory StateFactory) ScopeReconciler {
	return &scopeReconciler{
		stateFactory: stateFactory,
	}
}

type scopeReconciler struct {
	stateFactory StateFactory
}

func (r *scopeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	state := r.newState(req)
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *scopeReconciler) newState(req ctrl.Request) *State {
	return r.stateFactory.NewState(req)
}

func (r *scopeReconciler) newAction() composed.Action {
	// Deprovisioning is done from the skr.CloudManager module CR loop
	// it will determine if all cloud resources are deleted from SKR
	// and if so will delete installed CRDs, remove finalizers from Kyma and Scope,
	// and remove skr from the looper
	//
	// This Scope/Kyma loop is handling provisioning only, and
	// if module is present in the Kyma and in Ready state it has to
	//  * add finalizer to Kyma
	//  * create Scope with finalizer
	//  * add SKR to the looper
	return composed.ComposeActions(
		"kymaMain",
		loadKyma,            // stops if Kyma not found
		findKymaModuleState, // stops if module not present

		// module is in Ready state

		addKymaFinalizer,
		loadScopeObj, // stops if Scope exists

		// scope does not exist

		createGardenerClient,
		loadShoot,
		loadGardenerCredentials,
		createScope,
		ensureScopeCommonFields,
		saveScope,

		skrActivate,

		composed.StopAndForgetAction,
	)
}
