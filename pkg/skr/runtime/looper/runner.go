package looper

import (
	"context"
	"errors"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skrmanager "github.com/kyma-project/cloud-manager/pkg/skr/runtime/manager"
	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/registry"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sync"
	"time"
)

var (
	providerSpecificTypes = map[string]cloudcontrolv1beta1.ProviderType{
		// AWS
		fmt.Sprintf("%T", &cloudresourcesv1beta1.AwsNfsVolume{}): cloudcontrolv1beta1.ProviderAws,

		// GCP
		fmt.Sprintf("%T", &cloudresourcesv1beta1.GcpNfsVolume{}): cloudcontrolv1beta1.ProviderGCP,
	}
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

func (r *skrRunner) isObjectActive(provider *cloudcontrolv1beta1.ProviderType, obj client.Object) bool {
	if provider == nil {
		return true
	}
	pt, ptDefined := providerSpecificTypes[fmt.Sprintf("%T", obj)]
	if !ptDefined {
		return true
	}
	if pt == *provider {
		return true
	}
	return false
}

func (r *skrRunner) Run(ctx context.Context, skrManager skrmanager.SkrManager, opts ...RunOption) (err error) {
	if r.started {
		return errors.New("already started")
	}
	logger := skrManager.GetLogger()
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
				skrProvidersPath: os.Getenv("SKR_PROVIDERS"),
				logger:           logger,
			}
			err = instlr.Handle(ctx, string(*options.provider), skrManager)
			if err != nil {
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
			if r.isObjectActive(options.provider, indexer.Obj()) {
				err = indexer.IndexField(ctx, skrManager.GetFieldIndexer())
				if err != nil {
					return
				}
			} else {
				logger.
					WithValues(
						"object", fmt.Sprintf("%T", indexer.Obj()),
						"field", indexer.Field(),
						"provider", string(*options.provider),
					).
					Info("Not creating indexer of object due to non matching provider")
			}
		}

		for _, b := range r.registry.Builders() {
			if r.isObjectActive(options.provider, b.GetForObj()) {
				err = b.SetupWithManager(skrManager, rArgs)
				if err != nil {
					return
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

		err = skrManager.Start(timeoutCtx)
		if err != nil {
			skrManager.GetLogger().Error(err, "error starting SKR manager")
		}
	})
	return
}
