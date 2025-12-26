package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/common/statewithscope"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	awsnfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance"
	azurenfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/nfsinstance"
	gcpnfsinstancev1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v1"
	sapnfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/nfsinstance"
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
	gcpStateFactory   gcpnfsinstancev1.StateFactory
	sapStateFactory   sapnfsinstance.StateFactory
}

func NewNfsInstanceReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	awsStateFactory awsnfsinstance.StateFactory,
	azureStateFactory azurenfsinstance.StateFactory,
	gcpStateFactory gcpnfsinstancev1.StateFactory,
	sapStateFactory sapnfsinstance.StateFactory,
) NfsInstanceReconciler {
	return &nfsInstanceReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
		awsStateFactory:      awsStateFactory,
		azureStateFactory:    azureStateFactory,
		gcpStateFactory:      gcpStateFactory,
		sapStateFactory:      sapStateFactory,
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
					composed.NewCase(statewithscope.AwsProviderPredicate, awsnfsinstance.New(r.awsStateFactory)),
					composed.NewCase(statewithscope.AzureProviderPredicate, azurenfsinstance.New(r.azureStateFactory)),
					composed.NewCase(statewithscope.GcpProviderPredicate, gcpnfsinstancev1.New(r.gcpStateFactory)),
					composed.NewCase(statewithscope.OpenStackProviderPredicate, sapnfsinstance.New(r.sapStateFactory)),
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
