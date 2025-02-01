package migrateFinalizers

import (
	"context"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Migration interface {
	Run(ctx context.Context) error
}

type migration struct {
	successHandler successHandler
	runner         runner
	kindsProvider  kindInfoProvider
	client         client.Client
	logger         logr.Logger
}

func (m *migration) Run(ctx context.Context) error {
	isRecorded, err := m.successHandler.IsRecorded(ctx)
	if err != nil {
		return err
	}
	if isRecorded {
		return nil
	}

	if err := m.runner.Run(ctx, newRunnerOptions(m.client, m.logger, m.kindsProvider, newMigrator())); err != nil {
		return err
	}

	err = m.successHandler.Record(ctx)
	if err != nil {
		return err
	}

	return nil
}

func NewMigrationForKcp(ctx context.Context, kcpClient client.Client, logger logr.Logger) Migration {
	return &migration{
		successHandler: newKcpSuccessHandler(kcpNamespace, kcpClient),
		runner:         newRunner(),
		kindsProvider:  newKindsForKcp,
		client:         kcpClient,
		logger:         logger,
	}
}

func NewMigrationForSkr(ctx context.Context, kymaName string, kcpClient client.Client, skrClient client.Client, logger logr.Logger) Migration {
	return &migration{
		successHandler: newSkrSuccessHandler(kymaName, kcpNamespace, kcpClient),
		runner:         newRunner(),
		kindsProvider:  newKindsForSkr,
		client:         skrClient,
		logger:         logger,
	}
}
