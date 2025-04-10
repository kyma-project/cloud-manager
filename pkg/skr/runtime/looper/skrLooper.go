package looper

import (
	"context"
	"errors"
	"github.com/elliotchance/pie/v2"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/kyma-project/cloud-manager/pkg/metrics"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/config"
	skrmanager "github.com/kyma-project/cloud-manager/pkg/skr/runtime/manager"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/registry"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/kyma-project/cloud-manager/pkg/util/debugged"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sync"
	"time"
)

type ActiveSkrCollection interface {
	AddScope(scope *cloudcontrolv1beta1.Scope)
	RemoveScope(scope *cloudcontrolv1beta1.Scope)
	AddKyma(kyma *unstructured.Unstructured)
	RemoveKyma(kyma *unstructured.Unstructured)
	Contains(kymaName string) bool
	GetKymaNames() []string
}

type ActiveSkrCollectionAdmin interface {
	ActiveSkrCollection
	Logger() logr.Logger
	Queue() *CyclicQueue
}

//type activeSkrCtxKeyType string
//
//const activeSkrCtxKey activeSkrCtxKeyType = "activeSkrCtxKey"
//
//func ActiveSkrCollectionToCtx(ctx context.Context, val ActiveSkrCollection) context.Context {
//	return context.WithValue(ctx, activeSkrCtxKey, val)
//}
//
//func ActiveSkrCollectionFromCtx(ctx context.Context) ActiveSkrCollection {
//	val := ctx.Value(activeSkrCtxKey)
//	return val.(ActiveSkrCollection)
//}

func NewActiveSkrCollection(logger logr.Logger) ActiveSkrCollectionAdmin {
	return &activeSkrCollection{
		queue:  NewCyclicQueue(),
		logger: logger,
	}
}

var _ ActiveSkrCollectionAdmin = &activeSkrCollection{}

type activeSkrCollection struct {
	queue  *CyclicQueue
	logger logr.Logger
}

func (l *activeSkrCollection) Logger() logr.Logger {
	return l.logger
}

func (l *activeSkrCollection) Queue() *CyclicQueue {
	return l.queue
}

func (l *activeSkrCollection) AddScope(scope *cloudcontrolv1beta1.Scope) {
	l.add(scope)
}

func (l *activeSkrCollection) AddKyma(kyma *unstructured.Unstructured) {
	l.add(kyma)
}

func (l *activeSkrCollection) add(obj client.Object) {
	kymaName := obj.GetName()

	if l.queue.Contains(kymaName) {
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

	l.logger.WithValues(
		"kymaName", kymaName,
		"globalAccountId", globalAccountId,
		"subaccountId", subaccountId,
		"shootName", shootName,
		"region", region,
		"brokerPlanName", brokerPlanName,
	).Info("Adding Kyma to SkrLooper")

	l.queue.Add(kymaName)

	metrics.
		SkrRuntimeModuleActiveCount.WithLabelValues(kymaName, globalAccountId, subaccountId, shootName, region, brokerPlanName).
		Add(1)
}

func (l *activeSkrCollection) RemoveScope(scope *cloudcontrolv1beta1.Scope) {
	l.remove(scope)
}

func (l *activeSkrCollection) RemoveKyma(kyma *unstructured.Unstructured) {
	l.remove(kyma)
}

func (l *activeSkrCollection) remove(obj client.Object) {
	kymaName := obj.GetName()
	if !l.queue.Contains(kymaName) {
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

	l.logger.WithValues(
		"kymaName", kymaName,
		"globalAccountId", globalAccountId,
		"subaccountId", subaccountId,
		"shootName", shootName,
		"region", region,
		"brokerPlanName", brokerPlanName,
	).Info("Removing Kyma from SkrLooper")

	l.queue.Remove(kymaName)

	metrics.
		SkrRuntimeModuleActiveCount.WithLabelValues(kymaName, globalAccountId, subaccountId, shootName, region, brokerPlanName).
		Add(-1)
}

func (l *activeSkrCollection) Contains(kymaName string) bool {
	return l.queue.Contains(kymaName)
}

func (l *activeSkrCollection) GetKymaNames() []string {
	return pie.Map(l.queue.Items(), func(x interface{}) string {
		return x.(string)
	})
}

// =====================================================================

type SkrLooper interface {
	manager.Runnable
	ActiveSkrCollection
}

func New(activeSkrCollection ActiveSkrCollectionAdmin, kcpCluster cluster.Cluster, skrScheme *runtime.Scheme, reg registry.SkrRegistry, logger logr.Logger) SkrLooper {
	return &skrLooper{
		ActiveSkrCollectionAdmin: activeSkrCollection,
		logger:                   logger,
		kcpCluster:               kcpCluster,
		managerFactory:           skrmanager.NewFactory(kcpCluster.GetAPIReader(), "kcp-system", skrScheme),
		skrStatusSaver:           NewSkrStatusSaver(NewSkrStatusRepo(kcpCluster.GetClient()), "kcp-system"),
		registry:                 reg,
		concurrency:              config.SkrRuntimeConfig.Concurrency,
	}
}

type skrLooper struct {
	ActiveSkrCollectionAdmin

	logger         logr.Logger
	kcpCluster     cluster.Cluster
	managerFactory skrmanager.Factory
	registry       registry.SkrRegistry
	concurrency    int
	skrStatusSaver SkrStatusSaver

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

	l.logger.Info("SkrLooper started")
	l.ctx = ctx
	l.wg.Add(l.concurrency)
	for x := 0; x < l.concurrency; x++ {
		go l.worker(x)
	}

	<-ctx.Done()

	l.logger.Info("SkrLooper context closed, shutting down the queue")
	l.Queue().Shutdown()
	l.logger.Info("SkrLooper waiting workers to finish")
	l.wg.Wait()
	l.logger.Info("SkrLooper stopped")
	return nil
}

func (l *skrLooper) worker(id int) {
	defer l.wg.Done()
	logger := l.logger.WithValues("skrWorkerId", id)
	logger.Info("SKR Looper worker started")
	for {
		shouldStop := func() bool {
			item, shuttingDown := l.Queue().Get()
			defer l.Queue().Done(item)
			if shuttingDown {
				logger.Info("SKR Looper worker shutting down")
				return true
			}
			kymaName := item.(string)
			time.Sleep(util.Timing.T100ms())
			l.handleOneSkr(id, kymaName)
			return false
		}()
		if shouldStop {
			break
		}
	}
	logger.Info("SKR Looper return from worker")
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
	skrManager, scope, kyma, err := l.managerFactory.CreateManager(ctx, kymaName, logger)
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
	skrManager.GetScheme()

	ctx = feature.ContextBuilderFromCtx(ctx).
		Landscape(os.Getenv("LANDSCAPE")).
		LoadFromScope(scope).
		LoadFromKyma(kyma).
		Plane(types.PlaneSkr).
		Build(ctx)

	logger = feature.DecorateLogger(ctx, logger)

	//logger.Info("Starting SKR Runner")
	runner := NewSkrRunner(l.registry, l.kcpCluster, l.skrStatusSaver, kymaName)
	to := 10 * time.Second
	if debugged.Debugged {
		to = 15 * time.Minute
	}

	err = runner.Run(ctx, skrManager, WithTimeout(to), WithProvider(scope.Spec.Provider))
	if err != nil {
		logger.Error(err, "Error running SKR Runner")
	}
	//logger.Info("SKR Runner stopped")
}
