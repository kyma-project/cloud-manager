package azurerwxvolumerestore

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	commonScope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type reconcilerFactory struct {
	storageClientProvider azureclient.ClientProvider[client.Client]
}

type reconciler struct {
	factory *stateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	state := r.factory.NewState(request)
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

// TODO: fill out the rest of actions
func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"azureRwxVolumeRestoreMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.AwsNfsVolumeRestore{}),
		commonScope.New(),
		composed.IfElse(
			composed.Not(CompletedOrDeletedRestorePredicate),
			composed.ComposeActions("AzureRwxVolumeNotCompletedOrDeleted",
				actions.PatchAddCommonFinalizer(),
				loadAzureRwxVolumeBackup,
				loadPersistentVolumeClaim,
				loadPersistentVolume,
				createAzureStorageClient,
				setProcessing,
				findAzureRestoreJob,
				startAzureRestore,
				checkRestoreJob,
			),
			nil),
		actions.PatchRemoveCommonFinalizer(),
		composed.StopAndForgetAction,
	)
}

func NewReconcilerFactory(clientProvider azureclient.ClientProvider[client.Client]) skrruntime.ReconcilerFactory {
	return &reconcilerFactory{storageClientProvider: clientProvider}
}

func (f *reconcilerFactory) New(args skrruntime.ReconcilerArguments) reconcile.Reconciler {
	return &reconciler{
		factory: newStateFactory(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(args.SkrCluster)),
			args.KymaRef,
			composed.NewStateClusterFromCluster(args.KcpCluster),
			f.storageClientProvider,
		),
	}
}

func CompletedOrDeletedRestorePredicate(_ context.Context, state composed.State) bool {
	isDeleted := composed.IsMarkedForDeletion(state.Obj())
	currentState := state.Obj().(*cloudresourcesv1beta1.AzureRwxVolumeRestore).Status.State
	return isDeleted || currentState == cloudresourcesv1beta1.JobStateDone || currentState == cloudresourcesv1beta1.JobStateFailed
}
