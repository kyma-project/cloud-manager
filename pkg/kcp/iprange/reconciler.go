package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	awsiprange "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange"
	azureiprange "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/iprange"
	gcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

type IPRangeReconciler interface {
	reconcile.Reconciler
}

type ipRangeReconciler struct {
	composedStateFactory composed.StateFactory
	focalStateFactory    focal.StateFactory

	awsStateFactory   awsiprange.StateFactory
	azureStateFactory azureiprange.StateFactory
	gcpStateFactory   gcpiprange.StateFactory
}

func NewIPRangeReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	awsStateFactory awsiprange.StateFactory,
	azureStateFactory azureiprange.StateFactory,
	gcpStateFactory gcpiprange.StateFactory,
) IPRangeReconciler {
	return &ipRangeReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
		awsStateFactory:      awsStateFactory,
		azureStateFactory:    azureStateFactory,
		gcpStateFactory:      gcpStateFactory,
	}
}

func (r *ipRangeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if Ignore != nil && Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}

	state := r.newFocalState(req.NamespacedName)
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *ipRangeReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"main",
		feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.IpRange{}),
		focal.New(),
		func(ctx context.Context, st composed.State) (error, context.Context) {
			return composed.ComposeActions(
				"ipRangeCommon",
				// common IpRange common actions here
				composed.If(
					shouldAllocateIpRange,
					composed.BuildSwitchAction(
						"allocateIpRangeProviderSwitch",
						nil,
						composed.NewCase(focal.AwsProviderPredicate, awsiprange.NewAllocateIpRangeAction(r.awsStateFactory)),
						composed.NewCase(focal.AzureProviderPredicate, azureiprange.NewAllocateIpRangeAction(r.azureStateFactory)),
						composed.NewCase(focal.GcpProviderPredicate, gcpiprange.NewAllocateIpRangeAction(r.gcpStateFactory)),
					),
					allocateIpRange,
				),
				kcpNetworkInit,
				kcpNetworkLoad,
				kcpNetworkCreate,
				kcpNetworkWait,
				composed.If(
					shouldPeerWithKymaNetwork,
					kymaNetworkLoad,
					kymaNetworkWait,
					kymaPeeringLoad,
					kymaPeeringWait,
					kymaPeeringCreate,
				),
				// and now branch to provider specific flow
				composed.BuildSwitchAction(
					"providerSwitch",
					nil,
					composed.NewCase(focal.AwsProviderPredicate, awsiprange.New(r.awsStateFactory)),
					composed.NewCase(focal.AzureProviderPredicate, azureiprange.New(r.azureStateFactory)),
					composed.NewCase(focal.GcpProviderPredicate, gcpiprange.New(r.gcpStateFactory)),
				),
			)(ctx, newState(st.(focal.State)))
		},
	)
}

func (r *ipRangeReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.IpRange{}),
	)
}
