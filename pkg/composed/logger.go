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
	if ctx == nil {
		LoggerFromCtx(context.Background()).Error(errors.New("the ctx is not supplied to LogErrorAndReturn"), "Logical error")
		return result, ctx
	}
	logger := LoggerFromCtx(ctx)
	logger.Error(err, msg)
	return result, ctx
}

// warnLevel maps, via zapr's `0 - level` inversion, to zapcore.WarnLevel → "WARNING".
const warnLevel = -1

// LogWarning emits msg at GCP severity WARNING — the WARNING sibling of
// LogErrorAndReturn. Severity→helper mapping per logging policy #10191:
// INFO→Logger.Info, WARNING→LogWarning, ERROR→LogErrorAndReturn.
// go-logr's V() can't go negative, so WARNING is only reachable through the sink.
func LogWarning(ctx context.Context, msg string, keysAndValues ...any) {
	if ctx == nil {
		return
	}
	// WithCallDepth(1) skips this frame so `caller` points at the call site;
	// relies on LogFilterSink forwarding WithCallDepth. Pinned by TestLogWarningAttributesToCallSite.
	LoggerFromCtx(ctx).WithCallDepth(1).GetSink().Info(warnLevel, msg, keysAndValues...)
}
