package iprange

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/util"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewReconcilerFactory() skrruntime.ReconcilerFactory {
	return &reconcilerFactory{}
}

type reconcilerFactory struct {
}

func (f *reconcilerFactory) New(args skrruntime.ReconcilerArguments) reconcile.Reconciler {
	return &reconciler{
		factory: newStateFactory(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(args.SkrCluster)),
			args.ScopeProvider,
			composed.NewStateClusterFromCluster(args.KcpCluster),
			args.Provider,
		),
	}
}

type reconciler struct {
	factory *stateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}

	state, err := r.factory.NewState(ctx, req)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating IpRange state: %w", err)
	}
	action := r.newAction()

	return composed.Handling().
		WithMetrics("skriprange", util.RequestObjToString(req)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crIpRangeMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.IpRange{}),
		composed.LoadObj,
		updateId,
		preventCidrChange,
		validateCidr,
		preventCidrOverlap,
		removeOverlapCondition,
		loadKcpIpRange,
		checkQuota,
		addFinalizer,
		createKcpIpRange,
		setProcessingStateForDeletion,
		preventDeleteOnAwsNfsVolumeUsage,
		preventDeleteOnGcpNfsVolumeUsage,
		preventDeleteOnAzureRedisInstanceUsage,
		preventDeleteOnAwsRedisInstanceUsage,
		preventDeleteOnGcpRedisInstanceUsage,
		preventDeleteOnAwsRedisClusterUsage,
		deleteKcpIpRange,
		removeFinalizer,
		updateStatus,
		composed.StopAndForgetAction,
	)
}
