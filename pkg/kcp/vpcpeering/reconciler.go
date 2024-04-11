package vpcpeering

import (
	"context"
	cloudcontrolb1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcpeering"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type VPCPeeringReconciler interface {
	reconcile.Reconciler
}

type vpcPeeringReconciler struct {
	composedStateFactory composed.StateFactory
	focalStateFactory    focal.StateFactory

	awsStateFactory vpcpeering.StateFactory
}

func NewVpcPeeringReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	awsStateFactory vpcpeering.StateFactory,
) VPCPeeringReconciler {
	return &vpcPeeringReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
		awsStateFactory:      awsStateFactory,
	}
}

func (r *vpcPeeringReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	state := r.newFocalState(request.NamespacedName)
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *vpcPeeringReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudcontrolb1beta1.VpcPeering{}),
	)
}

func (r *vpcPeeringReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"main",
		focal.New(),
		func(ctx context.Context, st composed.State) (error, context.Context) {
			return composed.ComposeActions(
				"vpcPeeringCommon",
				composed.BuildSwitchAction(
					"providerSwitch",
					nil,
					composed.NewCase(focal.AwsProviderPredicate, vpcpeering.New(r.awsStateFactory)),
				),
			)(ctx, newState(st.(focal.State)))
		},
	)
}
