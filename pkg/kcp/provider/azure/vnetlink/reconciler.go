package vnetlink

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/types"
)

type AzureVNetLinkReconciler interface {
	reconcile.Reconciler
}

type azureVNetLinkReconciler struct {
	composedStateFactory composed.StateFactory
	focalStateFactory    focal.StateFactory
	stateFactory         StateFactory
}

func NewAzureVNetLinkReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	stateFactory StateFactory) AzureVNetLinkReconciler {
	return &azureVNetLinkReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
		stateFactory:         stateFactory,
	}
}

func (r *azureVNetLinkReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	if Ignore != nil && Ignore.ShouldIgnoreKey(request) {
		return ctrl.Result{}, nil
	}

	state := r.newFocalState(request.NamespacedName)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("kcpazurevnetlink", util.RequestObjToString(request)).
		Handle(action(ctx, state))
}

func (r *azureVNetLinkReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"main",
		focal.New(),
		r.newFlow(),
	)
}

func (r *azureVNetLinkReconciler) newFlow() composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state, err := r.stateFactory.NewState(ctx, st.(focal.State))

		if err != nil {
			return composed.LogErrorAndReturn(err, "Failed to bootstrap AzureVNetLink state", composed.StopAndForget, ctx)
		}

		return composed.ComposeActions(
			"azureVNetLink",
			initState,
			initRemoteClient,
			statusInProgress,
			loadVNetLink,
			composed.IfElse(
				composed.MarkedForDeletionPredicate,
				composed.ComposeActions(
					"azureVNetLink-delete",
					deleteVNetLink,
					actions.PatchRemoveCommonFinalizer(),
				),
				composed.ComposeActions(
					"azureVNetLink-non-delete",
					actions.AddCommonFinalizer(),
					composed.If(
						predicateRequireVNetShootTag,
						loadPrivateDnsZone,
						waitPrivateDnsZoneTag,
					),
					createVNetLink,
					waitVNetLinkCompleted,
					updateStatus,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}

func (r *azureVNetLinkReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.AzureVNetLink{}),
	)
}
