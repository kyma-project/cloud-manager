package looper

import (
	"context"
	"errors"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/config"
	skrmanager "github.com/kyma-project/cloud-manager/pkg/skr/runtime/manager"
	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/registry"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sync"
	"time"
)

type RunOptions struct {
	timeout           time.Duration
	checkSkrReadiness bool
	provider          *cloudcontrolv1beta1.ProviderType
}

type RunOption = func(options *RunOptions)

func WithProvider(provider cloudcontrolv1beta1.ProviderType) RunOption {
	return func(options *RunOptions) {
		options.checkSkrReadiness = true
		options.provider = &provider
	}
}

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

func (r *skrRunner) isObjectActiveForProvider(scheme *runtime.Scheme, provider *cloudcontrolv1beta1.ProviderType, obj client.Object) bool {
	if provider == nil {
		return true
	}
	return common.ObjSupportsProvider(obj, scheme, string(*provider))
}

func (r *skrRunner) Run(ctx context.Context, skrManager skrmanager.SkrManager, opts ...RunOption) (err error) {
	if r.started {
		return errors.New("already started")
	}
	logger := skrManager.GetLogger()
	logger = feature.DecorateLogger(ctx, logger)
	logger.Info("SKR Runner running")
	options := &RunOptions{}
	for _, o := range opts {
		o(options)
	}
	r.runOnce.Do(func() {
		r.started = true
		defer func() {
			r.stopped = true
		}()

		if options.checkSkrReadiness {
			chkr := &checker{logger: logger}
			if !chkr.IsReady(ctx, skrManager) {
				logger.Info("SKR cluster is not ready")
				return
			}
		}

		if options.provider != nil {
			logger.Info(fmt.Sprintf("This SKR cluster is started with provider option %s", *options.provider))
			instlr := &installer{
				skrProvidersPath: config.SkrRuntimeConfig.ProvidersDir,
				scheme:           skrManager.GetScheme(),
				logger:           logger,
			}
			err = instlr.Handle(ctx, string(*options.provider), skrManager)
			if err != nil {
				err = fmt.Errorf("installer error: %w", err)
				logger.
					WithValues("provider", options.provider).
					Error(err, "Error installing dependencies")
				return
			}
		}

		rArgs := reconcile2.ReconcilerArguments{
			KymaRef:    skrManager.KymaRef(),
			KcpCluster: r.kcpCluster,
			SkrCluster: skrManager,
			Provider:   options.provider,
		}

		for _, indexer := range r.registry.Indexers() {
			ctx := feature.ContextBuilderFromCtx(ctx).
				Provider(util.CastInterfaceToString(options.provider)).
				KindsFromObject(indexer.Obj(), skrManager.GetScheme()).
				FeatureFromObject(indexer.Obj(), skrManager.GetScheme()).
				Build(ctx)
			logger := feature.DecorateLogger(ctx, logger)

			if r.isObjectActiveForProvider(skrManager.GetScheme(), options.provider, indexer.Obj()) &&
				!feature.ApiDisabled.Value(ctx) {
				err = indexer.IndexField(ctx, skrManager.GetFieldIndexer())
				if err != nil {
					err = fmt.Errorf("index filed error for %T: %w", indexer.Obj(), err)
					return
				}
				logger.
					WithValues(
						"object", fmt.Sprintf("%T", indexer.Obj()),
						"field", indexer.Field(),
						"provider", ptr.Deref(options.provider, ""),
					).
					Info("Starting indexer")
			} else {
				logger.
					WithValues(
						"object", fmt.Sprintf("%T", indexer.Obj()),
						"field", indexer.Field(),
						"provider", string(*options.provider),
					).
					Info("Not creating indexer due to disabled API")
			}
		}

		for _, b := range r.registry.Builders() {
			ctx := feature.ContextBuilderFromCtx(ctx).
				Provider(util.CastInterfaceToString(options.provider)).
				KindsFromObject(b.GetForObj(), skrManager.GetScheme()).
				FeatureFromObject(b.GetForObj(), skrManager.GetScheme()).
				Build(ctx)
			logger := feature.DecorateLogger(ctx, logger)

			if r.isObjectActiveForProvider(skrManager.GetScheme(), options.provider, b.GetForObj()) &&
				!feature.ApiDisabled.Value(ctx) {
				err = b.SetupWithManager(skrManager, rArgs)
				if err != nil {
					err = fmt.Errorf("setup with manager error for %T: %w", b.GetForObj(), err)
					return
				}
				logger.
					WithValues(
						"object", fmt.Sprintf("%T", b.GetForObj()),
						"provider", ptr.Deref(options.provider, ""),
					).
					Info("Starting controller")
			} else {
				logger.
					WithValues(
						"object", fmt.Sprintf("%T", b.GetForObj()),
						"provider", string(*options.provider),
					).
					Info("Not starting controller due to disabled API")
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
		if err != nil {
			skrManager.GetLogger().Error(err, "error starting SKR manager")
		}
	})
	return
}
