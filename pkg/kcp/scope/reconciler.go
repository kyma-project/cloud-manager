package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/statewithscope"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	azureexposeddata "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/exposedData"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	scopeclient "github.com/kyma-project/cloud-manager/pkg/kcp/scope/client"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

type ScopeReconciler interface {
	reconcile.Reconciler
}

func New(
	mgr manager.Manager,
	awsStsClientProvider awsclient.GardenClientProvider[scopeclient.AwsStsClient],
	activeSkrCollection skrruntime.ActiveSkrCollection,
	// keep gcpServiceUsageClientProvider separate from the expose data client, since SRE will start enabling APIs soon,
	// so gcpServiceUsageClientProvider will be removed completely
	gcpServiceUsageClientProvider gcpclient.ClientProvider[gcpclient.ServiceUsageClient],
	azureStateFactory azureexposeddata.StateFactory,
) ScopeReconciler {
	return NewScopeReconciler(
		NewStateFactory(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(mgr)),
			activeSkrCollection,
			awsStsClientProvider,
			gcpServiceUsageClientProvider,
		),
		azureStateFactory,
	)
}

func NewScopeReconciler(
	stateFactory StateFactory,
	azureStateFactory azureexposeddata.StateFactory,
) ScopeReconciler {
	return &scopeReconciler{
		stateFactory:      stateFactory,
		azureStateFactory: azureStateFactory,
	}
}

type scopeReconciler struct {
	stateFactory StateFactory

	azureStateFactory azureexposeddata.StateFactory
}

func (r *scopeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if Ignore != nil && Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}

	state := r.newState(req)
	action := r.newAction()

	// Scope reconciler is triggered very often due to KLM constant changes on watched Kyma
	// HandleWithoutLogging should be used, so no reconciliation outcome is logged since it most cases
	// the reconciler will do nothing since no change regarding CloudManager was done on Kyma
	// so it will just produce an unnecessary log entry "Reconciliation finished without control error - doing stop and forget"
	// To accommodate this non-functional requirement to keep logs tidy and prevent excessive and not so usable log entries
	// in cases when Scope actually did something we have to accept the discomfort of not having this log entry
	return composed.Handling().
		WithMetrics("scope", util.RequestObjToString(req)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *scopeReconciler) newState(req ctrl.Request) *State {
	return r.stateFactory.NewState(req)
}

func (r *scopeReconciler) newAction() composed.Action {
	return composed.ComposeActionsNoName(
		composed.LoadObjNoStopIfNotFound, // loads Scope
		providerFromScopeToState,
		gardenerClusterLoad,
		networkReferenceKymaLoad,
		gardenerClusterExtractShootName,
		logScope,

		composed.IfElse(
			shouldScopeExist,

			composed.ComposeActionsNoName(
				// scope should EXIST
				composed.If(
					isScopeCreateOrUpdateNeeded,
					gardenerClientCreate,
					shootNameMustHave,
					shootLoad,
					gardenerCredentialsLoad,
					scopeCreate,
					scopeEnsureCommonFields,
					scopeSave,
				),
				networkReferenceKymaCreate,
				networkReferenceKymaWaitReady,
				apiEnable,

				// collect exposed data from cloud providers
				composed.If(
					isExposedDataReadNeeded,
					composed.Switch(
						nil,
						composed.NewCase(statewithscope.AwsProviderPredicate, composed.Noop),
						composed.NewCase(statewithscope.AzureProviderPredicate, azureexposeddata.New(r.azureStateFactory)),
						composed.NewCase(statewithscope.GcpProviderPredicate, composed.Noop),
					),
					exposedDataSave,
				),

				conditionReady,
			),

			composed.ComposeActionsNoName(
				// scope should NOT exist

				// just in case stop the SKR Looper,
				// but kyma should not exist at this point
				// and should have been already removed
				skrDeactivate,
				networkReferenceKymaDelete,
				// nuke reconciler deletes all resources in the scope
				// and their reconcilers require Scope to be able to reach cloud api
				// thus we first must create the Nuke and wait until it's completed
				// only then it's safe to delete the scope
				nukeLoad,
				nukeCreate,
				nukeWaitCompleted,
				// now we can safely delete the Scope only when all resources have been nuked
				scopeDelete,
				composed.StopAndForgetAction,
			),
		),
	)
}

func shouldScopeExist(_ context.Context, st composed.State) bool {
	state := st.(*State)

	if state.gardenerCluster == nil {
		return false
	}
	if composed.IsMarkedForDeletion(state.gardenerCluster) {
		return false
	}

	return true
}

func isScopeCreateOrUpdateNeeded(ctx context.Context, st composed.State) bool {
	state := st.(*State)

	if !composed.IsObjLoaded(ctx, st) {
		// scope does not exist
		return true
	}

	// check if labels from GardenerCluster are copied to Scope
	for _, label := range cloudcontrolv1beta1.ScopeLabels {
		if _, ok := state.ObjAsScope().Labels[label]; !ok {
			return true
		}
	}

	// check if GCP scope needs to be updated with worker info from shoot
	if state.ObjAsScope().Spec.Provider == cloudcontrolv1beta1.ProviderGCP {
		if len(state.ObjAsScope().Spec.Scope.Gcp.Workers) == 0 {
			return true
		}
	}

	// check if Azure scope needs to be updated with nodes info from shoot
	if state.ObjAsScope().Spec.Provider == cloudcontrolv1beta1.ProviderAzure {
		if state.ObjAsScope().Spec.Scope.Azure.Network.Nodes == "" {
			return true
		}
	}

	return false
}

func isExposedDataReadNeeded(ctx context.Context, st composed.State) bool {
	if !feature.ExposeData.Value(ctx) {
		return false
	}
	if !composed.IsObjLoaded(ctx, st) {
		return false
	}
	state := st.(*State)
	if state.ObjAsScope().Status.ExposedData.ReadTime == nil {
		return true
	}

	diff := time.Since(state.ObjAsScope().Status.ExposedData.ReadTime.Time)

	return diff > time.Hour
}
