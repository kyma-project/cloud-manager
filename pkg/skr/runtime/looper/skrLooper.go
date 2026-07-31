package looper

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/kyma-project/cloud-manager/pkg/metrics"
	skrruntimeconfig "github.com/kyma-project/cloud-manager/pkg/skr/runtime/config"
	skrmanager "github.com/kyma-project/cloud-manager/pkg/skr/runtime/manager"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/registry"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/kyma-project/cloud-manager/pkg/util/debugged"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type ActiveSkrCollection interface {
	AddScope(ctx context.Context, scope *cloudcontrolv1beta1.Scope)
	RemoveScope(ctx context.Context, scope *cloudcontrolv1beta1.Scope)
	AddKyma(ctx context.Context, kyma *unstructured.Unstructured)
	RemoveKyma(ctx context.Context, kyma *unstructured.Unstructured)
	// Notify enqueues a notification-driven reconcile for kymaName into the fast
	// notification sleeve. It is a no-op if the SKR is not currently active.
	Notify(kymaName string)
	Contains(kymaName string) bool
	GetKymaNames() []string
}

type ActiveSkrCollectionAdmin interface {
	ActiveSkrCollection
	CyclicQueue() *Queue
	NotificationQueue() *Queue
	Gate() *SkrGate
}

func NewActiveSkrCollection() ActiveSkrCollectionAdmin {
	return &activeSkrCollection{
		cyclicQueue: NewQueue(),
		notifQueue:  NewQueue(),
		gate:        NewSkrGate(),
	}
}

var _ ActiveSkrCollectionAdmin = &activeSkrCollection{}

type activeSkrCollection struct {
	// cyclicQueue holds active SKRs for the background round-robin sleeve. Its
	// membership set is the source of truth for "is this SKR active" (Contains).
	cyclicQueue *Queue
	// notifQueue holds notification-driven SKRs for the fast user-facing sleeve.
	notifQueue *Queue
	// gate guarantees at most one live manager per SKR across both sleeves.
	gate *SkrGate
}

func (l *activeSkrCollection) CyclicQueue() *Queue       { return l.cyclicQueue }
func (l *activeSkrCollection) NotificationQueue() *Queue { return l.notifQueue }
func (l *activeSkrCollection) Gate() *SkrGate            { return l.gate }

func (l *activeSkrCollection) AddScope(ctx context.Context, scope *cloudcontrolv1beta1.Scope) {
	l.add(ctx, scope)
}

func (l *activeSkrCollection) AddKyma(ctx context.Context, kyma *unstructured.Unstructured) {
	l.add(ctx, kyma)
}

func (l *activeSkrCollection) add(ctx context.Context, obj client.Object) {
	kymaName := obj.GetName()

	labels := obj.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	globalAccountId := labels[cloudcontrolv1beta1.LabelScopeGlobalAccountId]
	subaccountId := labels[cloudcontrolv1beta1.LabelScopeSubaccountId]
	shootName := labels[cloudcontrolv1beta1.LabelScopeShootName]
	region := labels[cloudcontrolv1beta1.LabelScopeRegion]
	brokerPlanName := labels[cloudcontrolv1beta1.LabelScopeBrokerPlanName]

	logger := composed.LoggerFromCtx(ctx)
	logger.WithValues(
		"kymaName", kymaName,
		"globalAccountId", globalAccountId,
		"subaccountId", subaccountId,
		"shootName", shootName,
		"region", region,
		"brokerPlanName", brokerPlanName,
	).Info("Adding Kyma to SkrLooper")

	// The workqueue dirty-set dedups, so a re-add of an already-queued SKR is a no-op
	// there; but the module-active metric must only count a genuine activation, so
	// guard it on the membership set.
	alreadyActive := l.cyclicQueue.Contains(kymaName)
	l.cyclicQueue.Add(kymaName)
	if !alreadyActive {
		metrics.
			SkrRuntimeModuleActiveCount.WithLabelValues(kymaName, globalAccountId, subaccountId, shootName, region, brokerPlanName).
			Add(1)
	}
}

func (l *activeSkrCollection) Notify(kymaName string) {
	// Drop notifications for SKRs that are not active (never added, or deactivated).
	// Re-activation only ever comes from the KCP reconciler via AddKyma/AddScope.
	if !l.cyclicQueue.Contains(kymaName) {
		return
	}
	l.notifQueue.Add(kymaName)
}

func (l *activeSkrCollection) RemoveScope(ctx context.Context, scope *cloudcontrolv1beta1.Scope) {
	l.remove(ctx, scope)
}

func (l *activeSkrCollection) RemoveKyma(ctx context.Context, kyma *unstructured.Unstructured) {
	l.remove(ctx, kyma)
}

func (l *activeSkrCollection) remove(ctx context.Context, obj client.Object) {
	kymaName := obj.GetName()
	if !l.cyclicQueue.Contains(kymaName) {
		return
	}

	labels := obj.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	globalAccountId := labels[cloudcontrolv1beta1.LabelScopeGlobalAccountId]
	subaccountId := labels[cloudcontrolv1beta1.LabelScopeSubaccountId]
	shootName := labels[cloudcontrolv1beta1.LabelScopeShootName]
	region := labels[cloudcontrolv1beta1.LabelScopeRegion]
	brokerPlanName := labels[cloudcontrolv1beta1.LabelScopeBrokerPlanName]

	logger := composed.LoggerFromCtx(ctx)
	logger.WithValues(
		"kymaName", kymaName,
		"globalAccountId", globalAccountId,
		"subaccountId", subaccountId,
		"shootName", shootName,
		"region", region,
		"brokerPlanName", brokerPlanName,
	).Info("Removing Kyma from SkrLooper")

	// Clear membership on both queues. This never aborts a running manager nor
	// touches the gate claim; the owning worker's Release frees the claim when its
	// manager finishes (graceful teardown).
	l.cyclicQueue.Remove(kymaName)
	l.notifQueue.Remove(kymaName)

	metrics.
		SkrRuntimeModuleActiveCount.WithLabelValues(kymaName, globalAccountId, subaccountId, shootName, region, brokerPlanName).
		Add(-1)
}

func (l *activeSkrCollection) Contains(kymaName string) bool {
	return l.cyclicQueue.Contains(kymaName)
}

func (l *activeSkrCollection) GetKymaNames() []string {
	return l.cyclicQueue.Items()
}

// =====================================================================

type SkrLooper interface {
	manager.Runnable
	ActiveSkrCollection
}

func New(activeSkrCollection ActiveSkrCollectionAdmin, kcpCluster cluster.Cluster, reg registry.SkrRegistry, logger logr.Logger) SkrLooper {
	l := &skrLooper{
		ActiveSkrCollectionAdmin: activeSkrCollection,
		logger:                   logger,
		kcpCluster:               kcpCluster,
		managerFactory:           skrmanager.NewFactory(kcpCluster.GetAPIReader(), "kcp-system"),
		skrStatusSaver:           NewSkrStatusSaver(NewSkrStatusRepo(kcpCluster.GetClient()), "kcp-system"),
		registry:                 reg,
		notificationConcurrency:  skrruntimeconfig.SkrRuntimeConfig.NotificationConcurrency,
		cyclicConcurrency:        skrruntimeconfig.SkrRuntimeConfig.CyclicConcurrency,
		cyclicMinInterval:        skrruntimeconfig.SkrRuntimeConfig.SkrCyclicMinInterval,
		gateConflictDelay:        skrruntimeconfig.SkrRuntimeConfig.SkrGateConflictRetryDelay,
		reconcileTimeout:         debugged.When(15*time.Minute, 10*time.Second),
	}
	// A full cyclic pass takes ~ fleet*reconcileTimeout/cyclicConcurrency; the
	// cyclicMinInterval floor is only binding while the fleet is below
	// cyclicConcurrency*cyclicMinInterval/reconcileTimeout. At/above that size FIFO
	// re-add is used (fair, order-stable); below it the AddAfter floor is kept.
	l.cyclicImmediateThreshold = max(1, int(time.Duration(l.cyclicConcurrency)*l.cyclicMinInterval/l.reconcileTimeout))
	l.handleFn = l.handleOneSkr
	return l
}

type skrLooper struct {
	ActiveSkrCollectionAdmin

	logger         logr.Logger
	kcpCluster     cluster.Cluster
	managerFactory skrmanager.Factory
	registry       registry.SkrRegistry
	skrStatusSaver SkrStatusSaver

	notificationConcurrency int
	cyclicConcurrency       int
	cyclicMinInterval       time.Duration
	gateConflictDelay       time.Duration

	// reconcileTimeout bounds a single handleOneSkr connection (10s normally, 15min
	// under the `debug` build tag). It is also the denominator of the cyclic re-add
	// threshold, so keeping it in one field keeps the two in sync.
	reconcileTimeout time.Duration

	// cyclicImmediateThreshold is the active-fleet size at/above which the cyclic
	// worker re-adds an SKR immediately (FIFO) instead of via AddAfter(cyclicMinInterval).
	// Above it a full cyclic pass already exceeds cyclicMinInterval, so the interval
	// floor is non-binding and the AddAfter readyAt heap only destabilises rotation
	// order; FIFO gives a stable, fair round-robin. Below it the floor still protects
	// a small fleet from hot-looping. Derived in New() from the worker count.
	cyclicImmediateThreshold int

	// handleFn is the per-SKR handler; defaults to handleOneSkr and is injectable
	// so the pool logic is testable without a live per-SKR manager (envtest).
	handleFn func(skrWorkerId int, kymaName string)

	// wg the WorkGroup for workers
	wg      sync.WaitGroup
	started bool

	// ctx the Context looper was started with
	ctx context.Context
}

func (l *skrLooper) Start(ctx context.Context) error {
	if l.started {
		return errors.New("looper already started")
	}
	l.started = true

	l.logger.Info(
		"SkrLooper started",
		"notificationConcurrency", l.notificationConcurrency,
		"cyclicConcurrency", l.cyclicConcurrency,
		"cyclicImmediateThreshold", l.cyclicImmediateThreshold,
	)
	l.ctx = ctx

	l.wg.Add(l.notificationConcurrency + l.cyclicConcurrency)
	for x := 0; x < l.notificationConcurrency; x++ {
		go l.notificationWorker(x)
	}
	for x := 0; x < l.cyclicConcurrency; x++ {
		go l.cyclicWorker(x)
	}

	<-ctx.Done()

	l.logger.Info("SkrLooper context closed, shutting down the queues")
	l.NotificationQueue().ShutDown()
	l.CyclicQueue().ShutDown()
	l.logger.Info("SkrLooper waiting workers to finish")
	l.wg.Wait()
	l.logger.Info("SkrLooper stopped")
	return nil
}

// processOne performs one guarded Get→handle cycle on q. It returns true when the
// queue is shutting down (the worker should exit). reAdd runs on the SUCCESS path
// only (after handle returns) — never on the shutdown, membership-drop, or gate-
// conflict paths.
func (l *skrLooper) processOne(id int, q *Queue, sleeve string, reAdd func(kymaName string)) bool {
	item, shuttingDown := q.Get()
	if shuttingDown {
		return true
	}
	func() {
		defer q.Done(item) // OUTERMOST defer

		// Guard 1 — membership: drop stray work for deactivated runtimes. A stale
		// delayed re-add OR a stray notification can deliver a kymaName that was
		// removed while queued. Re-activation only ever comes from the KCP reconciler.
		if !l.Contains(item) {
			return // drop: no claim, no connect, no re-add
		}

		// Guard 2 — cross-sleeve single-manager guarantee.
		if !l.Gate().TryClaim(item) {
			metrics.SkrLooperGateConflictTotal.WithLabelValues(sleeve).Inc()
			q.AddAfter(item, l.gateConflictDelay) // flat delay, to back; move on
			return
		}
		defer l.Gate().Release(item) // INNER defer — runs before q.Done (LIFO)

		l.handleFn(id, item)

		reAdd(item) // success path only
	}()
	return false
}

func (l *skrLooper) notificationWorker(id int) {
	defer l.wg.Done()
	logger := l.logger.WithValues("skrNotificationWorkerId", id)
	logger.Info("SKR Looper notification worker started")
	q := l.NotificationQueue()
	// FIFO drain: no self re-add. On success also push the cyclic entry so the
	// background sleeve does not immediately re-do the same work. Guarded on membership
	// so a Remove that fired mid-flight is not undone by re-adding to the cyclic queue.
	// Delay is Add(), i.e. move-to-tail — consistent with the cyclic worker's FIFO re-add
	// path: both mean "take a normal turn at the back of the cyclic rotation".
	for !l.processOne(id, q, "notification", func(kymaName string) {
		if l.Contains(kymaName) {
			l.CyclicQueue().Delay(kymaName)
		}
	}) {
	}
	logger.Info("SKR Looper notification worker returning")
}

func (l *skrLooper) cyclicWorker(id int) {
	defer l.wg.Done()
	logger := l.logger.WithValues("skrCyclicWorkerId", id)
	logger.Info("SKR Looper cyclic worker started")
	q := l.CyclicQueue()
	for !l.processOne(id, q, "cyclic", l.cyclicReAdd) {
	}
	logger.Info("SKR Looper cyclic worker returning")
}

// cyclicReAdd re-schedules an SKR for the next cyclic pass on the success path. It
// re-adds only while still a member: a Remove that fired during handle (the SKR's own
// "cleanup done, safe to deactivate" signal) must stop the cycle — otherwise re-adding
// would re-record membership and re-activate a deactivated SKR.
//
// Above cyclicImmediateThreshold the fleet is large enough that a full pass already
// exceeds cyclicMinInterval, so the interval floor is non-binding and routing through
// AddAfter's readyAt heap only destabilises rotation order (per-cycle handle-time
// variance random-walks each SKR's heap position, causing unfair, oscillating cadence).
// Immediate FIFO Add appends to the ready tail and preserves rotation order across
// cycles → stable, fair round-robin. Below the threshold the small fleet keeps the
// AddAfter floor so a handful of SKRs is not hot-looped.
func (l *skrLooper) cyclicReAdd(kymaName string) {
	q := l.CyclicQueue()
	if !l.Contains(kymaName) {
		return
	}
	if q.MembershipLen() >= l.cyclicImmediateThreshold {
		q.Add(kymaName)
	} else {
		q.AddAfter(kymaName, l.cyclicMinInterval)
	}
}

func (l *skrLooper) handleOneSkr(skrWorkerId int, kymaName string) {
	defer func() {
		metrics.SkrRuntimeReconcileTotal.WithLabelValues(kymaName).Inc()
	}()
	logger := l.logger.WithValues(
		"skrWorkerId", skrWorkerId,
		"kyma", kymaName,
	)
	ctx := composed.LoggerIntoCtx(l.ctx, logger)
	skrManager, scope, err := l.managerFactory.CreateManager(ctx, kymaName, logger)
	if errors.Is(err, context.DeadlineExceeded) {
		return
	}
	if errors.Is(err, context.Canceled) {
		return
	}
	if errors.Is(err, &skrmanager.ScopeNotFoundError{}) {
		logger.
			WithValues("error", err.Error()).
			Info("SKR scope not found")
		time.Sleep(util.Timing.T100ms())
		return
	}
	if err != nil {
		logger.Error(err, "error creating Manager")
		time.Sleep(util.Timing.T100ms())
		return
	}

	ctx = feature.ContextBuilderFromCtx(ctx).
		Landscape(os.Getenv("LANDSCAPE")).
		LoadFromScope(scope).
		Plane(types.PlaneSkr).
		Build(ctx)

	logger = feature.DecorateLogger(ctx, logger)

	runner := NewSkrRunner(l.registry, l.kcpCluster, l.skrStatusSaver, kymaName)

	err = runner.Run(ctx, skrManager, WithTimeout(l.reconcileTimeout), WithProvider(scope.Spec.Provider))
	if util.IgnoreContextCanceledAndDeadlineExceeded(err) != nil {
		if !apierrors.IsTimeout(err) {
			logger.Error(err, "Error running SKR Runner")
		}
	}
}
