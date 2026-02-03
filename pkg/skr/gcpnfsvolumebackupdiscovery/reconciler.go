package gcpnfsvolumebackupdiscovery

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsbackupclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client"
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

	state, err := r.newState(ctx, req.NamespacedName)
	if err != nil {
		logger.Error(err, "Error getting the GcpNfsVolumeBackupDiscovery state object")
	}

	action := r.newAction()

	return composed.Handling().
		WithMetrics("gcpnfsvolumebackupdiscovery", util.RequestObjToString(req)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *Reconciler) newState(ctx context.Context, name types.NamespacedName) (*State, error) {
	return r.stateFactory.NewState(ctx,
		r.composedStateFactory.NewState(name, &cloudresourcesv1beta1.GcpNfsVolumeBackupDiscovery{}),
	)
}

func (r *Reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crGcpNfsVolumeBackupDiscoveryMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.GcpNfsVolumeBackupDiscovery{}),
		composed.LoadObj,
		setProcessing,
		shortCircuit,
		loadScope,
		loadAvailableBackups,
		updateStatus,
		composed.StopAndForgetAction,
	)
}

func NewReconciler(kymaRef klog.ObjectRef, kcpCluster cluster.Cluster, skrCluster cluster.Cluster,
	fileBackupClientProvider gcpclient.ClientProvider[gcpnfsbackupclient.FileBackupClient], env abstractions.Environment) Reconciler {
	compSkrCluster := composed.NewStateCluster(skrCluster.GetClient(), skrCluster.GetAPIReader(), skrCluster.GetEventRecorderFor("cloud-resources"), skrCluster.GetScheme()) //nolint:staticcheck // SA1019
	compKcpCluster := composed.NewStateCluster(kcpCluster.GetClient(), kcpCluster.GetAPIReader(), kcpCluster.GetEventRecorderFor("cloud-control"), kcpCluster.GetScheme())   //nolint:staticcheck // SA1019
	composedStateFactory := composed.NewStateFactory(compSkrCluster)
	stateFactory := NewStateFactory(kymaRef, compKcpCluster, compSkrCluster, fileBackupClientProvider, env)
	return Reconciler{
		composedStateFactory: composedStateFactory,
		stateFactory:         stateFactory,
	}
}
