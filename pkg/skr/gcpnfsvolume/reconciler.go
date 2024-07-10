package gcpnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/skr/common/defaultiprange"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

type Reconciler struct {
	composedStateFactory composed.StateFactory
	stateFactory         StateFactory
}

func (r *Reconciler) Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}
	state := r.newState(req.NamespacedName)
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *Reconciler) newState(name types.NamespacedName) *State {
	return r.stateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudresourcesv1beta1.GcpNfsVolume{}),
	)
}

func (r *Reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crGcpNfsVolumeMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.GcpNfsVolume{}),
		composed.LoadObj,
		composed.IfElse(EmptyLocationPredicate(), loadScope, nil),
		composed.ComposeActions(
			"crGcpNfsVolumeValidateSpec",
			validateIpRange, validateFileShareName, validateCapacity, validatePV, validatePVC, validateLocation),
		setProcessing,
		defaultiprange.New(),
		addFinalizer,
		loadKcpIpRange,
		loadKcpNfsInstance,
		loadPersistenceVolume,
		sanitizeReleasedVolume,
		loadPersistentVolumeClaim,
		modifyKcpNfsInstance,
		removePersistenceVolumeClaimFinalizer,
		removePersistenceVolumeFinalizer,
		deletePersistentVolumeClaim,
		deletePVForNameChange,
		deletePersistenceVolume,
		deleteKcpNfsInstance,
		removeFinalizer,
		createPersistenceVolume,
		modifyPersistenceVolume,
		createPersistentVolumeClaim,
		modifyPersistentVolumeClaim,
		updateStatus,
		composed.StopAndForgetAction,
	)
}

func NewReconciler(kymaRef klog.ObjectRef, kcpCluster cluster.Cluster, skrCluster cluster.Cluster) Reconciler {
	compSkrCluster := composed.NewStateCluster(skrCluster.GetClient(), skrCluster.GetAPIReader(), skrCluster.GetEventRecorderFor("cloud-resources"), skrCluster.GetScheme())
	compKcpCluster := composed.NewStateCluster(kcpCluster.GetClient(), kcpCluster.GetAPIReader(), kcpCluster.GetEventRecorderFor("cloud-control"), kcpCluster.GetScheme())
	composedStateFactory := composed.NewStateFactory(compSkrCluster)
	stateFactory := NewStateFactory(kymaRef, compKcpCluster, compSkrCluster)
	return Reconciler{
		composedStateFactory: composedStateFactory,
		stateFactory:         stateFactory,
	}
}

func EmptyLocationPredicate() composed.Predicate {
	return func(ctx context.Context, state composed.State) bool {
		return len(state.Obj().(*cloudresourcesv1beta1.GcpNfsVolume).Spec.Location) == 0
	}
}
