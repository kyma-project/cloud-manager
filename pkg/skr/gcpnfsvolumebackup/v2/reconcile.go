package v2

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsbackupclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	"github.com/kyma-project/cloud-manager/pkg/util"
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
	logger := composed.LoggerFromCtx(ctx)

	//Create state object
	state, err := r.newState(ctx, req.NamespacedName)
	if err != nil {
		logger.Error(err, "Error getting the GcpNfsVolumeBackup state object")
	}

	//Create action handler.
	action := r.newAction()

	return composed.Handling().
		WithMetrics("gcpnfsvolumebackup", util.RequestObjToString(req)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *Reconciler) newState(ctx context.Context, name types.NamespacedName) (*State, error) {
	return r.stateFactory.NewState(ctx,
		r.composedStateFactory.NewState(name, &cloudresourcesv1beta1.GcpNfsVolumeBackup{}),
	)
}

func (r *Reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crGcpNfsVolumeBackupMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.GcpNfsVolumeBackup{}),
		composed.LoadObj,
		composeActions(),
	)
}

func NewReconciler(kymaRef klog.ObjectRef, kcpCluster cluster.Cluster, skrCluster cluster.Cluster,
	fileBackupClientProvider gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]) Reconciler {
	compSkrCluster := composed.NewStateClusterFromCluster(skrCluster)
	compKcpCluster := composed.NewStateClusterFromCluster(kcpCluster)
	composedStateFactory := composed.NewStateFactory(compSkrCluster)
	stateFactory := NewStateFactory(kymaRef, compKcpCluster, compSkrCluster, fileBackupClientProvider)
	return Reconciler{
		composedStateFactory: composedStateFactory,
		stateFactory:         stateFactory,
	}
}

func composeActions() composed.Action {
	return composed.ComposeActions(
		"gcpNfsVolumeBackupV2",
		loadScope,
		shortCircuitCompleted,
		markFailed,
		actions.AddCommonFinalizer(),

		loadNfsBackup,
		loadGcpNfsVolume,

		checkBackupOperation,

		composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
			composed.ComposeActions(
				"gcpNfsVolumeBackupV2-create",
				createNfsBackup,
				waitBackupReady,
				addLabelsToNfsBackup,
				updateStatus,
			),
			composed.ComposeActions(
				"gcpNfsVolumeBackupV2-delete",
				deleteNfsBackup,
				waitBackupDeleted,
				actions.RemoveCommonFinalizer(),
			),
		),
		composed.StopAndForgetAction,
	)
}
