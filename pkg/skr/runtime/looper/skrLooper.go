package looper

import (
	"context"
	"errors"
	"github.com/elliotchance/pie/v2"
	"github.com/go-logr/logr"
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
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sync"
	"time"
)

type ActiveSkrCollection interface {
	AddKymaUnstructured(kyma *unstructured.Unstructured)
	RemoveKymaUnstructured(kyma *unstructured.Unstructured)
	Contains(kymaName string) bool
	GetKymaNames() []string
}

func NewActiveSkrCollection(logger logr.Logger) ActiveSkrCollection {
	return &activeSkrCollection{
		queue:  NewCyclicQueue(),
		logger: logger,
	}
}

type activeSkrCollection struct {
	queue  *CyclicQueue
	logger logr.Logger
}

func (l *activeSkrCollection) AddKymaUnstructured(kyma *unstructured.Unstructured) {

	kymaName := kyma.GetName()

	if l.queue.Contains(kymaName) {
		return
	}

	labels := kyma.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	globalAccountId := labels["kyma-project.io/global-account-id"]
	subaccountId := labels["kyma-project.io/subaccount-id"]
	shootName := labels["kyma-project.io/shoot-name"]
	region := labels["kyma-project.io/region"]
	brokerPlanName := labels["kyma-project.io/broker-plan-name"]

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

func (l *activeSkrCollection) RemoveKymaUnstructured(kyma *unstructured.Unstructured) {
	kymaName := kyma.GetName()
	if !l.queue.Contains(kymaName) {
		return
	}

	labels := kyma.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	globalAccountId := labels["kyma-project.io/global-account-id"]
	subaccountId := labels["kyma-project.io/subaccount-id"]
	shootName := labels["kyma-project.io/shoot-name"]
	region := labels["kyma-project.io/region"]
	brokerPlanName := labels["kyma-project.io/broker-plan-name"]

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

func New(kcpCluster cluster.Cluster, skrScheme *runtime.Scheme, reg registry.SkrRegistry, logger logr.Logger) SkrLooper {
	return &skrLooper{
		activeSkrCollection: NewActiveSkrCollection(logger).(*activeSkrCollection),
		kcpCluster:          kcpCluster,
		managerFactory:      skrmanager.NewFactory(kcpCluster.GetAPIReader(), "kcp-system", skrScheme),
		registry:            reg,
		concurrency:         config.SkrRuntimeConfig.Concurrency,
	}
}

type skrLooper struct {
	*activeSkrCollection

	kcpCluster     cluster.Cluster
	managerFactory skrmanager.Factory
	registry       registry.SkrRegistry
	concurrency    int

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

	l.queue.Shutdown()
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
			logger.Info("SKR Looper worker about to read SKR ID")
			item, shuttingDown := l.queue.Get()
			defer l.queue.Done(item)
			if shuttingDown {
				return true
			}
			kymaName := item.(string)
			time.Sleep(time.Second)
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
		"skrKymaName", kymaName,
	)
	skrManager, scope, kyma, err := l.managerFactory.CreateManager(l.ctx, kymaName, logger)
	if errors.Is(err, &skrmanager.ScopeNotFoundError{}) {
		logger.
			WithValues("error", err.Error()).
			Info("SKR scope not found")
		time.Sleep(util.Timing.T1000ms())
		return
	}
	if err != nil {
		logger.Error(err, "error creating Manager")
		time.Sleep(util.Timing.T1000ms())
		return
	}
	skrManager.GetScheme()

	ctx := feature.ContextBuilderFromCtx(l.ctx).
		Landscape(os.Getenv("LANDSCAPE")).
		LoadFromScope(scope).
		LoadFromKyma(kyma).
		Plane(types.PlaneSkr).
		Build(l.ctx)

	logger = feature.DecorateLogger(ctx, logger)

	logger.Info("Starting SKR Runner")
	runner := NewSkrRunner(l.registry, l.kcpCluster)
	to := 10 * time.Second
	if debugged.Debugged {
		to = 15 * time.Minute
	}

	err = runner.Run(ctx, skrManager, WithTimeout(to), WithProvider(scope.Spec.Provider))
	if err != nil {
		logger.Error(err, "Error running SKR Runner")
	}
	logger.Info("SKR Runner stopped")
}
