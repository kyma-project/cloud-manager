package awsnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"github.com/kyma-project/cloud-manager/pkg/skr/common/defaultiprange"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
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
			args.KymaRef,
			composed.NewStateClusterFromCluster(args.KcpCluster),
		),
	}
}

type reconciler struct {
	factory *stateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	if Ignore.ShouldIgnoreKey(request) {
		return ctrl.Result{}, nil
	}

	state := r.factory.NewState(request)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("awsnfsvolume", util.RequestObjToString(request)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crAwsNfsVolumeMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.AwsNfsVolume{}),
		composed.LoadObj,
		composed.ComposeActions(
			"crAwsNfsVolumeValidateSpec",
			validatePersistentVolume, validatePersistentVolumeClaim,
		),
		defaultiprange.New(),

		loadVolume,
		sanitizeReleasedVolume,
		loadPersistentVolumeClaim,
		addFinalizer,
		updateId,
		loadKcpNfsInstance,
		createKcpNfsInstance,
		updateStatus,
		createVolume,
		createPersistentVolumeClaim,
		requeueWaitKcpStatus,
		stopIfNotBeingDeleted,

		// this below executes only when marked for deletion

		removePersistenceVolumeClaimFinalizer,
		deletePVC,
		waitPVCDeleted,

		removePersistenceVolumeFinalizer,
		deletePv,
		waitPvDeleted,

		deleteKcpNfsInstance,
		waitKcpNfsInstanceDeleted,

		removeFinalizer,

		composed.StopAndForgetAction,
	)
}
