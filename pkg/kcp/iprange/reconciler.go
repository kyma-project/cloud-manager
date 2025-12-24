package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/common/statewithscope"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	awsiprange "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange"
	azureiprange "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/iprange"
	gcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange"
	sapiprange "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/iprange"
	"github.com/kyma-project/cloud-manager/pkg/util"
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
	gcpV3StateFactory gcpiprange.V3StateFactory // v3 refactored state factory (NEW pattern)
	gcpV2StateFactory gcpiprange.V2StateFactory // v2 legacy state factory
	sapStateFactory   sapiprange.StateFactory
}

func NewIPRangeReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	awsStateFactory awsiprange.StateFactory,
	azureStateFactory azureiprange.StateFactory,
	gcpV3StateFactory gcpiprange.V3StateFactory,
	gcpV2StateFactory gcpiprange.V2StateFactory,
	sapStateFactory sapiprange.StateFactory,
) IPRangeReconciler {
	return &ipRangeReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
		awsStateFactory:      awsStateFactory,
		azureStateFactory:    azureStateFactory,
		gcpV3StateFactory:    gcpV3StateFactory,
		gcpV2StateFactory:    gcpV2StateFactory,
		sapStateFactory:      sapStateFactory,
	}
}

func (r *ipRangeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if Ignore != nil && Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}

	state := r.newFocalState(req.NamespacedName)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("kcpiprange", util.RequestObjToString(req)).
		Handle(action(ctx, state))
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
				actions.PatchAddCommonFinalizer(),
				composed.If(
					shouldAllocateIpRange,
					composed.BuildSwitchAction(
						"allocateIpRangeProviderSwitch",
						nil,
						composed.NewCase(statewithscope.AwsProviderPredicate, awsiprange.NewAllocateIpRangeAction(r.awsStateFactory)),
						composed.NewCase(statewithscope.AzureProviderPredicate, azureiprange.NewAllocateIpRangeAction(r.azureStateFactory)),
						composed.NewCase(statewithscope.GcpProviderPredicate, gcpiprange.NewAllocateIpRangeAction(r.gcpV3StateFactory, r.gcpV2StateFactory)),
						composed.NewCase(statewithscope.OpenStackProviderPredicate, sapiprange.NewAllocateIpRangeAction(r.sapStateFactory)),
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
					kymaPeeringCreate,
					kymaPeeringWait,
				),
				// prevent delete if used before switching to provider specific flow
				// so cloud resources in the provider are not tried to be deleted
				// before dependant objects are first deleted gracefully
				composed.If(
					composed.MarkedForDeletionPredicate,
					preventDeleteOnNfsInstanceUsage,
					preventDeleteOnRedisInstanceUsage,
					preventDeleteOnRedisClusterUsage,
				),
				// and now branch to provider specific flow
				composed.If(
					// call providers only if network is not deleted yet, they have a strong
					// dependency on the KCP IpRange's Network
					// due to kcpNetworkDeleteWait() the requeue will happen when provider
					// finished the deprovisioning and KCP Network is deleted and then
					// waited for (requeued) to not exist anymore
					shouldCallProviderFlow,
					composed.BuildSwitchAction(
						"providerSwitch",
						nil,
						composed.NewCase(statewithscope.AwsProviderPredicate, awsiprange.New(r.awsStateFactory)),
						composed.NewCase(statewithscope.AzureProviderPredicate, azureiprange.New(r.azureStateFactory)),
						composed.NewCase(statewithscope.GcpProviderPredicate, gcpiprange.New(r.gcpV3StateFactory, r.gcpV2StateFactory)),
						composed.NewCase(statewithscope.OpenStackProviderPredicate, sapiprange.New(r.sapStateFactory)),
					),
				),
				// delete
				composed.If(
					composed.MarkedForDeletionPredicate,
					kymaPeeringDelete,
					kymaPeeringDeleteWait,
					kcpNetworkDelete,
					actions.PatchRemoveCommonFinalizer(),
				),
				statusReady,
			)(ctx, newState(st.(focal.State)))
		},
	)
}

func (r *ipRangeReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.IpRange{}),
	)
}

func shouldCallProviderFlow(ctx context.Context, st composed.State) bool {
	state := st.(*State)
	if state.Network() == nil && composed.MarkedForDeletionPredicate(ctx, state) {
		return false
	}
	return true
}
