package sapnfssnapshotschedule

import (
	"context"
	"fmt"
	"strings"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"github.com/kyma-project/cloud-manager/pkg/skr/backupschedule"
	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

type Reconciler struct {
	composedStateFactory composed.StateFactory
	stateFactory         StateFactory
}

func (r *Reconciler) Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := composed.LoggerFromCtx(ctx)

	state, err := r.newState(ctx, req.NamespacedName)
	if err != nil {
		logger.Error(err, "Error getting the SapNfsVolumeSnapshotSchedule state object")
	}

	action := r.newAction()

	return composed.Handling().
		WithMetrics(strings.ToLower(fmt.Sprintf("%T", state.Obj())), util.RequestObjToString(req)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *Reconciler) newState(ctx context.Context, name types.NamespacedName) (*State, error) {
	return r.stateFactory.NewState(ctx,
		r.composedStateFactory.NewState(name, &cloudresourcesv1beta1.SapNfsVolumeSnapshotSchedule{}),
	)
}

func (r *Reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"sapNfsVolumeSnapshotSchedule",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.SapNfsVolumeSnapshotSchedule{}),
		composed.LoadObj,
		actions.AddCommonFinalizer(),
		loadSnapshots,
		composed.IfElse(
			composed.Not(composed.MarkedForDeletionPredicate),
			// Main (create) path
			composed.ComposeActions(
				"sapNfsVolumeSnapshotSchedule-main",
				// Common scheduling actions
				backupschedule.CheckCompleted,
				backupschedule.CheckSuspension,
				backupschedule.ValidateSchedule,
				backupschedule.ValidateTimes,
				backupschedule.CalculateOnetimeSchedule,
				backupschedule.CalculateRecurringSchedule,
				backupschedule.EvaluateNextRun,
				// SAP-specific actions
				loadScope,
				loadSource,
				createSnapshot,
				deleteSnapshots,
				setStatusToActive,
			),
			// Delete path
			composed.ComposeActions(
				"sapNfsVolumeSnapshotSchedule-delete",
				deleteCascade,
				actions.RemoveCommonFinalizer(),
			),
		),
		composed.StopAndForgetAction,
	)
}

func NewReconciler(scopeProvider scopeprovider.ScopeProvider, kcpCluster cluster.Cluster, skrCluster cluster.Cluster,
	env abstractions.Environment, clk clock.Clock) Reconciler {
	compSkrCluster := composed.NewStateClusterFromCluster(skrCluster)
	compKcpCluster := composed.NewStateClusterFromCluster(kcpCluster)
	composedStateFactory := composed.NewStateFactory(compSkrCluster)
	stateFactory := NewStateFactory(scopeProvider, compKcpCluster, compSkrCluster, env, clk)

	return Reconciler{
		composedStateFactory: composedStateFactory,
		stateFactory:         stateFactory,
	}
}
