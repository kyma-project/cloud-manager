package composed

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func LoggerFromCtx(ctx context.Context) logr.Logger {
	return log.FromContext(ctx)
}

func LoggerIntoCtx(ctx context.Context, logger logr.Logger) context.Context {
	newCtx := log.IntoContext(ctx, logger)

	return newCtx
}

func LogErrorAndReturn(err error, msg string, result error, ctx context.Context) (error, context.Context) {
	logger := LoggerFromCtx(ctx)
	if ctx == nil {
		logger.Error(errors.New("the ctx is not supplied to LogErrorAndReturn"), "Logical error")
	}
	logger.Error(err, msg)
	return result, ctx
}
