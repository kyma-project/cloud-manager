package azurerwxpv

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"github.com/kyma-project/cloud-manager/pkg/skr/azurerwxpv/client"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
)

type reconciler struct {
	factory *stateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {

	state := r.factory.NewState(request)

	action := r.newAction()

	return composed.Handling().
		WithMetrics("azurerwxpv", util.RequestObjToString(request)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"azureRwxPVMain",
		feature.LoadFeatureContextFromObj(&corev1.PersistentVolume{}),
		composed.LoadObj,
		composed.If(
			composed.All(
				feature.FFNukeBackupsAzure.Predicate(),
				AzureRwxPvPredicate(),
			),

			composed.ComposeActions("AzureRwxPvDeleted",
				waitBeforeDelete,
				loadScope,
				createAzureClient,
				loadAzureFileShare,
				loadAzureRecoveryVaults,
				loadAzureProtectedItem,
				stopAzureProtection,
				deleteAzureFileShare,
			),
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

func AzureRwxPvPredicate() composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		state := st.(*State)
		pv := state.ObjAsPV()
		value := pv.Spec.CSI != nil && pv.Spec.CSI.Driver == "file.csi.azure.com" &&
			pv.Spec.PersistentVolumeReclaimPolicy == "Delete" && pv.Status.Phase == corev1.VolumeReleased
		if value {
			composed.LoggerFromCtx(ctx).Info("Reconciling Azure PV", "name", pv.Name)
		}
		return value
	}
}
