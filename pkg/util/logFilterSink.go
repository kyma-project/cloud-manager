package util

import (
	"github.com/go-logr/logr"
	"regexp"
)

type LogFilterSink struct {
	inner                          logr.LogSink
	unhandledErrorDeadlineExceeded *regexp.Regexp
	contextCanceledReflected       *regexp.Regexp
}

func NewLogFilterSink(inner logr.LogSink) *LogFilterSink {
	return &LogFilterSink{
		inner:                          inner,
		unhandledErrorDeadlineExceeded: regexp.MustCompile(`"Unhandled Error" err="pkg/mod/k8s\.io/client-go@v0\.\d+\.\d+/tools/cache/reflector\.go:\d+: Failed to watch \*v1beta1\.[a-zA-Z]+: context deadline exceeded" logger="UnhandledError"`),
		contextCanceledReflected:       regexp.MustCompile(`reflector.go:\d+] pkg/mod/k8s\.io/client-go@v0\.\d+\.\d+/tools/cache/reflector.go:\d+: watch of \*v1beta1\.[a-zA-Z0..9]+ ended with: an error on the server \("unable to decode an event from the watch stream: context canceled"\) has prevented the request from succeeding`),
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
	if l.isMsgOk(msg) {
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
