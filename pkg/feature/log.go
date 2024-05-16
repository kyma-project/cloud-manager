package feature

import (
	"context"
	"github.com/go-logr/logr"
)

func DecorateLogger(ctx context.Context, logger logr.Logger) logr.Logger {
	ffCtx := ContextFromCtx(ctx)
	if ffCtx == nil {
		return logger
	}
	return attributesToLogger(ffCtx.GetCustom(), logger)
}

func attributesToLogger(attributes map[string]interface{}, logger logr.Logger) logr.Logger {
	var loggerValues []any
	for _, lk := range loggerKeys {
		v, ok := attributes[lk]
		if ok {
			loggerValues = append(loggerValues, lk, v)
		}
	}
	if len(loggerValues) > 0 {
		newLogger := logger.WithValues(loggerValues...)
		return newLogger
	}
	return logger
}
