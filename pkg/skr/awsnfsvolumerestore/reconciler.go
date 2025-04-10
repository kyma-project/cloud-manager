package awsnfsvolumerestore

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumerestore/client"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewReconcilerFactory(
	awsClientProvider awsclient.SkrClientProvider[client.Client],
	env abstractions.Environment,
) skrruntime.ReconcilerFactory {
	return &reconcilerFactory{
		awsClientProvider: awsClientProvider,
		env:               env,
	}
}

type reconcilerFactory struct {
	awsClientProvider awsclient.SkrClientProvider[client.Client]
	env               abstractions.Environment
}

func (f *reconcilerFactory) New(args skrruntime.ReconcilerArguments) reconcile.Reconciler {
	return &reconciler{
		factory: newStateFactory(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(args.SkrCluster)),
			commonscope.NewStateFactory(
				composed.NewStateClusterFromCluster(args.KcpCluster),
				args.KymaRef,
			),
			f.awsClientProvider,
			f.env,
		),
	}
}

type reconciler struct {
	factory *stateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	state := r.factory.NewState(request)
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"AwsNfsVolumeRestoreMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.AwsNfsVolumeRestore{}),
		commonscope.New(),
		composed.IfElse(
			composed.Not(CompletedRestorePredicate),
			composed.ComposeActions("AwsNfsVolumeNotCompleted",
				actions.PatchAddCommonFinalizer(),
				loadSkrAwsNfsVolumeBackup,
				stopIfBackupNotReady,
				loadSkrAwsNfsVolume,
				stopIfVolumeNotReady,
				setIdempotencyToken,
				createAwsClient,
				startAwsRestore,
				checkRestoreJob),
			nil),
		actions.PatchRemoveCommonFinalizer(),
		composed.StopAndForgetAction,
	)
}

func CompletedRestorePredicate(_ context.Context, state composed.State) bool {
	currentState := state.Obj().(*cloudresourcesv1beta1.AwsNfsVolumeRestore).Status.State
	return currentState == cloudresourcesv1beta1.JobStateDone || currentState == cloudresourcesv1beta1.JobStateFailed
}
