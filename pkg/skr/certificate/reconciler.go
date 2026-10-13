package certificate

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	awscertprovider "github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/certificate"
	certificateclient "github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/certificate/client"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewReconcilerFactory(
	awsClientProvider awsclient.SkrClientProvider[certificateclient.Client],
	env abstractions.Environment,
) skrruntime.ReconcilerFactory {
	return &reconcilerFactory{
		awsClientProvider: awsClientProvider,
		env:               env,
	}
}

type reconcilerFactory struct {
	awsClientProvider awsclient.SkrClientProvider[certificateclient.Client]
	env               abstractions.Environment
}

func (f *reconcilerFactory) New(args skrruntime.ReconcilerArguments) reconcile.Reconciler {
	baseStateFactory := composed.NewStateFactory(composed.NewStateClusterFromCluster(args.SkrCluster))

	scopeStateFactory := commonscope.NewStateFactory(
		composed.NewStateClusterFromCluster(args.KcpCluster),
		args.ScopeProvider,
	)

	awsStateFactory := awscertprovider.NewStateFactory(f.awsClientProvider, f.env)

	return &reconciler{
		baseStateFactory:  baseStateFactory,
		scopeStateFactory: scopeStateFactory,
		awsStateFactory:   awsStateFactory,
	}
}

type reconciler struct {
	baseStateFactory  composed.StateFactory
	scopeStateFactory commonscope.StateFactory
	awsStateFactory   awscertprovider.StateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	state, err := r.newState(ctx, request)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating Certificate state: %w", err)
	}
	action := r.newAction()

	return composed.Handling().
		WithMetrics("certificate", util.RequestObjToString(request)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *reconciler) newState(ctx context.Context, request reconcile.Request) (*State, error) {
	scopeState, err := r.scopeStateFactory.NewState(
		ctx,
		request.NamespacedName,
		r.baseStateFactory.NewState(request.NamespacedName, &cloudresourcesv1beta1.Certificate{}),
	)
	if err != nil {
		return nil, err
	}

	return newState(scopeState).(*State), nil
}

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crCertificateMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.Certificate{}),
		commonscope.LoadObjWithScope(),
		statusInitial,
		composed.BuildSwitchAction(
			"providerSwitch",
			nil,
			composed.NewCase(
				awsProviderPredicate,
				r.awsProviderAction(),
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
	return awscertprovider.New(r.awsStateFactory)
}
