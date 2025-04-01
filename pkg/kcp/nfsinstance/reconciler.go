package nfsinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	awsnfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance"
	azurenfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/nfsinstance"
	cceenfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/nfsinstance"
	gcpnfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

type NfsInstanceReconciler interface {
	reconcile.Reconciler
}

type nfsInstanceReconciler struct {
	composedStateFactory composed.StateFactory
	focalStateFactory    focal.StateFactory

	awsStateFactory   awsnfsinstance.StateFactory
	azureStateFactory azurenfsinstance.StateFactory
	gcpStateFactory   gcpnfsinstance.StateFactory
	cceeStateFactory  cceenfsinstance.StateFactory
}

func NewNfsInstanceReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	awsStateFactory awsnfsinstance.StateFactory,
	azureStateFactory azurenfsinstance.StateFactory,
	gcpStateFactory gcpnfsinstance.StateFactory,
	cceeStateFactory cceenfsinstance.StateFactory,
) NfsInstanceReconciler {
	return &nfsInstanceReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
		awsStateFactory:      awsStateFactory,
		azureStateFactory:    azureStateFactory,
		gcpStateFactory:      gcpStateFactory,
		cceeStateFactory:     cceeStateFactory,
	}
}

func (r *nfsInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if Ignore != nil && Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}

	state := r.newFocalState(req.NamespacedName)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("nfsinstance", util.RequestObjToString(req)).
		Handle(action(ctx, state))
}

func (r *nfsInstanceReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"main",
		feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.NfsInstance{}),
		focal.New(),
		func(ctx context.Context, st composed.State) (error, context.Context) {
			return composed.ComposeActions(
				"nfsInstanceCommon",
				// common NfsInstance common actions here
				loadIpRange,
				copyStatusHostsToHost,
				// and now branch to provider specific flow
				composed.BuildSwitchAction(
					"providerSwitch",
					nil,
					composed.NewCase(focal.AwsProviderPredicate, awsnfsinstance.New(r.awsStateFactory)),
					composed.NewCase(focal.AzureProviderPredicate, azurenfsinstance.New(r.azureStateFactory)),
					composed.NewCase(focal.GcpProviderPredicate, gcpnfsinstance.New(r.gcpStateFactory)),
					composed.NewCase(focal.OpenStackProviderPredicate, cceenfsinstance.New(r.cceeStateFactory)),
				),
			)(ctx, newState(st.(focal.State)))
		},
	)
}

func (r *nfsInstanceReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.NfsInstance{}),
	)
}
