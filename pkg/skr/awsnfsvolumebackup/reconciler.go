package awsnfsvolumebackup

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsClient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	backupclient "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
	commonScope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewReconcilerFactory(
	awsClientProvider awsClient.SkrClientProvider[backupclient.Client],
	env abstractions.Environment,
) skrruntime.ReconcilerFactory {
	return &reconcilerFactory{
		awsClientProvider: awsClientProvider,
		env:               env,
	}
}

type reconcilerFactory struct {
	awsClientProvider awsClient.SkrClientProvider[backupclient.Client]
	env               abstractions.Environment
}

func (f *reconcilerFactory) New(args skrruntime.ReconcilerArguments) reconcile.Reconciler {
	return &reconciler{
		factory: newStateFactory(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(args.SkrCluster)),
			commonScope.NewStateFactory(
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
		"AwsNfsVolumeBackupMain",
		commonScope.New(),
		loadSkrAwsNfsVolume,
		stopIfVolumeNotReady,
		loadKcpAwsNfsInstance,
		createAwsClient,
		loadVault,

		composed.StopAndForgetAction,
	)
}
