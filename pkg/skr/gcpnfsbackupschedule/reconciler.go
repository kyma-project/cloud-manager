package gcpnfsbackupschedule

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
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
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
		logger.Error(err, "Error getting the GcpNfsBackupSchedule state object")
	}

	action := r.newAction()

	return composed.Handling().
		WithMetrics(strings.ToLower(fmt.Sprintf("%T", state.Obj())), util.RequestObjToString(req)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *Reconciler) newState(ctx context.Context, name types.NamespacedName) (*State, error) {
	return r.stateFactory.NewState(ctx,
		r.composedStateFactory.NewState(name, &cloudresourcesv1beta1.GcpNfsBackupSchedule{}),
	)
}

func (r *Reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"gcpNfsBackupScheduleV2",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.GcpNfsBackupSchedule{}),
		composed.LoadObj,
		actions.AddCommonFinalizer(),
		loadBackups,
		composed.IfElse(
			composed.Not(composed.MarkedForDeletionPredicate),
			// Main (create) path
			composed.ComposeActions(
				"gcpNfsBackupScheduleV2-main",
				// Common scheduling actions
				backupschedule.CheckCompleted,
				backupschedule.CheckSuspension,
				backupschedule.ValidateSchedule,
				backupschedule.ValidateTimes,
				backupschedule.CalculateOnetimeSchedule,
				backupschedule.CalculateRecurringSchedule,
				backupschedule.EvaluateNextRun,
				// GCP-specific actions
				loadScope,
				loadSource,
				createBackup,
				deleteBackups,
				updateStatus,
			),
			// Delete path
			composed.ComposeActions(
				"gcpNfsBackupScheduleV2-delete",
				deleteCascade,
				actions.RemoveCommonFinalizer(),
			),
		),
		composed.StopAndForgetAction,
	)
}

func NewReconciler(kymaRef klog.ObjectRef, kcpCluster cluster.Cluster, skrCluster cluster.Cluster,
	env abstractions.Environment, clk clock.Clock) Reconciler {
	compSkrCluster := composed.NewStateClusterFromCluster(skrCluster)
	compKcpCluster := composed.NewStateClusterFromCluster(kcpCluster)
	composedStateFactory := composed.NewStateFactory(compSkrCluster)
	stateFactory := NewStateFactory(kymaRef, compKcpCluster, compSkrCluster, env, clk)

	return Reconciler{
		composedStateFactory: composedStateFactory,
		stateFactory:         stateFactory,
	}
}
