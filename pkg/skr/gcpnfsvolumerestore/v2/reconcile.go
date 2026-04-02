package v2

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsbackupclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	gcpnfsrestoreclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client/v2"
	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/types"
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
		logger.Error(err, "Error getting the GcpNfsVolumeRestore state object")
	}

	//Create action handler.
	action := r.newAction()

	return composed.Handling().
		WithMetrics("gcpnfsvolumerestore", util.RequestObjToString(req)).
		WithNoLog().
		Handle(action(ctx, state))
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
		composeActions(),
	)
}

func NewReconciler(scopeProvider scopeprovider.ScopeProvider, kcpCluster cluster.Cluster, skrCluster cluster.Cluster,
	fileRestoreClientProvider gcpclient.GcpClientProvider[gcpnfsrestoreclientv2.FileRestoreClient],
	fileBackupClientProvider gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient],
) Reconciler {
	compSkrCluster := composed.NewStateClusterFromCluster(skrCluster)
	compKcpCluster := composed.NewStateClusterFromCluster(kcpCluster)
	composedStateFactory := composed.NewStateFactory(compSkrCluster)
	stateFactory := NewStateFactory(scopeProvider, compKcpCluster, compSkrCluster, fileRestoreClientProvider, fileBackupClientProvider)
	return Reconciler{
		composedStateFactory: composedStateFactory,
		stateFactory:         stateFactory,
	}
}

func composeActions() composed.Action {
	return composed.ComposeActions(
		"gcpNfsVolumeRestoreV2",
		loadScope,
		clientCreate,
		shortCircuitCompleted,
		actions.AddCommonFinalizer(),

		loadGcpNfsVolume,
		populateBackupUrl,
		loadBackup,

		checkRestoreOperation,

		composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
			composed.ComposeActions(
				"gcpNfsVolumeRestoreV2-create",
				setProcessing,
				checkRestorePermissions,
				acquireLease,
				findRestoreOperation,
				runNfsRestore,
				releaseLease,
			),
			composed.ComposeActions(
				"gcpNfsVolumeRestoreV2-delete",
				releaseLease,
				actions.RemoveCommonFinalizer(),
			),
		),
		composed.StopAndForgetAction,
	)
}
