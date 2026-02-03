package backupschedule

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

type Reconciler struct {
	composedStateFactory composed.StateFactory
	stateFactory         StateFactory
	backupImpl           backupImpl
}

func (r *Reconciler) Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := composed.LoggerFromCtx(ctx)

	//Create state object
	state, err := r.newState(ctx, req.NamespacedName)
	if err != nil {
		logger.Error(err, "Error getting the GcpNfsBackupSchedule state object")
	}

	//Create action handler.
	action := r.newAction()

	return composed.Handling().
		WithMetrics(strings.ToLower(fmt.Sprintf("%T", state.Obj())), util.RequestObjToString(req)).
		WithNoLog().
		Handle(action(ctx, state))
}

func (r *Reconciler) newState(ctx context.Context, name types.NamespacedName) (*State, error) {
	if r.backupImpl == nil {
		return nil, errors.New("unknown backup schedule type, no implementation available")
	}
	state, err := r.stateFactory.NewState(ctx,
		r.composedStateFactory.NewState(name, r.backupImpl.emptyScheduleObject()),
	)
	if err == nil {
		state.backupImpl = r.backupImpl
	}
	return state, err
}

func (r *Reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"BackupScheduleMain",
		feature.LoadFeatureContextFromObj(r.backupImpl.emptyScheduleObject()),
		composed.LoadObj,
		addFinalizer,
		checkCompleted,
		checkSuspension,
		validateSchedule,
		validateTimes,
		calculateOnetimeSchedule,
		calculateRecurringSchedule,
		evaluateNextRun,
		loadScope,
		loadSource,
		loadBackups,
		createBackup,
		deleteBackups,
		deleteCascade,
		removeFinalizer,
		composed.StopAndForgetAction,
	)
}

func NewReconciler(kymaRef klog.ObjectRef, kcpCluster cluster.Cluster, skrCluster cluster.Cluster,
	env abstractions.Environment, scheduleType ScheduleType) Reconciler {
	compSkrCluster := composed.NewStateCluster(skrCluster.GetClient(), skrCluster.GetAPIReader(), skrCluster.GetEventRecorderFor("cloud-resources"), skrCluster.GetScheme()) //nolint:staticcheck // SA1019
	compKcpCluster := composed.NewStateCluster(kcpCluster.GetClient(), kcpCluster.GetAPIReader(), kcpCluster.GetEventRecorderFor("cloud-control"), kcpCluster.GetScheme())   //nolint:staticcheck // SA1019
	composedStateFactory := composed.NewStateFactory(compSkrCluster)
	stateFactory := NewStateFactory(kymaRef, compKcpCluster, compSkrCluster, env)

	return Reconciler{
		composedStateFactory: composedStateFactory,
		stateFactory:         stateFactory,
		backupImpl:           getBackupImpl(scheduleType),
	}
}
