package looper

import (
	"context"
	"errors"
	"github.com/elliotchance/pie/v2"
	"github.com/go-logr/logr"
	skrmanager "github.com/kyma-project/cloud-manager/components/kcp/pkg/skr/manager"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/skr/registry"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sync"
	"time"
)

type SkrLooper interface {
	manager.Runnable
	AddKymaName(kymaName string)
	RemoveKymaName(kymaName string)
}

func New(kcpCluster cluster.Cluster, skrScheme *runtime.Scheme, reg registry.SkrRegistry, logger logr.Logger) SkrLooper {
	return &skrLooper{
		kcpCluster:     kcpCluster,
		managerFactory: skrmanager.NewFactory(kcpCluster.GetAPIReader(), "kcp-system", skrScheme),
		registry:       reg,
		logger:         logger,
		concurrency:    1,
		ch:             make(chan string, 1),
	}
}

type skrLooper struct {
	mu sync.RWMutex

	kcpCluster     cluster.Cluster
	managerFactory skrmanager.Factory
	registry       registry.SkrRegistry
	logger         logr.Logger
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

func (l *skrLooper) AddKymaName(kymaName string) {
	l.mu.Lock()
	l.kymaNames = append(l.kymaNames, kymaName)
	l.mu.Unlock()
}

func (l *skrLooper) RemoveKymaName(kymaName string) {
	l.mu.Lock()
	idx := pie.FindFirstUsing(l.kymaNames, func(value string) bool {
		return value == kymaName
	})
	if idx > -1 {
		l.kymaNames = pie.Delete(l.kymaNames, idx)
	}
	l.mu.Unlock()
}

func (l *skrLooper) Start(ctx context.Context) error {
	if l.started {
		return errors.New("looper already started")
	}

	l.logger.Info("SkrLooper started")
	l.ctx = ctx
	l.ch = make(chan string, l.concurrency)
	l.started = true
	for x := 0; x < l.concurrency; x++ {
		go l.worker()
	}
	go l.emitActiveKymaNames()

	<-ctx.Done()

	l.stopped = true
	l.wg.Wait()
	close(l.ch)
	l.ch = nil
	l.logger.Info("SkrLooper stopped")
	return nil
}

func (l *skrLooper) getKymaNames() []string {
	var kymaNames []string
	l.mu.RLock()
	kymaNames = make([]string, len(l.kymaNames))
	for x := range l.kymaNames {
		kymaNames[x] = l.kymaNames[x]
	}
	l.mu.RUnlock()
	return kymaNames
}

func (l *skrLooper) emitActiveKymaNames() {
	l.wg.Add(1)
	defer l.wg.Done()
	for !l.stopped {
		kymaNames := l.getKymaNames()
		for _, kn := range kymaNames {
			if l.stopped {
				return
			}
			select {
			case <-l.ctx.Done():
				return
			case l.ch <- kn:
			}
		}
		time.Sleep(time.Second)
	}
}

func (l *skrLooper) worker() {
	l.wg.Add(1)
	defer l.wg.Done()
	for !l.stopped {
		kymaName := <-l.ch
		if l.stopped {
			break
		}
		time.Sleep(time.Second)
		l.handleOneSkr(kymaName)
	}
}

func (l *skrLooper) handleOneSkr(kymaName string) {
	logger := l.logger.WithValues("skrKymaName", kymaName)
	mngr, err := l.managerFactory.CreateManager(l.ctx, kymaName, logger)
	if err != nil {
		logger.Error(err, "error creating Manager")
		time.Sleep(5 * time.Second)
		return
	}

	logger.Info("Starting runner")
	runner := NewSkrRunner(l.registry, mngr)
	runner.Run(l.ctx, mngr, WithTimeout(30*time.Second))
	logger.Info("Runner stopped")
}
