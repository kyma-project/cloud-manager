package util

import (
	"regexp"

	"github.com/go-logr/logr"
)

type LogFilterSink struct {
	inner                          logr.LogSink
	unhandledErrorDeadlineExceeded *regexp.Regexp
	contextCanceledReflected       *regexp.Regexp
	failedWaitingForInformerToSync *regexp.Regexp
}

func NewLogFilterSink(inner logr.LogSink) *LogFilterSink {
	return &LogFilterSink{
		inner:                          inner,
		unhandledErrorDeadlineExceeded: regexp.MustCompile(`context deadline exceeded`),
		contextCanceledReflected:       regexp.MustCompile(`context canceled`),
		failedWaitingForInformerToSync: regexp.MustCompile(`Timeout: failed waiting for .* Informer to sync`),
	}
}

func (l *LogFilterSink) isMsgOk(msg string) bool {
	switch msg {
	case "Starting workers",
		"Starting EventSource",
		"Starting Controller",
		"Shutdown signal received, waiting for all workers to finish",
		"All workers finished":
		return false
	}
	if l.failedWaitingForInformerToSync.MatchString(msg) {
		return false
	}
	// "Unhandled Error" err="pkg/mod/k8s.io/client-go@v0.32.0/tools/cache/reflector.go:251: Failed to watch *v1beta1.AzureVpcPeering: context deadline exceeded" logger="UnhandledError"
	if l.unhandledErrorDeadlineExceeded.MatchString(msg) {
		return false
	}
	// reflector.go:492] pkg/mod/k8s.io/client-go@v0.33.0/tools/cache/reflector.go:251: watch of *v1beta1.Nukase ended with: an error on the server ("unable to decode an event from the watch stream: context canceled") has prevented the request from succeeding
	if l.contextCanceledReflected.MatchString(msg) {
		return false
	}
	return true
}

func (l *LogFilterSink) Init(info logr.RuntimeInfo) {
	l.inner.Init(info)
}

func (l *LogFilterSink) Enabled(level int) bool {
	return l.inner.Enabled(level)
}

func (l *LogFilterSink) Info(level int, msg string, keysAndValues ...any) {
	if l.isMsgOk(msg) {
		l.inner.Info(level, msg, keysAndValues...)
	}
}

func (l *LogFilterSink) Error(err error, msg string, keysAndValues ...any) {
	if l.isMsgOk(msg) && (err == nil || l.isMsgOk(err.Error())) {
		l.inner.Error(err, msg, keysAndValues...)
	}
}

func (l *LogFilterSink) WithValues(keysAndValues ...any) logr.LogSink {
	newLogger := *l
	newLogger.inner = l.inner.WithValues(keysAndValues...)
	return &newLogger
}

func (l *LogFilterSink) WithName(name string) logr.LogSink {
	newLogger := *l
	newLogger.inner = l.inner.WithName(name)
	return &newLogger
}
