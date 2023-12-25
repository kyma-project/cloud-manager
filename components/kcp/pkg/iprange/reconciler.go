package iprange

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/scope"
	awsiprange "github.com/kyma-project/cloud-resources/components/kcp/pkg/provider/aws/iprange"
	azureiprange "github.com/kyma-project/cloud-resources/components/kcp/pkg/provider/azure/iprange"
	gcpiprange "github.com/kyma-project/cloud-resources/components/kcp/pkg/provider/gcp/iprange"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

type IPRangeReconciler struct {
	composedStateFactory composed.StateFactory
	focalStateFactory    focal.StateFactory
	scopeStateFactory    scope.StateFactory

	awsStateFactory   awsiprange.StateFactory
	azureStateFactory azureiprange.StateFactory
	gcpStateFactory   gcpiprange.StateFactory
}

func NewIPRangeReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	scopeStateFactory scope.StateFactory,
	awsStateFactory awsiprange.StateFactory,
	azureStateFactory azureiprange.StateFactory,
	gcpStateFactory gcpiprange.StateFactory,
) *IPRangeReconciler {
	return &IPRangeReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
		scopeStateFactory:    scopeStateFactory,
		awsStateFactory:      awsStateFactory,
		azureStateFactory:    azureStateFactory,
		gcpStateFactory:      gcpStateFactory,
	}
}

func (r *IPRangeReconciler) Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	state := r.newFocalState(req.NamespacedName)
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *IPRangeReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"main",
		focal.New(),
		scope.New(r.scopeStateFactory),
		func(ctx context.Context, st composed.State) (error, context.Context) {
			return composed.ComposeActions(
				"ipRangeCommon",
				// some IpRange common actions here
			)(ctx, newState(st.(focal.State)))
		},
		composed.BuildSwitchAction(
			"providerSwitch",
			nil,
			composed.NewCase(focal.AwsProviderPredicate, awsiprange.New(r.awsStateFactory)),
			composed.NewCase(focal.AzureProviderPredicate, azureiprange.New(r.azureStateFactory)),
			composed.NewCase(focal.GcpProviderPredicate, gcpiprange.New(r.gcpStateFactory)),
		),
	)
}

func (r *IPRangeReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudresourcesv1beta1.IpRange{}),
	)
}
