package composed

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/go-logr/zapr"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// newObservedWarningCtx builds a logger whose chain mirrors production
// (zapr sink wrapped in util.LogFilterSink) over a zaptest observer, and returns
// a ctx carrying it plus the observed-log recorder.
func newObservedWarningCtx() (context.Context, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.DebugLevel)
	zapLog := zap.New(core, zap.AddCaller())
	logger := zapr.NewLogger(zapLog)
	logger = logger.WithSink(util.NewLogFilterSink(logger.GetSink()))
	return LoggerIntoCtx(context.Background(), logger), logs
}

func TestLogWarningEmitsWarnSeverity(t *testing.T) {
	ctx, logs := newObservedWarningCtx()

	LogWarning(ctx, "something odd", "objectKind", "NfsInstance", "count", 2)

	entries := logs.All()
	require.Len(t, entries, 1)
	assert.Equal(t, zapcore.WarnLevel, entries[0].Level, "must emit at WarnLevel (GCP WARNING)")
	assert.Equal(t, "something odd", entries[0].Message)
	ctxMap := entries[0].ContextMap()
	assert.Equal(t, "NfsInstance", ctxMap["objectKind"])
	assert.EqualValues(t, 2, ctxMap["count"])
}

func TestLogWarningAttributesToCallSite(t *testing.T) {
	ctx, logs := newObservedWarningCtx()

	LogWarning(ctx, "attributed") // <- caller should resolve to THIS line's file

	entries := logs.All()
	require.Len(t, entries, 1)
	caller := entries[0].Caller
	require.True(t, caller.Defined, "caller must be captured")
	assert.Equal(t, "logger_test.go", filepath.Base(caller.File),
		"caller must attribute to the call site, not logger.go/logFilterSink.go; got %q", caller.File)
}

func TestLogWarningRespectsMessageFilter(t *testing.T) {
	ctx, logs := newObservedWarningCtx()

	// "Starting workers" is a literal the LogFilterSink suppresses.
	LogWarning(ctx, "Starting workers")

	assert.Empty(t, logs.All(), "a suppressed message must produce no entry")
}

func TestLogWarningNilContextDoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		LogWarning(nil, "no ctx", "k", "v") //nolint:staticcheck // deliberately testing nil ctx
	})
}

func TestLogWarningLoggerlessContextDoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		LogWarning(context.Background(), "no logger set")
	})
}
