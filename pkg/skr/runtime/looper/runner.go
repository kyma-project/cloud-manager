package looper

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
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
)

type RunOptions struct {
	timeout           time.Duration
	checkSkrReadiness bool
	provider          *cloudcontrolv1beta1.ProviderType
}

type RunOption = func(options *RunOptions)

func WithoutProvider() RunOption {
	return func(options *RunOptions) {
		options.provider = nil
	}
}

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

func NewSkrRunnerWithNoopStatusSaver(reg registry.SkrRegistry, kcpCluster cluster.Cluster, kymaName string) SkrRunner {
	return NewSkrRunner(reg, kcpCluster, NewNoopStatusSaver(), kymaName)
}

func NewSkrRunner(reg registry.SkrRegistry, kcpCluster cluster.Cluster, skrStatusSaver SkrStatusSaver, kymaName string) SkrRunner {
	return &skrRunner{
		kcpCluster:     kcpCluster,
		registry:       reg,
		skrStatusSaver: skrStatusSaver,
		kymaName:       kymaName,
	}
}

type skrRunner struct {
	kcpCluster     cluster.Cluster
	registry       registry.SkrRegistry
	skrStatusSaver SkrStatusSaver
	kymaName       string

	runOnce sync.Once
	started bool
	stopped bool
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
	//logger.Info("SKR Runner running")
	options := &RunOptions{}
	for _, o := range opts {
		o(options)
	}

	r.runOnce.Do(func() {
		r.started = true

		skrStatus := NewSkrStatus(ctx)

		defer func() {
			r.stopped = true
			r.saveSkrStatus(ctx, skrStatus, logger)
		}()

		if options.checkSkrReadiness {
			chkr := &checker{logger: logger}
			if !chkr.IsReady(ctx, skrManager) {
				logger.Info("SKR cluster is not ready")
				skrStatus.NotReady()
				return
			}
		}

		if options.provider != nil {
			//logger.Info(fmt.Sprintf("This SKR cluster is started with provider option %s", ptr.Deref(options.provider, "")))
			instlr := &installer{
				skrStatus:        skrStatus,
				skrProvidersPath: config.SkrRuntimeConfig.ProvidersDir,
				logger:           logger,
			}
			err = instlr.Handle(ctx, string(ptr.Deref(options.provider, "")), ToCluster(skrManager))
			if err != nil {
				err = fmt.Errorf("installer error: %w", err)
				logger.
					WithValues("optionsProvider", ptr.Deref(options.provider, "")).
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
			//logger := feature.DecorateLogger(ctx, logger)

			handle := skrStatus.Handle(ctx, "Indexer")
			handle.WithObj(indexer.Obj())

			if r.isObjectActiveForProvider(skrManager.GetScheme(), options.provider, indexer.Obj()) &&
				!feature.ApiDisabled.Value(ctx) {
				handle.Starting()
				err = indexer.IndexField(ctx, skrManager.GetFieldIndexer())
				if err != nil {
					handle.Error(err)
					err = fmt.Errorf("index filed error for %T: %w", indexer.Obj(), err)
					return
				}
				//logger.
				//	WithValues(
				//		"object", fmt.Sprintf("%T", indexer.Obj()),
				//		"field", indexer.Field(),
				//		"optionsProvider", ptr.Deref(options.provider, ""),
				//	).
				//	Info("Starting indexer")
			} else {
				handle.ApiDisabled()
				//logger.
				//	WithValues(
				//		"object", fmt.Sprintf("%T", indexer.Obj()),
				//		"field", indexer.Field(),
				//		"optionsProvider", ptr.Deref(options.provider, ""),
				//	).
				//	Info("Not creating indexer due to disabled API")
			}
		}

		for _, b := range r.registry.Builders() {
			ctx := feature.ContextBuilderFromCtx(ctx).
				Provider(util.CastInterfaceToString(options.provider)).
				KindsFromObject(b.GetForObj(), skrManager.GetScheme()).
				FeatureFromObject(b.GetForObj(), skrManager.GetScheme()).
				Build(ctx)
			//logger := feature.DecorateLogger(ctx, logger)

			handle := skrStatus.Handle(ctx, "Controller")

			if r.isObjectActiveForProvider(skrManager.GetScheme(), options.provider, b.GetForObj()) &&
				!feature.ApiDisabled.Value(ctx) {
				handle.Starting()
				err = b.SetupWithManager(skrManager, rArgs)
				if err != nil {
					handle.Error(err)
					err = fmt.Errorf("setup with manager error for %T: %w", b.GetForObj(), err)
					return
				}
				//logger.
				//	WithValues(
				//		"registryBuilderObject", fmt.Sprintf("%T", b.GetForObj()),
				//		"optionsProvider", ptr.Deref(options.provider, ""),
				//	).
				//	Info("Starting controller")
			} else {
				handle.ApiDisabled()
				//logger.
				//	WithValues(
				//		"registryBuilderObject", fmt.Sprintf("%T", b.GetForObj()),
				//		"optionsProvider", ptr.Deref(options.provider, ""),
				//	).
				//	Info("Not starting controller due to disabled API")
			}
		}

		skrStatus.Connected()

		// this is a happy path saving, all other places are covered with the called form defer
		// we want to save skrStatus here before manager is started and waited to timeout
		// since it tracks its save status with IsSaved, after this call to save when defer is
		// executed it will not save it again
		r.saveSkrStatus(ctx, skrStatus, logger)

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

func (r *skrRunner) saveSkrStatus(ctx context.Context, skrStatus *SkrStatus, logger logr.Logger) {
	// save it only once
	if skrStatus.IsSaved {
		return
	}

	err := r.skrStatusSaver.Save(ctx, skrStatus)
	if err != nil {
		logger.Error(err, "error saving SKR status")
	}
	skrStatus.IsSaved = true
}
