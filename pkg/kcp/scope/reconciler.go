package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
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
		"scopeMain",
		loadKyma, // stops if Kyma not found
		findKymaModuleState,
		loadScopeObj,

		composed.BuildSwitchAction(
			"scope-switch",
			nil,

			// module is disabled
			composed.NewCase(
				predicateShouldDisable(),
				composed.ComposeActions(
					"scope-disable",
					removeKymaFinalizer,
					skrDeactivate, // if module not present in status, remove kymaName from looper, delete scope, and stop and forget
				),
			),

			// module is enabled
			composed.NewCase(
				predicateShouldEnable(),
				composed.ComposeActions(
					"scope-enable",
					// module exists in Kyma status in some state (processing, ready, deleting, warning)
					// scope:
					//   * does not exist - has to be created
					//   * exist but waiting for api to be activated

					addKymaFinalizer,
					loadNetworks,
					composed.If(
						// scopeAlreadyCreatedBranching
						composed.All(ObjIsLoadedPredicate(), composed.Not(UpdateIsNeededPredicate())),
						composed.ComposeActions( // This is called only if scope exits, and it does not need to be updated.
							"enableApisAndActivateSkr",
							enableApis,
							addReadyCondition,
							skrActivate,
							composed.StopAndForgetAction),
					),
					// scope does not exist or needs to be updated
					createGardenerClient,
					findShootName,
					loadShoot,
					loadGardenerCredentials,
					createScope,
					ensureScopeCommonFields,
					saveScope,
					createKymaNetworkReference,
					composed.StopWithRequeueAction, // enableApisAndActivateSkr will be called in the next loop
				),
			),
		),
	)
}

func ObjIsLoadedPredicate() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		obj := st.Obj()
		return obj != nil && obj.GetName() != "" // empty object is created when state gets created
	}
}

func UpdateIsNeededPredicate() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		if ObjIsLoadedPredicate()(ctx, st) {
			state := st.(*State)
			if state.allNetworks.FindFirstByType(cloudcontrolv1beta1.NetworkTypeKyma) == nil {
				return true
			}
			switch state.ObjAsScope().Spec.Provider {
			case cloudcontrolv1beta1.ProviderGCP:
				return state.ObjAsScope().Spec.Scope.Gcp.Workers == nil || len(state.ObjAsScope().Spec.Scope.Gcp.Workers) == 0
			case cloudcontrolv1beta1.ProviderAzure:
				return state.ObjAsScope().Spec.Scope.Azure.Network.Nodes == ""
			default:
				return false
			}
		} else {
			return false
		}
	}
}
