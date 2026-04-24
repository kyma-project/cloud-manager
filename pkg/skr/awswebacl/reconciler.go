package awswebacl

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"github.com/kyma-project/cloud-manager/pkg/skr/awswebacl/client"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewReconcilerFactory(
	awsClientProvider awsclient.SkrClientProvider[client.Client],
	env abstractions.Environment,
) skrruntime.ReconcilerFactory {
	return &reconcilerFactory{
		awsClientProvider: awsClientProvider,
		env:               env,
	}
}

type reconcilerFactory struct {
	awsClientProvider awsclient.SkrClientProvider[client.Client]
	env               abstractions.Environment
}

func (f *reconcilerFactory) New(args skrruntime.ReconcilerArguments) reconcile.Reconciler {
	baseStateFactory := composed.NewStateFactory(composed.NewStateClusterFromCluster(args.SkrCluster))

	scopeStateFactory := commonscope.NewStateFactory(
		composed.NewStateClusterFromCluster(args.KcpCluster),
		args.ScopeProvider,
	)

	return &reconciler{
		factory: newStateFactory(
			baseStateFactory,
			scopeStateFactory,
			f.awsClientProvider,
			f.env,
		),
	}
}

type reconciler struct {
	factory *stateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	state, err := r.factory.NewState(ctx, request)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating AwsWebAcl state: %w", err)
	}
	action := r.newAction()

	return composed.Handling().
		WithMetrics("awswebacl", util.RequestObjToString(request)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crAwsWebAclMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.AwsWebAcl{}),
		commonscope.LoadObjWithScope(),
		createAwsClient,
		loadWebAcl,
		composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
			composed.ComposeActions(
				"awsWebAcl-create",
				actions.AddCommonFinalizer(),
				statusInitial,
				createWebAcl,
				checkUpdateNeeded,
				updateWebAcl,
				// Logging configuration (after WebACL creation)
				loadLoggingConfiguration,
				ensureLogGroup,
				configureLogging,
				updateStatus,
			),
			composed.ComposeActions(
				"awsWebAcl-delete",
				// Logging cleanup (before WebACL deletion)
				deleteLoggingConfiguration,
				deleteLogGroupIfManaged,
				deleteWebAcl,
				actions.RemoveCommonFinalizer(),
				composed.StopAndForgetAction,
			),
		),
		composed.StopAndForgetAction,
	)
}
