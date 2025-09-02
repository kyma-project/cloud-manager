package sim

import (
	"context"
	"net/http"
	"sync"

	"github.com/go-logr/logr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var _ manager.Manager = &simManager{}

func NewManager(clsrt cluster.Cluster, logger logr.Logger) manager.Manager {
	return &simManager{
		Cluster: clsrt,
		logger:  logger,
	}
}

type simManager struct {
	cluster.Cluster

	logger      logr.Logger
	controllers []manager.Runnable
}

func (m *simManager) Start(ctx context.Context) error {
	var wg sync.WaitGroup
	for _, r := range m.controllers {
		rr := r
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := rr.Start(ctx)
			if err != nil {
				logger := m.logger
				if ctrl, ok := rr.(controller.Controller); ok {
					logger = ctrl.GetLogger()
				}
				logger.Error(err, "error starting controller")
				return
			}
		}()
	}

	<-ctx.Done()
	wg.Wait()

	return nil
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
	return m.logger
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
