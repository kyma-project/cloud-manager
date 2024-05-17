package looper

import (
	"context"
	"errors"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/kyma-project/cloud-manager/pkg/metrics"
	skrmanager "github.com/kyma-project/cloud-manager/pkg/skr/runtime/manager"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/registry"
	"github.com/kyma-project/cloud-manager/pkg/util/debugged"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
	return &activeSkrCollection{logger: logger}
}

type activeSkrCollection struct {
	sync.RWMutex
	kymaNames []string
	logger    logr.Logger
}

func (l *activeSkrCollection) AddKymaUnstructured(kyma *unstructured.Unstructured) {
	l.Lock()
	defer l.Unlock()

	kymaName := kyma.GetName()

	if pie.Contains(l.kymaNames, kymaName) {
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

	l.kymaNames = append(l.kymaNames, kymaName)

	metrics.SkrRuntimeModuleActiveCount.WithLabelValues(kymaName, globalAccountId, subaccountId, shootName, region, brokerPlanName).Add(1)
}

func (l *activeSkrCollection) RemoveKymaUnstructured(kyma *unstructured.Unstructured) {
	l.Lock()
	defer l.Unlock()

	kymaName := kyma.GetName()
	idx := pie.FindFirstUsing(l.kymaNames, func(value string) bool {
		return value == kymaName
	})
	if idx < 0 {
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

	l.kymaNames = pie.Delete(l.kymaNames, idx)

	metrics.SkrRuntimeModuleActiveCount.WithLabelValues(kymaName, globalAccountId, subaccountId, shootName, region, brokerPlanName).Add(-1)
}

func (l *activeSkrCollection) Contains(kymaName string) bool {
	l.Lock()
	defer l.Unlock()
	if pie.Contains(l.kymaNames, kymaName) {
		return true
	}
	return false
}

func (l *activeSkrCollection) GetKymaNames() []string {
	l.RLock()
	defer l.RUnlock()
	var kymaNames []string
	kymaNames = make([]string, len(l.kymaNames))
	for x := range l.kymaNames {
		kymaNames[x] = l.kymaNames[x]
	}
	return kymaNames
}

// =====================================================================

type SkrLooper interface {
	manager.Runnable
	ActiveSkrCollection
}

func New(kcpCluster cluster.Cluster, skrScheme *runtime.Scheme, reg registry.SkrRegistry, logger logr.Logger) SkrLooper {
	return &skrLooper{
		activeSkrCollection: &activeSkrCollection{logger: logger},
		kcpCluster:          kcpCluster,
		managerFactory:      skrmanager.NewFactory(kcpCluster.GetAPIReader(), "kcp-system", skrScheme),
		registry:            reg,
		concurrency:         1,
		ch:                  make(chan string, 1),
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
	stopped bool

	// ch the channel trough which kymaNames sent to the workers
	ch chan string

	// ctx the Context looper was started with
	ctx context.Context

	// kymaNames slice of active SKRs that have to be looped trough
	kymaNames []string
}

func (l *skrLooper) Start(ctx context.Context) error {
	if l.started {
		return errors.New("looper already started")
	}
	l.started = true

	l.logger.Info("SkrLooper started")
	l.ctx = ctx
	l.ch = make(chan string)
	l.wg.Add(1)
	go l.emitActiveKymaNames()
	for x := 0; x < l.concurrency; x++ {
		l.wg.Add(1)
		go l.worker(x)
	}

	<-ctx.Done()

	l.stopped = true
	close(l.ch)
	l.wg.Wait()
	l.ch = nil
	l.logger.Info("SkrLooper stopped")
	return nil
}

func (l *skrLooper) emitActiveKymaNames() {
	defer l.wg.Done()
	//l.logger.Info("SKR Looper emitter started")
	for !l.stopped {
		//l.logger.Info("SKR Looper emitter getting active kymas")
		kymaNames := l.GetKymaNames()
		l.logger.Info(fmt.Sprintf("SKR Looper emitter got %d active kymas", len(kymaNames)))
		for _, kn := range kymaNames {
			if l.stopped {
				return
			}
			//l.logger.WithValues("kymaName", kn).Info("SKR Looper emitter about to write to ch")
			select {
			case <-l.ctx.Done():
				l.logger.Info("SKR Looper emitter context closed")
				return
			case l.ch <- kn:
				l.logger.WithValues("kymaName", kn).Info("SKR Looper emitter wrote to ch")
			}
		}
		time.Sleep(time.Second)
	}
	l.logger.Info("SKR Looper return from emitter")
}

func (l *skrLooper) worker(id int) {
	defer l.wg.Done()
	l.logger.Info(fmt.Sprintf("SKR Looper worker %d started", id))
	for !l.stopped {
		l.logger.Info(fmt.Sprintf("SKR Looper worker %d about to read from ch", id))
		kymaName := <-l.ch
		if l.stopped {
			break
		}
		time.Sleep(time.Second)
		l.handleOneSkr(kymaName)
	}
	l.logger.Info(fmt.Sprintf("SKR Looper return from worker %d", id))
}

func (l *skrLooper) handleOneSkr(kymaName string) {
	defer func() {
		metrics.SkrRuntimeReconcileTotal.WithLabelValues(kymaName).Inc()
	}()
	logger := l.logger.WithValues("skrKymaName", kymaName)
	skrManager, scope, kyma, err := l.managerFactory.CreateManager(l.ctx, kymaName, logger)
	if err != nil {
		logger.Error(err, "error creating Manager")
		time.Sleep(5 * time.Second)
		return
	}
	skrManager.GetScheme()

	ctx := feature.ContextBuilderFromCtx(l.ctx).
		LoadFromScope(scope).
		LoadFromKyma(kyma).
		Plane(types.PlaneSkr).
		Build(l.ctx)

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
