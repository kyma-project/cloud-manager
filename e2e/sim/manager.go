package sim

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var _ Manager = &simManager{}

type Manager interface {
	manager.Manager
	Start(ctx context.Context) error
}

func NewManager(clsrt cluster.Cluster, logger logr.Logger) Manager {
	return &simManager{
		Cluster: clsrt,
		logger:  logger,
	}
}

type simManager struct {
	cluster.Cluster

	logger      logr.Logger
	controllers []manager.Runnable

	started bool
	stopped bool
}

func (m *simManager) IsStarted() bool {
	return m.started
}

func (m *simManager) IsStopped() bool {
	return m.stopped
}

func (m *simManager) Start(ctx context.Context) error {
	if m.started {
		return errors.New("manager already started")
	}
	var wg sync.WaitGroup
	var result error
	for _, r := range m.controllers {
		rr := r
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := rr.Start(ctx)
			if err != nil {
				result = multierror.Append(result, err)
				logger := m.logger
				if cc, ok := rr.(controller.Controller); ok {
					logger = cc.GetLogger()
				}
				logger.Error(err, "error starting controller")
				return
			}
		}()
	}

	// start cluster only if not started - ie it returned cache.ErrCacheNotStarted
	// * manager that runs the whole sim is started with already started kcp cluster
	// * manager that runs one skr is started with not started skr cluster

	err := m.Cluster.GetCache().Get(ctx, types.NamespacedName{Namespace: "foo", Name: "foo"}, &corev1.ConfigMap{})
	if errors.Is(err, &cache.ErrCacheNotStarted{}) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := m.Cluster.Start(ctx); err != nil {
				if !strings.Contains(err.Error(), "already started") {
					result = multierror.Append(result, err)
					m.logger.Error(err, "error starting cluster")
				}
			}
		}()
	}

	<-ctx.Done()
	wg.Wait()
	m.stopped = true

	return result
}

func (m *simManager) Add(runnable manager.Runnable) error {
	m.controllers = append(m.controllers, runnable)
	return nil
}

func (m *simManager) Elected() <-chan struct{} {
	//TODO implement me
	panic("implement me")
}

func (m *simManager) AddMetricsServerExtraHandler(path string, handler http.Handler) error {
	//TODO implement me
	panic("implement me")
}

func (m *simManager) AddHealthzCheck(name string, check healthz.Checker) error {
	//TODO implement me
	panic("implement me")
}

func (m *simManager) AddReadyzCheck(name string, check healthz.Checker) error {
	//TODO implement me
	panic("implement me")
}

func (m *simManager) GetWebhookServer() webhook.Server {
	//TODO implement me
	panic("implement me")
}

func (m *simManager) GetLogger() logr.Logger {
	return ctrl.Log
}

func (m *simManager) GetControllerOptions() config.Controller {
	result := config.Controller{
		GroupKindConcurrency:    nil,
		MaxConcurrentReconciles: 1,
		CacheSyncTimeout:        0,
		RecoverPanic:            nil,
		NeedLeaderElection:      nil,
		SkipNameValidation:      ptr.To(true),
	}
	return result
}
