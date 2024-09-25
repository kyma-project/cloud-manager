package gcpnfsvolumerestore

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	restoreclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client"
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
	logger := composed.LoggerFromCtx(ctx)

	//Create state object
	state, err := r.newState(ctx, req.NamespacedName)
	if err != nil {
		logger.Error(err, "Error getting the GcpNfsVolumeRestore state object")
	}

	//Create action handler.
	action := r.newAction()

	//Handle request
	return composed.Handle(action(ctx, state))
}

func (r *Reconciler) newState(ctx context.Context, name types.NamespacedName) (*State, error) {
	return r.stateFactory.NewState(ctx,
		r.composedStateFactory.NewState(name, &cloudresourcesv1beta1.GcpNfsVolumeRestore{}),
	)
}

func (r *Reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crGcpNfsVolumeRestoreMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.GcpNfsVolumeRestore{}),
		composed.LoadObj,
		composed.IfElse(
			composed.Not(CompletedRestorePredicate),
			composed.ComposeActions("crGcpNfsVolumeNotCompleted",
				setProcessing,
				addFinalizer,
				loadGcpNfsVolume,
				loadGcpNfsVolumeBackup,
				loadScope,
				acquireLease,
				findRestoreOperation,
				runNfsRestore,
				checkRestoreOperation,
			),
			nil),
		releaseLease,
		removeFinalizer,
		composed.StopAndForgetAction,
	)
}

func NewReconciler(kymaRef klog.ObjectRef, kcpCluster cluster.Cluster, skrCluster cluster.Cluster,
	fileRestoreClientProvider gcpclient.ClientProvider[restoreclient.FileRestoreClient], env abstractions.Environment) Reconciler {
	compSkrCluster := composed.NewStateCluster(skrCluster.GetClient(), skrCluster.GetAPIReader(), skrCluster.GetEventRecorderFor("cloud-resources"), skrCluster.GetScheme())
	compKcpCluster := composed.NewStateCluster(kcpCluster.GetClient(), kcpCluster.GetAPIReader(), kcpCluster.GetEventRecorderFor("cloud-control"), kcpCluster.GetScheme())
	composedStateFactory := composed.NewStateFactory(compSkrCluster)
	stateFactory := NewStateFactory(kymaRef, compKcpCluster, compSkrCluster, fileRestoreClientProvider, env)
	return Reconciler{
		composedStateFactory: composedStateFactory,
		stateFactory:         stateFactory,
	}
}

func CompletedRestorePredicate(_ context.Context, state composed.State) bool {
	currentState := state.Obj().(*cloudresourcesv1beta1.GcpNfsVolumeRestore).Status.State
	return currentState == cloudresourcesv1beta1.JobStateDone || currentState == cloudresourcesv1beta1.JobStateFailed
}
