package nuke

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	awsnuke "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nuke"
	azurenuke "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/nuke"
	gcpnuke "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nuke"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type NukeReconciler interface {
	reconcile.Reconciler
}

func New(
	baseStateFactory composed.StateFactory,
	activeSkrCollection skrruntime.ActiveSkrCollection,
	gcpStateFactory gcpnuke.StateFactory,
	awsStateFactory awsnuke.StateFactory,
	azureStateFactory azurenuke.StateFactory,
) NukeReconciler {
	return &nukeReconciler{
		stateFactory: NewStateFactory(
			baseStateFactory,
			focal.NewStateFactory(),
			activeSkrCollection,
		),
		gcpStateFactory:   gcpStateFactory,
		awsStateFactory:   awsStateFactory,
		azureStateFactory: azureStateFactory,
	}
}

type nukeReconciler struct {
	stateFactory      StateFactory
	gcpStateFactory   gcpnuke.StateFactory
	awsStateFactory   awsnuke.StateFactory
	azureStateFactory azurenuke.StateFactory
}

func (r *nukeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	state := r.stateFactory.NewState(req)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("nuke", util.RequestObjToString(req)).
		Handle(action(ctx, state))
}

func (r *nukeReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"nukeMain",
		feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.Nuke{}),
		focal.NewWithOptionalScope(),
		composed.If(
			// if Nuke not marked for deletion
			composed.Not(composed.MarkedForDeletionPredicate),
			shortCircuitCompleted,
			loadResources,
			resourceStatusDiscovered,
			deleteResources,
			resourceStatusDeleting,
			resourceStatusDeleted,
			composed.If(
				composed.All(
					feature.FFNukeBackupsGcp.Predicate(),
					focal.GcpProviderPredicate,
				),
				gcpnuke.New(r.gcpStateFactory),
			),
			composed.If(
				composed.All(
					feature.FFNukeBackupsAws.Predicate(),
					focal.AwsProviderPredicate,
				),
				awsnuke.New(r.awsStateFactory),
			),
			composed.If(
				composed.All(
					feature.FFRwxBackupAzure.Predicate(),
					focal.AzureProviderPredicate,
				),
				azurenuke.New(r.azureStateFactory),
			),
			checkIfAllDeleted,
			scopeDelete,
			statusCompleted,
		),
	)
}
