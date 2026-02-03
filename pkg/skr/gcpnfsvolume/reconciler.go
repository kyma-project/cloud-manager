package gcpnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsbackupclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client"
	"github.com/kyma-project/cloud-manager/pkg/util"

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
	logger := composed.LoggerFromCtx(ctx)
	if Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}
	state, err := r.newState(ctx, req.NamespacedName)
	if err != nil {
		logger.Error(err, "Error getting the GcpNfsVolumeRestore state object")
	}
	action := r.newAction()

	return composed.Handling().
		WithMetrics("gcpnfsvolume", util.RequestObjToString(req)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *Reconciler) newState(ctx context.Context, name types.NamespacedName) (*State, error) {
	return r.stateFactory.NewState(
		ctx,
		r.composedStateFactory.NewState(name, &cloudresourcesv1beta1.GcpNfsVolume{}),
	)
}

func (r *Reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crGcpNfsVolumeMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.GcpNfsVolume{}),
		composed.LoadObj,
		loadScope,
		composed.ComposeActions(
			"crGcpNfsVolumeValidateSpec",
			validateIpRange, validatePV, validatePVC),
		setProcessing,
		defaultiprange.New(),
		addFinalizer,
		loadKcpIpRange,
		loadKcpNfsInstance,
		updateStatusId,
		composed.IfElse(
			composed.All(composed.Not(composed.MarkedForDeletionPredicate), SourceBackupPredicate(), NoKcpNfsInstancePredicate()),
			composed.ComposeActions(
				"restoreFromSourceBackup",
				loadScope,
				loadBackup,
				checkRestorePermissions,
				populateBackupUrl),
			nil,
		),
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

func NewReconciler(
	kymaRef klog.ObjectRef,
	kcpCluster cluster.Cluster,
	skrCluster cluster.Cluster,
	fileBackupClientProvider gcpclient.ClientProvider[gcpnfsbackupclient.FileBackupClient],
	env abstractions.Environment,
) Reconciler {
	compSkrCluster := composed.NewStateCluster(skrCluster.GetClient(), skrCluster.GetAPIReader(), skrCluster.GetEventRecorderFor("cloud-resources"), skrCluster.GetScheme()) //nolint:staticcheck // SA1019
	compKcpCluster := composed.NewStateCluster(kcpCluster.GetClient(), kcpCluster.GetAPIReader(), kcpCluster.GetEventRecorderFor("cloud-control"), kcpCluster.GetScheme())   //nolint:staticcheck // SA1019
	composedStateFactory := composed.NewStateFactory(compSkrCluster)
	stateFactory := NewStateFactory(kymaRef, compKcpCluster, compSkrCluster, fileBackupClientProvider, env)
	return Reconciler{
		composedStateFactory: composedStateFactory,
		stateFactory:         stateFactory,
	}
}

func SourceBackupPredicate() composed.Predicate {
	return func(ctx context.Context, state composed.State) bool {
		return len(state.Obj().(*cloudresourcesv1beta1.GcpNfsVolume).Spec.SourceBackup.Name) > 0 || len(state.Obj().(*cloudresourcesv1beta1.GcpNfsVolume).Spec.SourceBackupUrl) > 0
	}
}

func NoKcpNfsInstancePredicate() composed.Predicate {
	return func(ctx context.Context, state composed.State) bool {
		//If KcpNfsInstance is not null, it means that NfsInstance object has already been created on KCP and we don't need backup to populate the field.
		//This allows backup to be deleted after NfsInstance is created.
		return state.(*State).KcpNfsInstance == nil
	}
}
