package sapnfsvolumesnapshotrestore

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
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
		r.composedStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.SapNfsVolumeSnapshotRestore{}),
	)
	if err != nil {
		logger.Error(err, "Error creating SapNfsVolumeSnapshotRestore state")
		return ctrl.Result{}, fmt.Errorf("error creating SapNfsVolumeSnapshotRestore state: %w", err)
	}

	action := r.newAction()

	return composed.Handling().
		WithMetrics("sapnfsvolumesnapshotrestore", util.RequestObjToString(req)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *Reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crSapNfsVolumeSnapshotRestoreMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.SapNfsVolumeSnapshotRestore{}),
		composed.LoadObj,
		composeActions(),
	)
}

func composeActions() composed.Action {
	return composed.ComposeActions(
		"sapNfsVolumeSnapshotRestore",
		loadScope,
		clientCreate,
		shortCircuitCompleted,
		actions.AddCommonFinalizer(),

		sourceSnapshotLoad,

		composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
			composed.ComposeActions(
				"sapNfsVolumeSnapshotRestore-create",
				statusInProgress,
				composed.IfElse(isInPlaceRestore,
					composed.ComposeActions(
						"sapNfsVolumeSnapshotRestore-inPlace",
						destinationVolumeLoad,
						acquireLease,
						validateInPlace,
						restoreInPlace,
						restoreInPlaceWait,
						releaseLease,
					),
					composed.ComposeActions(
						"sapNfsVolumeSnapshotRestore-newVolume",
						restoreNewVolume,
						restoreNewVolumeWait,
						statusDone,
					),
				),
			),
			composed.ComposeActions(
				"sapNfsVolumeSnapshotRestore-delete",
				releaseLease,
				actions.RemoveCommonFinalizer(),
			),
		),
		composed.StopAndForgetAction,
	)
}

func NewReconcilerFactory() *ReconcilerFactory {
	return &ReconcilerFactory{}
}

type ReconcilerFactory struct{}

func (f *ReconcilerFactory) New(
	scopeProvider scopeprovider.ScopeProvider,
	kcpCluster composed.StateCluster,
	skrCluster composed.StateCluster,
	provider sapclient.SapClientProvider[sapclient.SnapshotClient],
) *Reconciler {
	composedStateFactory := composed.NewStateFactory(skrCluster)
	stateFactory := NewStateFactory(scopeProvider, kcpCluster, skrCluster, provider)
	return &Reconciler{
		composedStateFactory: composedStateFactory,
		stateFactory:         stateFactory,
	}
}
