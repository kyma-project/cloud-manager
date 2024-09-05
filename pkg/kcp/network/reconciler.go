package network

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	awsnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/network"
	azurenetwork "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/network"
	gcpnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/network"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type NetworkReconciler interface {
	reconcile.Reconciler
}

type networkReconciler struct {
	composedStateFactory composed.StateFactory
	focalStateFactory    focal.StateFactory

	awsStateFactory   awsnetwork.StateFactory
	azureStateFactory azurenetwork.StateFactory
	gcpStateFactory   gcpnetwork.StateFactory
}

func NewNetworkReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
) NetworkReconciler {
	return &networkReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
	}
}

func (r *networkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if Ignore != nil && Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}

	state := r.newFocalState(req.NamespacedName)
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *networkReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"main",
		feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.NfsInstance{}),
		focal.New(),
		func(ctx context.Context, st composed.State) (error, context.Context) {
			return composed.ComposeActions(
				"networkCommon",
				// common Network actions here

				// and now branch to provider specific flow
				composed.BuildSwitchAction(
					"providerSwitch",
					nil,
					composed.NewCase(focal.AwsProviderPredicate, awsnetwork.New(r.awsStateFactory)),
					composed.NewCase(focal.AzureProviderPredicate, azurenetwork.New(r.azureStateFactory)),
					composed.NewCase(focal.GcpProviderPredicate, gcpnetwork.New(r.gcpStateFactory)),
				),
			)(ctx, newState(st.(focal.State)))
		},
	)
}

func (r *networkReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.NfsInstance{}),
	)
}