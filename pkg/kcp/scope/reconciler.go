package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	scopeclient "github.com/kyma-project/cloud-manager/pkg/kcp/scope/client"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ScopeReconciler interface {
	reconcile.Reconciler
}

func New(
	mgr manager.Manager,
	awsStsClientProvider awsclient.GardenClientProvider[scopeclient.AwsStsClient],
	activeSkrCollection skrruntime.ActiveSkrCollection,
	gcpServiceUsageClientProvider gcpclient.ClientProvider[gcpclient.ServiceUsageClient],
) ScopeReconciler {
	return NewScopeReconciler(NewStateFactory(
		composed.NewStateFactory(composed.NewStateClusterFromCluster(mgr)),
		awsStsClientProvider,
		activeSkrCollection,
		gcpServiceUsageClientProvider,
	))
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
	if Ignore != nil && Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}

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
		loadScopeObj,
		composed.BuildBranchingAction(
			"scopeAlreadyCreatedBranching",
			ObjIsLoadedPredicate(),
			composed.ComposeActions( // This is called only if scope exits.
				"enableApisAndActivateSkr",
				enableApis,
				addReadyCondition,
				skrActivate,
				composed.StopAndForgetAction),
			nil),
		// scope does not exist
		createGardenerClient,
		findShootName,
		loadShoot,
		loadGardenerCredentials,
		createScope,
		ensureScopeCommonFields,
		saveScope,
		composed.StopWithRequeueAction, // enableApisAndActivateSkr will be called in the next loop
	)
}

func ObjIsLoadedPredicate() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		obj := st.Obj()
		return obj != nil && obj.GetName() != "" // empty object is created when state gets created
	}
}
