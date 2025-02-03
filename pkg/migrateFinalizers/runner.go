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
	reader        client.Reader
	writer        client.Writer
	Logger        logr.Logger
	KindsProvider kindInfoProvider
	Migrator      migrator
}

func newRunnerOptions(reader client.Reader, writer client.Writer, logger logr.Logger, kindsProvider kindInfoProvider, migrator migrator) *RunnerOptions {
	return &RunnerOptions{
		finalizerInfo: newFinalizerInfo(),
		reader:        reader,
		writer:        writer,
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
			reader:        options.reader,
			writer:        options.writer,
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
