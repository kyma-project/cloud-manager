package sapnfsvolumesnapshot

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"github.com/kyma-project/cloud-manager/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Reconciler struct {
	composedStateFactory composed.StateFactory
	stateFactory         StateFactory
}

func (r *Reconciler) Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := composed.LoggerFromCtx(ctx)

	state, err := r.stateFactory.NewState(ctx,
		r.composedStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.SapNfsVolumeSnapshot{}),
	)
	if err != nil {
		logger.Error(err, "Error creating SapNfsVolumeSnapshot state")
		return ctrl.Result{}, fmt.Errorf("error creating SapNfsVolumeSnapshot state: %w", err)
	}

	action := r.newAction()

	return composed.Handling().
		WithMetrics("sapnfsvolumesnapshot", util.RequestObjToString(req)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *Reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crSapNfsVolumeSnapshotMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.SapNfsVolumeSnapshot{}),
		composed.LoadObj,
		composeActions(),
	)
}

func composeActions() composed.Action {
	return composed.ComposeActions(
		"sapNfsVolumeSnapshot",
		loadScope,
		clientCreate,
		shortCircuit,
		markFailed,
		actions.AddCommonFinalizer(),

		snapshotLoad,
		sourceVolumeLoad,

		ttlExpiry,

		composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
			composed.ComposeActions(
				"sapNfsVolumeSnapshot-create",
				idGenerate,
				snapshotCreate,
				snapshotWaitAvailable,
				statusReady,
			),
			composed.ComposeActions(
				"sapNfsVolumeSnapshot-delete",
				statusDeleting,
				snapshotDelete,
				snapshotWaitDeleted,
				actions.RemoveCommonFinalizer(),
			),
		),
		composed.StopAndForgetAction,
	)
}
