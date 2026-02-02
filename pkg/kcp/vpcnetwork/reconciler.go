package vpcnetwork

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	kcpcommonaction "github.com/kyma-project/cloud-manager/pkg/kcp/commonAction"
	awsvpcnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcnetwork"
	azurevpcnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcnetwork"
	sapvpcnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/vpcnetwork"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type VpcNetworkReconciler interface {
	reconcile.Reconciler
}

func New(
	composedStateFactory composed.StateFactory,
	kcpCommonStateFactory kcpcommonaction.StateFactory,
	awsStateFactory awsvpcnetwork.StateFactory,
	azureStateFactory azurevpcnetwork.StateFactory,
	sapStateFactory sapvpcnetwork.StateFactory,
) VpcNetworkReconciler {
	return &vpcNetworkReconciler{
		composedStateFactory:  composedStateFactory,
		kcpCommonStateFactory: kcpCommonStateFactory,
		awsStateFactory:       awsStateFactory,
		azureStateFactory:     azureStateFactory,
		sapStateFactory:       sapStateFactory,
	}
}

type vpcNetworkReconciler struct {
	composedStateFactory  composed.StateFactory
	kcpCommonStateFactory kcpcommonaction.StateFactory

	awsStateFactory   awsvpcnetwork.StateFactory
	azureStateFactory azurevpcnetwork.StateFactory
	sapStateFactory   sapvpcnetwork.StateFactory
}

func (r *vpcNetworkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if Ignore != nil && Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}

	kcpCommonState := r.newKcpCommonState(req.NamespacedName)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("vpcnetwork", util.RequestObjToString(req)).
		Handle(action(ctx, kcpCommonState))
}

func (r *vpcNetworkReconciler) newKcpCommonState(name types.NamespacedName) kcpcommonaction.State {
	return r.kcpCommonStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.VpcNetwork{}),
	)
}

func (r *vpcNetworkReconciler) newAction() composed.Action {
	providerFlow := composed.Switch(
		nil,
		composed.NewCase(kcpcommonaction.AwsProviderPredicate, awsvpcnetwork.New(r.awsStateFactory)),
		composed.NewCase(kcpcommonaction.AzureProviderPredicate, azurevpcnetwork.New(r.azureStateFactory)),
		composed.NewCase(kcpcommonaction.GcpProviderPredicate, composed.Noop),
		composed.NewCase(kcpcommonaction.OpenStackProviderPredicate, sapvpcnetwork.New(r.sapStateFactory)),
	)

	return composed.ComposeActionsNoName(
		feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.VpcNetwork{}),
		kcpcommonaction.New(),
		func(ctx context.Context, st composed.State) (error, context.Context) {
			return composed.ComposeActionsNoName(
				// vpc network actions
				composed.Noop,
				composed.IfElse(
					composed.MarkedForDeletionPredicate,
					composed.ComposeActionsNoName(
						// delete
						providerFlow,
						actions.PatchRemoveCommonFinalizer(),
					),
					composed.ComposeActionsNoName(
						// create
						actions.PatchAddCommonFinalizer(),
						specCidrBlocksValidate,
						providerFlow,
						statusReady,
					),
				),
			)(ctx, newState(st.(kcpcommonaction.State)))
		},
	)
}
