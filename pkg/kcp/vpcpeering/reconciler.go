package vpcpeering

import (
	"context"
	cloudcontrolb1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	aws "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcpeering"
	azure "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering"
	gcp "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type VPCPeeringReconciler interface {
	reconcile.Reconciler
}

type vpcPeeringReconciler struct {
	composedStateFactory composed.StateFactory
	focalStateFactory    focal.StateFactory

	awsStateFactory   aws.StateFactory
	azureStateFactory azure.StateFactory
	gcpStateFactory   gcp.StateFactory
}

func NewVpcPeeringReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	awsStateFactory aws.StateFactory,
	azureStateFactory azure.StateFactory,
	gcpStateFactory gcp.StateFactory,
) VPCPeeringReconciler {
	return &vpcPeeringReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
		awsStateFactory:      awsStateFactory,
		azureStateFactory:    azureStateFactory,
		gcpStateFactory:      gcpStateFactory,
	}
}

func (r *vpcPeeringReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	if Ignore != nil && Ignore.ShouldIgnoreKey(request) {
		return ctrl.Result{}, nil
	}

	state := r.newFocalState(request.NamespacedName)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("vpcpeering", util.RequestObjToString(request)).
		Handle(action(ctx, state))
}

func (r *vpcPeeringReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudcontrolb1beta1.VpcPeering{}),
	)
}

func (r *vpcPeeringReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"main",
		feature.LoadFeatureContextFromObj(&cloudcontrolb1beta1.VpcPeering{}),
		focal.New(),
		func(ctx context.Context, st composed.State) (error, context.Context) {
			return composed.ComposeActions(
				"vpcPeeringCommon",
				composed.BuildSwitchAction(
					"providerSwitch",
					nil,
					composed.NewCase(focal.AwsProviderPredicate, aws.New(r.awsStateFactory)),
					composed.NewCase(focal.AzureProviderPredicate, azure.New(r.azureStateFactory)),
					composed.NewCase(focal.GcpProviderPredicate, gcp.New(r.gcpStateFactory)),
				),
			)(ctx, newState(st.(focal.State)))
		},
	)
}
