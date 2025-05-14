package azurerwxvolumebackup

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"github.com/kyma-project/cloud-manager/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
)

type reconciler struct {
	factory *stateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {

	if Ignore != nil && Ignore.ShouldIgnoreKey(request) {
		return ctrl.Result{}, nil
	}

	state := r.factory.NewState(request)

	action := r.newAction()

	return composed.Handling().
		WithMetrics("azurerwxvolumebackup", util.RequestObjToString(request)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"azureRwxVolumeBackupMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.AzureRwxVolumeBackup{}),
		commonscope.New(),

		loadPersistentVolumeClaim,
		loadPersistentVolume,
		createClient,
		createVault,
		getProtectedResourceName,

		composed.IfElse(
			composed.Not(CompletedOrDeletedPredicate),
			composed.ComposeActions("AzureRwxVolumeBackupNotCompletedOrDeleted",
				actions.PatchAddCommonFinalizer(),
				createBackupPolicy,
				protectFileshare,
				createBackup,
			), nil,
		),
		composed.If(
			composed.MarkedForDeletionPredicate,
			composed.ComposeActions("AzureRwxVolumeBackupMarkedForDeletion"),
			actions.PatchRemoveCommonFinalizer(),
			deleteProtectedItem,
		),
		composed.StopAndForgetAction,
	)
}

func NewReconciler(args skrruntime.ReconcilerArguments, clientProvider azureclient.ClientProvider[client.Client]) reconcile.Reconciler {

	return &reconciler{
		factory: newStateFactory(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(args.SkrCluster)),
			commonscope.NewStateFactory(
				composed.NewStateClusterFromCluster(args.KcpCluster),
				args.KymaRef),
			clientProvider,
		),
	}
}

func CompletedOrDeletedPredicate(_ context.Context, state composed.State) bool {
	isDeleted := composed.IsMarkedForDeletion(state.Obj())
	currentState := state.Obj().(*cloudresourcesv1beta1.AzureRwxVolumeBackup).Status.State
	//return isDeleted || currentState == cloudresourcesv1beta1.JobStateDone || currentState == cloudresourcesv1beta1.JobStateFailed
	return isDeleted || currentState == cloudresourcesv1beta1.AzureRwxBackupDone || currentState == cloudresourcesv1beta1.AzureRwxBackupFailed

}
