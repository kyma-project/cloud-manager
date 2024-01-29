package nfsinstance

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
	awsnfsinstance "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/nfsinstance"
	azurenfsinstance "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/azure/nfsinstance"
	gcpnfsinstance "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/nfsinstance"
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
}

func NewNfsInstanceReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	awsStateFactory awsnfsinstance.StateFactory,
	azureStateFactory azurenfsinstance.StateFactory,
	gcpStateFactory gcpnfsinstance.StateFactory,
) NfsInstanceReconciler {
	return &nfsInstanceReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
		awsStateFactory:      awsStateFactory,
		azureStateFactory:    azureStateFactory,
		gcpStateFactory:      gcpStateFactory,
	}
}

func (r *nfsInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	state := r.newFocalState(req.NamespacedName)
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *nfsInstanceReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"main",
		focal.New(),
		func(ctx context.Context, st composed.State) (error, context.Context) {
			return composed.ComposeActions(
				"nfsInstanceCommon",
				// common NfsInstance common actions here
				loadIpRange,
				// and now branch to provider specific flow
				composed.BuildSwitchAction(
					"providerSwitch",
					nil,
					composed.NewCase(focal.AwsProviderPredicate, awsnfsinstance.New(r.awsStateFactory)),
					composed.NewCase(focal.AzureProviderPredicate, azurenfsinstance.New(r.azureStateFactory)),
					composed.NewCase(focal.GcpProviderPredicate, gcpnfsinstance.New(r.gcpStateFactory)),
				),
			)(ctx, newState(st.(focal.State)))
		},
	)
}

func (r *nfsInstanceReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudresourcesv1beta1.NfsInstance{}),
	)
}
