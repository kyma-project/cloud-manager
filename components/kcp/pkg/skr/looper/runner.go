package looper

import (
	"context"
	"github.com/go-logr/logr"
	skrmanager "github.com/kyma-project/cloud-manager/components/kcp/pkg/skr/manager"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/skr/registry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
	"sync"
	"time"
)

type RunOptions struct {
	timeout time.Duration
}

type RunOption = func(options *RunOptions)

func WithTimeout(timeout time.Duration) RunOption {
	return func(options *RunOptions) {
		options.timeout = timeout
	}
}

type SkrRunner interface {
	Run(ctx context.Context, mngr skrmanager.SkrManager, opts ...RunOption)
	GetControllerOptions(
		descriptor registry.Descriptor,
		skrManager skrmanager.SkrManager,
	) controller.Options
}

func NewSkrRunner(reg registry.SkrRegistry, kcpCluster cluster.Cluster) SkrRunner {
	return &skrRunner{
		kcpCluster: kcpCluster,
		registry:   reg,
	}
}

type skrRunner struct {
	kcpCluster cluster.Cluster
	registry   registry.SkrRegistry
	runOnce    sync.Once
}

func (r *skrRunner) Run(ctx context.Context, mngr skrmanager.SkrManager, opts ...RunOption) {
	options := &RunOptions{}
	for _, o := range opts {
		o(options)
	}
	r.runOnce.Do(func() {
		for _, descr := range r.registry.GetDescriptors(mngr) {
			logger2 := mngr.GetLogger().WithValues("controller", descr.Name)
			ctrl, err := controller.New(descr.Name, mngr, r.GetControllerOptions(
				descr,
				mngr,
			))
			if err != nil {
				logger2.Error(err, "error creating controller")
				continue
			}
			for _, w := range descr.Watches {
				err = ctrl.Watch(w.Src, w.EventHandler, w.Predicates...)
				if err != nil {
					logger2.WithValues("watch", w.Name).Error(err, "error watching source")
					continue
				}
			}
		}

		if options.timeout == 0 {
			options.timeout = time.Minute
		}
		timeoutCtx, cancel := context.WithTimeout(ctx, options.timeout)
		defer cancel()
		err := mngr.Start(timeoutCtx)
		if err != nil {
			mngr.GetLogger().Error(err, "error starting manager")
		}
	})
}

func (r *skrRunner) GetControllerOptions(descriptor registry.Descriptor, skrManager skrmanager.SkrManager) controller.Options {
	controllerName := descriptor.Name
	if len(controllerName) == 0 {
		controllerName = strings.ToLower(descriptor.GVK.Kind)
	}
	logger := skrManager.GetLogger().WithValues(
		"controller", controllerName,
		"controllerGroup", descriptor.GVK.Group,
		"controllerKind", descriptor.GVK.Kind,
	)
	ctrlOptions := controller.Options{
		Reconciler: descriptor.ReconcilerFactory.New(r.kcpCluster, skrManager),
		// TODO: copy other options,
		// for details check:
		//  * sigs.k8s.io/controller-runtime@v0.16.3/pkg/controller/controller.go
		//  * sigs.k8s.io/controller-runtime@v0.16.3/pkg/builder/controller.go
		MaxConcurrentReconciles: skrManager.GetControllerOptions().MaxConcurrentReconciles,
		LogConstructor: func(req *reconcile.Request) logr.Logger {
			logger := logger
			if req != nil {
				logger = logger.WithValues(
					descriptor.GVK.Kind, klog.KRef(req.Namespace, req.Name),
					"namespace", req.Namespace, "name", req.Name,
				)
			}
			return logger
		},
	}
	return ctrlOptions
}
