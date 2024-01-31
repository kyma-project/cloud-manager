package looper

import (
	"context"
	"github.com/go-logr/logr"
	skrmanager "github.com/kyma-project/cloud-manager/pkg/skr/runtime/manager"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/registry"
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
	Run(ctx context.Context, skrManager skrmanager.SkrManager, opts ...RunOption)
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
	started    bool
	stopped    bool
}

func (r *skrRunner) Run(ctx context.Context, skrManager skrmanager.SkrManager, opts ...RunOption) {
	if r.started {
		return
	}
	options := &RunOptions{}
	for _, o := range opts {
		o(options)
	}
	r.runOnce.Do(func() {
		r.started = true
		descriptors := r.registry.GetDescriptors(skrManager)
		for _, descr := range descriptors {
			logger2 := skrManager.GetLogger().WithValues("controller", descr.Name)
			ctrl, err := controller.New(descr.Name, skrManager, r.GetControllerOptions(
				descr,
				skrManager,
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
		timeoutCtx, cancelInternal := context.WithTimeout(ctx, options.timeout)
		var cancelOnce sync.Once
		cancel := func() {
			cancelOnce.Do(cancelInternal)
		}
		defer cancel()

		go r.queueMonitor(timeoutCtx, descriptors, cancel)

		err := skrManager.Start(timeoutCtx)
		r.stopped = true
		if err != nil {
			skrManager.GetLogger().Error(err, "error starting manager")
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
		Reconciler: descriptor.ReconcilerFactory.New(skrManager.KymaRef(), r.kcpCluster, skrManager),
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

func (r *skrRunner) queueMonitor(ctx context.Context, descriptors registry.DescriptorList, cancel func()) {
	whereAllEmptyBefore := false
	allEmptyNow := false
	for !r.stopped {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(10 * time.Second)

			allEmptyNow = descriptors.AllQueuesEmpty()
			if allEmptyNow && whereAllEmptyBefore {
				cancel()
				return
			}
			whereAllEmptyBefore = allEmptyNow
		}
	}
}
