package vnetlink

import (
	"context"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vnetlink/dnsresolver"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vnetlink/dnszone"
	"github.com/kyma-project/cloud-manager/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/types"
)

type AzureVNetLinkReconciler interface {
	reconcile.Reconciler
}

type azureVNetLinkReconciler struct {
	composedStateFactory    composed.StateFactory
	focalStateFactory       focal.StateFactory
	dnsZoneStateFactory     dnszone.StateFactory
	dnsResolverStateFactory dnsresolver.StateFactory
}

func NewAzureVNetLinkReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	dnsZoneStateFactory dnszone.StateFactory,
	dnsResolverStateFactory dnsresolver.StateFactory) AzureVNetLinkReconciler {
	return &azureVNetLinkReconciler{
		composedStateFactory:    composedStateFactory,
		focalStateFactory:       focalStateFactory,
		dnsZoneStateFactory:     dnsZoneStateFactory,
		dnsResolverStateFactory: dnsResolverStateFactory,
	}
}

func (r *azureVNetLinkReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	if dnszone.Ignore != nil && dnszone.Ignore.ShouldIgnoreKey(request) {
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
		feature.LoadFeatureContextFromObj(&v1beta1.AzureVNetLink{}),
		focal.New(),
		func(ctx context.Context, st composed.State) (error, context.Context) {
			return composed.ComposeActions(
				"vnetlinkCommon",
				composed.BuildSwitchAction(
					"zoneResolverSwitch",
					nil,
					composed.NewCase(dnsZonePredicate, dnszone.New(r.dnsZoneStateFactory)),
					composed.NewCase(dnsResolverPredicate, dnsresolver.New(r.dnsResolverStateFactory)),
				),
			)(ctx, newState(st.(focal.State)))
		},
	)
}

func (r *azureVNetLinkReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &v1beta1.AzureVNetLink{}),
	)
}

func dnsZonePredicate(_ context.Context, st composed.State) bool {
	if link, ok := st.Obj().(*v1beta1.AzureVNetLink); ok {
		return len(link.Spec.RemotePrivateDnsZone) > 0
	}
	return false
}

func dnsResolverPredicate(_ context.Context, st composed.State) bool {
	if link, ok := st.Obj().(*v1beta1.AzureVNetLink); ok {
		return len(link.Spec.RemoteDnsForwardingRuleset) > 0
	}
	return false

}
