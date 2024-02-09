package looper

import (
	"context"
	"errors"
	skrmanager "github.com/kyma-project/cloud-manager/pkg/skr/runtime/manager"
	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/registry"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
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
	Run(ctx context.Context, skrManager skrmanager.SkrManager, opts ...RunOption) error
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

func (r *skrRunner) Run(ctx context.Context, skrManager skrmanager.SkrManager, opts ...RunOption) (err error) {
	if r.started {
		return errors.New("already started")
	}
	logger := skrManager.GetLogger()
	logger.Info("Starting SKR Runner")
	options := &RunOptions{}
	for _, o := range opts {
		o(options)
	}
	r.runOnce.Do(func() {
		r.started = true
		rArgs := reconcile2.ReconcilerArguments{
			KymaRef:    skrManager.KymaRef(),
			KcpCluster: r.kcpCluster,
			SkrCluster: skrManager,
		}

		for _, b := range r.registry.Builders() {
			err = b.SetupWithManager(skrManager, rArgs)
			if err != nil {
				return
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

		err = skrManager.Start(timeoutCtx)
		r.stopped = true
		if err != nil {
			skrManager.GetLogger().Error(err, "error starting SKR manager")
		}
	})
	return
}
