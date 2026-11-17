package cloudresources

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/statewithscope"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	awscloudresources "github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/cloudresources"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewReconcilerFactory(awsStateFactory awscloudresources.StateFactory) skrruntime.ReconcilerFactory {
	return &reconcilerFactory{
		awsStateFactory: awsStateFactory,
	}
}

type reconcilerFactory struct {
	awsStateFactory awscloudresources.StateFactory
}

func (f *reconcilerFactory) New(args skrruntime.ReconcilerArguments) reconcile.Reconciler {
	return &reconciler{
		factory: newStateFactory(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(args.SkrCluster)),
			commonscope.NewStateFactory(
				composed.NewStateClusterFromCluster(args.KcpCluster),
				args.ScopeProvider,
			),
			args.Provider,
			args.IgnoreWatchErrors,
		),
		awsStateFactory: f.awsStateFactory,
	}
}

type reconciler struct {
	factory *stateFactory

	awsStateFactory awscloudresources.StateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	state, err := r.factory.NewState(ctx, request)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating CloudResources state: %w", err)
	}
	action := r.newAction()

	return composed.Handling().
		WithMetrics("cloudresources", util.RequestObjToString(request)).
		WithNoLog().
		Handle(action(ctx, state))
}

// When module is deleted from the SKR Kyma spec
// Then module CR will get deletionTimestamp

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActionsNoName(
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.CloudResources{}),
		commonscope.New(),

		composed.If(
			composed.MarkedForDeletionPredicate,
			statusDeleting,
			checkIfResourcesExist,
			deleteCrds,
			removeFinalizer,
			composed.StopAndForgetAction,
		),

		composed.If(
			composed.NotMarkedForDeletionPredicate,
			handleServed,
			addFinalizer,
			composed.Switch(
				nil,
				composed.NewCase(statewithscope.AwsProviderPredicate, awscloudresources.New(r.awsStateFactory)),
			),
			statusReady,
			composed.StopAndForgetAction,
		),

		composed.StopAndForgetAction,
	)
}
