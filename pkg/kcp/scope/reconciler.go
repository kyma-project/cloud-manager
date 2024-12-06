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
	// Deprovisioning is started from the skr.CloudResources module CR loop
	// when it will determine that module CR has deletion timestamp,
	/// if all cloud resources are deleted from SKR,
	// it will delete installed CRDs, and remove own finalizer
	// KLM will figure it's deleted and remove it from the KCP Kyma status
	// This KCP Scope reconciler will then detect cloud-manager module is not present
	// in the Kyma status, and it will remove Kyma finalizer and deactivate SKR
	// and remove it from the SKR runtime loop
	//
	// This Scope/Kyma loop should:
	// * if module is present in the Kyma and in Ready state:
	//   * add finalizer to Kyma
	//   * create Scope with finalizer
	//   * add SKR to the looper
	// * if module is not present in the Kyma status:
	//   * remove Kyma finalizer
	//   * deactivate SKR and remove it from the SKR runtime loop
	return composed.ComposeActions(
		"scopeMain",
		loadScopeObj,
		loadNetworks,
		loadKyma, // stops if Kyma not found
		findKymaModuleState,

		// module is disabled
		// if module not present in status, remove kymaName from looper, delete scope, and stop and forget
		composed.If(
			predicateShouldDisable(),
			removeKymaFinalizer,
			skrDeactivate,
			kymaNetworkReferenceDelete,
			scopeDelete,
			composed.StopAndForgetAction,
		),

		// module exists in Kyma status in some state (processing, ready, deleting, warning)
		// scope:
		//   * does not exist - has to be created
		//   * exists but waiting for api to be activated
		composed.If(
			predicateShouldEnable(),
			addKymaFinalizer,

			// scope does not exist or needs to be updated
			composed.If(
				predicateScopeCreateOrUpdateNeeded(),
				createGardenerClient,
				findShootName,
				loadShoot,
				loadGardenerCredentials,
				scopeCreate,
				ensureScopeCommonFields,
				scopeSave,
				kymaNetworkReferenceCreate,
			),

			enableApis,
			addReadyCondition,
			skrActivate,
		),
	)
}
