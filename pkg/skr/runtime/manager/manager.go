package manager

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sync"
)

var _ SkrManager = &skrManager{}

type SkrManager interface {
	manager.Manager
	KymaRef() klog.ObjectRef
}

func New(cfg *rest.Config, skrScheme *runtime.Scheme, kymaRef klog.ObjectRef, logger logr.Logger) (SkrManager, error) {
	cls, err := cluster.New(cfg, func(clusterOptions *cluster.Options) {
		clusterOptions.Scheme = skrScheme
		clusterOptions.Logger = logger
		clusterOptions.Client = client.Options{
			Cache: &client.CacheOptions{
				Unstructured: true,
			},
		}
	})
	if err != nil {
		return nil, err
	}
	return &skrManager{
		Cluster: cls,
		kymaRef: kymaRef,
		logger:  logger,
	}, nil
}

type skrManager struct {
	cluster.Cluster

	kymaRef     klog.ObjectRef
	logger      logr.Logger
	controllers []manager.Runnable
}

func (m *skrManager) KymaRef() klog.ObjectRef {
	return m.kymaRef
}

func (m *skrManager) Start(ctx context.Context) error {
	m.logger.Info("SkrManager starting")
	m.controllers = append(m.controllers, m.Cluster)
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
	m.logger.Info("SkrManager stopped")

	return nil
}

func (m *skrManager) Add(runnable manager.Runnable) error {
	m.controllers = append(m.controllers, runnable)
	return nil
}

func (m *skrManager) Elected() <-chan struct{} {
	//TODO implement me
	panic("implement me")
}

func (m *skrManager) AddMetricsServerExtraHandler(path string, handler http.Handler) error {
	//TODO implement me
	panic("implement me")
}

func (m *skrManager) AddHealthzCheck(name string, check healthz.Checker) error {
	//TODO implement me
	panic("implement me")
}

func (m *skrManager) AddReadyzCheck(name string, check healthz.Checker) error {
	//TODO implement me
	panic("implement me")
}

func (m *skrManager) GetWebhookServer() webhook.Server {
	//TODO implement me
	panic("implement me")
}

func (m *skrManager) GetLogger() logr.Logger {
	return m.logger
}

func (m *skrManager) GetControllerOptions() config.Controller {
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
