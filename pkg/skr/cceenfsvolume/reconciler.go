package cceenfsvolume

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"github.com/kyma-project/cloud-manager/pkg/skr/common/defaultiprange"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewReconcilerFactory() skrruntime.ReconcilerFactory {
	return &reconcilerFactory{}
}

type reconcilerFactory struct{}

func (f *reconcilerFactory) New(args skrruntime.ReconcilerArguments) reconcile.Reconciler {
	return &reconciler{
		factory: newStateFactory(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(args.SkrCluster)),
			args.KymaRef,
			composed.NewStateClusterFromCluster(args.KcpCluster),
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
		"crCceeNfsVolumeMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.CceeNfsVolume{}),
		composed.LoadObj,
		pvValidate,
		pvcValidate,
		defaultiprange.New(),

		pvLoad,
		pvRemoveClaimRef,
		pvcLoad,
		actions.PatchAddFinalizer,
		idGenerate,

		kcpNfsInstanceLoad,
		kcpNfsInstanceCreate,
		statusCopy,
		waitKcpNfsInstanceStatus,

		updateSize,

		pvCreate,
		pvcCreate,

		stopIfNotBeingDeleted,

		// below executes only when marked for deletion

		pvcRemoveFinalizer,
		pvcDelete,
		pvcWaitDeleted,

		pvRemoveFinalizer,
		pvDelete,
		pvWaitDeleted,

		kcpNfsInstanceDelete,
		kcpNfsInstanceWaitDeleted,

		actions.PatchRemoveFinalizer,

		composed.StopAndForgetAction,

		// TODO add more actions here
	)
}
