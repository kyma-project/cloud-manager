package wafpolicy

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	awswafpolicy "github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/wafpolicy"
	"github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/wafpolicy/client"
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

	awsStateFactory := awswafpolicy.NewStateFactory(f.awsClientProvider, f.env)

	return &reconciler{
		baseStateFactory:  baseStateFactory,
		scopeStateFactory: scopeStateFactory,
		awsStateFactory:   awsStateFactory,
	}
}

type reconciler struct {
	baseStateFactory  composed.StateFactory
	scopeStateFactory commonscope.StateFactory
	awsStateFactory   awswafpolicy.StateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	state, err := r.newState(ctx, request)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating WafPolicy state: %w", err)
	}
	action := r.newAction()

	return composed.Handling().
		WithMetrics("wafpolicy", util.RequestObjToString(request)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *reconciler) newState(ctx context.Context, request reconcile.Request) (*State, error) {
	scopeState, err := r.scopeStateFactory.NewState(
		ctx,
		request.NamespacedName,
		r.baseStateFactory.NewState(request.NamespacedName, &cloudresourcesv1beta1.WafPolicy{}),
	)
	if err != nil {
		return nil, err
	}

	return &State{State: scopeState}, nil
}

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crWafPolicyMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.WafPolicy{}),
		commonscope.LoadObjWithScope(),
		composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
			composed.ComposeActions(
				"wafPolicy-create",
				actions.AddCommonFinalizer(),
				statusInitial,
				// Branch to provider-specific implementation
				composed.BuildSwitchAction(
					"providerSwitch",
					nil,
					composed.NewCase(
						awsProviderPredicate,
						r.awsProviderAction(),
					),
					// Future providers (Azure, GCP) will be added here
				),
			),
			composed.ComposeActions(
				"wafPolicy-delete",
				// Branch to provider-specific implementation
				composed.BuildSwitchAction(
					"providerSwitch",
					nil,
					composed.NewCase(
						awsProviderPredicate,
						r.awsProviderAction(),
					),
					// Future providers (Azure, GCP) will be added here
				),
				actions.RemoveCommonFinalizer(),
				composed.StopAndForgetAction,
			),
		),
		composed.StopAndForgetAction,
	)
}

func awsProviderPredicate(ctx context.Context, st composed.State) bool {
	state := st.(*State)
	return state.Scope().Spec.Provider == "aws"
}

func (r *reconciler) awsProviderAction() composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state := st.(*State)

		// Create AWS-specific state
		awsState, err := r.awsStateFactory.NewState(ctx, state)
		if err != nil {
			return err, ctx
		}

		// Execute AWS provider actions
		return awswafpolicy.New(r.awsStateFactory)(ctx, awsState)
	}
}
