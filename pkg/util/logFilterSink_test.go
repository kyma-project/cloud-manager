package util

import (
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
)

// recordingCallDepthSink implements logr.CallDepthLogSink and records the depth it
// was asked to climb plus whether an entry reached it.
type recordingCallDepthSink struct {
	recordedDepth int
	infoCount     int
}

func (s *recordingCallDepthSink) Init(logr.RuntimeInfo)            {}
func (s *recordingCallDepthSink) Enabled(int) bool                 { return true }
func (s *recordingCallDepthSink) Info(int, string, ...any)         { s.infoCount++ }
func (s *recordingCallDepthSink) Error(error, string, ...any)      {}
func (s *recordingCallDepthSink) WithValues(...any) logr.LogSink   { return s }
func (s *recordingCallDepthSink) WithName(string) logr.LogSink     { return s }
func (s *recordingCallDepthSink) WithCallDepth(d int) logr.LogSink { s.recordedDepth = d; return s }

var _ logr.CallDepthLogSink = &recordingCallDepthSink{}

// plainSink implements logr.LogSink but NOT logr.CallDepthLogSink.
type plainSink struct {
	infoCount  int
	errorCount int
}

func (s *plainSink) Init(logr.RuntimeInfo)          {}
func (s *plainSink) Enabled(int) bool               { return true }
func (s *plainSink) Info(int, string, ...any)       { s.infoCount++ }
func (s *plainSink) Error(error, string, ...any)    { s.errorCount++ }
func (s *plainSink) WithValues(...any) logr.LogSink { return s }
func (s *plainSink) WithName(string) logr.LogSink   { return s }

func TestLogFilterSinkForwardsCallDepth(t *testing.T) {
	inner := &recordingCallDepthSink{}
	logger := logr.New(NewLogFilterSink(inner))

	logger.WithCallDepth(3).Info("hello")

	assert.Equal(t, 3, inner.recordedDepth, "WithCallDepth must be forwarded to the inner sink")
	assert.Equal(t, 1, inner.infoCount, "the entry must still be emitted")
}

func TestLogFilterSinkCallDepthGracefulWhenInnerUnsupported(t *testing.T) {
	inner := &plainSink{}
	logger := logr.New(NewLogFilterSink(inner))

	assert.NotPanics(t, func() {
		logger.WithCallDepth(2).Info("hello")
	}, "WithCallDepth must not panic when the inner sink lacks CallDepthLogSink")
	assert.Equal(t, 1, inner.infoCount, "the entry must still be emitted")
}

func TestLogFilterSinkSuppressesNoiseMessages(t *testing.T) {
	suppressed := []string{
		"Starting workers", // literal
		"Failed to watch *v1beta1.X: context deadline exceeded",   // regex: deadline
		"watch ended with: ... context canceled",                  // regex: canceled
		"Timeout: failed waiting for *v1beta1.X Informer to sync", // regex: informer sync
	}
	for _, msg := range suppressed {
		inner := &plainSink{}
		logr.New(NewLogFilterSink(inner)).Info(msg)
		assert.Zerof(t, inner.infoCount, "message %q must be suppressed", msg)
	}

	inner := &plainSink{}
	logr.New(NewLogFilterSink(inner)).Info("an ordinary message")
	assert.Equal(t, 1, inner.infoCount, "ordinary messages must pass through")
}

func TestLogFilterSinkErrorFiltersByErrorText(t *testing.T) {
	inner := &plainSink{}
	logger := logr.New(NewLogFilterSink(inner))

	// The message is fine but the error text matches a suppression pattern.
	logger.Error(errors.New("dial tcp: context deadline exceeded"), "some operation")
	assert.Zero(t, inner.errorCount, "an error whose text is suppressed must be dropped")

	logger.Error(errors.New("real failure"), "some operation")
	assert.Equal(t, 1, inner.errorCount, "an ordinary error must pass through")
}
