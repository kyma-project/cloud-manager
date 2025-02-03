package migrateFinalizers

import (
	"context"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Migration interface {
	Run(ctx context.Context) (alreadyExecuted bool, err error)
}

type migration struct {
	successHandler successHandler
	runner         runner
	kindsProvider  kindInfoProvider
	reader         client.Reader
	writer         client.Writer
	logger         logr.Logger
}

func (m *migration) Run(ctx context.Context) (alreadyExecuted bool, err error) {
	isRecorded, err := m.successHandler.IsRecorded(ctx)
	if err != nil {
		return false, err
	}
	if isRecorded {
		return true, nil
	}

	if err := m.runner.Run(ctx, newRunnerOptions(m.reader, m.writer, m.logger, m.kindsProvider, newMigrator())); err != nil {
		return false, err
	}

	err = m.successHandler.Record(ctx)
	if err != nil {
		return false, err
	}

	return false, nil
}

func NewMigrationForKcp(kcpReader client.Reader, kcpWriter client.Writer, logger logr.Logger) Migration {
	return &migration{
		successHandler: newKcpSuccessHandler(kcpNamespace, kcpReader, kcpWriter),
		runner:         newRunner(),
		kindsProvider:  newKindsForKcp,
		reader:         kcpReader,
		writer:         kcpWriter,
		logger:         logger,
	}
}

func NewMigrationForSkr(kymaName string, kcpReader client.Reader, kcpWriter client.Writer, skrReader client.Reader, skrWriter client.Writer, logger logr.Logger) Migration {
	return &migration{
		successHandler: newSkrSuccessHandler(kymaName, kcpNamespace, kcpReader, kcpWriter),
		runner:         newRunner(),
		kindsProvider:  newKindsForSkr,
		reader:         skrReader,
		writer:         skrWriter,
		logger:         logger,
	}
}
