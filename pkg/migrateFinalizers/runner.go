package migrateFinalizers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type runner interface {
	Run(ctx context.Context, options *RunnerOptions) error
}

type KindInfo struct {
	Title     string
	List      client.ObjectList
	Namespace string
}

type RunnerOptions struct {
	*finalizerInfo
	Client        client.Client
	Logger        logr.Logger
	KindsProvider kindInfoProvider
	Migrator      migrator
}

func newRunnerOptions(clnt client.Client, logger logr.Logger, kindsProvider kindInfoProvider, migrator migrator) *RunnerOptions {
	return &RunnerOptions{
		finalizerInfo: newFinalizerInfo(),
		Client:        clnt,
		Logger:        logger,
		KindsProvider: kindsProvider,
		Migrator:      migrator,
	}
}

func newRunner() runner {
	return &defaultRunner{}
}

type defaultRunner struct{}

func (r *defaultRunner) Run(ctx context.Context, options *RunnerOptions) error {
	for _, kind := range options.KindsProvider() {
		logger := options.Logger.
			WithValues(
				"kind", kind.Title,
			)
		mo := &migratorOptions{
			finalizerInfo: options.finalizerInfo,
			Client:        options.Client,
			List:          kind.List,
			Logger:        logger,
			Namespace:     kind.Namespace,
		}
		err := options.Migrator.Migrate(ctx, mo)
		if err != nil {
			return fmt.Errorf("error running kind %s: %w", kind.Title, err)
		}
	}

	return nil
}
