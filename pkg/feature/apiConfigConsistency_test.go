package feature

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// capturingLogger returns a logr.Logger whose every entry is appended to *entries as
// a decoded JSON map (carrying "msg" and, for Error entries, "error").
func capturingLogger(entries *[]map[string]any) logr.Logger {
	return funcr.NewJSON(func(obj string) {
		data := map[string]any{}
		_ = json.Unmarshal([]byte(obj), &data)
		*entries = append(*entries, data)
	}, funcr.Options{})
}

func msgs(entries []map[string]any) []any {
	var out []any
	for _, e := range entries {
		out = append(out, e["msg"])
	}
	return out
}

// TestReportApiConfigFallback_RetrieverError: a retriever error emits the load-failure
// signal with the retriever error string attached.
func TestReportApiConfigFallback_RetrieverError(t *testing.T) {
	var logs []map[string]any
	retrieveErr := errors.New("open /does/not/exist: no such file")

	reportApiConfigFallback(capturingLogger(&logs), nil, retrieveErr)

	require.Len(t, logs, 1)
	assert.Equal(t, SignalConfigLoadFailure, logs[0]["msg"])
	assert.Contains(t, logs[0]["error"], "no such file",
		"the retriever error must be attached to the signal")
}

// TestReportApiConfigFallback_ParseFailure: a present-but-unparseable config emits the
// parse-failure signal, not the load-failure or inconsistency signal.
func TestReportApiConfigFallback_ParseFailure(t *testing.T) {
	var logs []map[string]any
	malformed, err := os.ReadFile("testdata/apiMalformed.yaml")
	require.NoError(t, err)

	reportApiConfigFallback(capturingLogger(&logs), malformed, nil)

	require.Len(t, logs, 1)
	assert.Equal(t, SignalConfigParseFailure, logs[0]["msg"])
	assert.Contains(t, logs[0]["error"], "yaml:", "parse-failure signal must carry the YAML parse error")
}

// TestReportApiConfigFallback_ApiDisabledOnly: a config with apiDisabled rules but no
// apiEnabled block emits the inconsistency signal with ErrApiConfigInconsistent attached.
func TestReportApiConfigFallback_ApiDisabledOnly(t *testing.T) {
	var logs []map[string]any
	cfg, err := os.ReadFile("testdata/apiEnabledNoBlock.yaml")
	require.NoError(t, err)

	reportApiConfigFallback(capturingLogger(&logs), cfg, nil)

	require.Len(t, logs, 1)
	assert.Equal(t, SignalApiEnabledMissing, logs[0]["msg"])
	assert.Contains(t, logs[0]["error"], ErrApiConfigInconsistent.Error())
}

// TestReportApiConfigFallback_CleanConfig: a config carrying both blocks emits no signal.
func TestReportApiConfigFallback_CleanConfig(t *testing.T) {
	var logs []map[string]any
	cfg, err := os.ReadFile("testdata/apiEnabled.yaml")
	require.NoError(t, err)

	reportApiConfigFallback(capturingLogger(&logs), cfg, nil)

	assert.Empty(t, logs, "a config with both apiEnabled and apiDisabled must not signal: %v", msgs(logs))
}

// TestReportApiConfigFallback_EmptyConfigNoSignal: an empty {} config has no apiDisabled
// block, so it is not flagged as inconsistent (the fail-closed behavior itself is asserted
// in TestApiEnabledEmptyConfigFailsClosed).
func TestReportApiConfigFallback_EmptyConfigNoSignal(t *testing.T) {
	var logs []map[string]any
	reportApiConfigFallback(capturingLogger(&logs), []byte("{}"), nil)
	assert.Empty(t, logs)
}

// TestReportApiConfigFallback_EmptyApiEnabledBlock: a present-but-empty `apiEnabled:`
// key (carries no rule) alongside a content-bearing apiDisabled block is still flagged —
// untagged APIs go dark just as if apiEnabled were omitted entirely.
func TestReportApiConfigFallback_EmptyApiEnabledBlock(t *testing.T) {
	var logs []map[string]any
	cfg := []byte("apiEnabled:\napiDisabled:\n  defaultRule:\n    variation: enabled\n")

	reportApiConfigFallback(capturingLogger(&logs), cfg, nil)

	require.Len(t, logs, 1)
	assert.Equal(t, SignalApiEnabledMissing, logs[0]["msg"],
		"a present-but-empty apiEnabled block must still be flagged")
}

// TestApiEnabledEmptyConfigFailsClosed: loading an empty {} config disables non-bootstrap
// APIs while the cloudresources bootstrap kind-group stays enabled.
func TestApiEnabledEmptyConfigFailsClosed(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)

	require.NoError(t, Initialize(ctx, logr.Discard(), WithFile("testdata/apiEmpty.yaml")))

	sch := commonscheme.SkrScheme
	nonBootstrap := ContextBuilderFromCtx(ctx).Feature(types.FeatureNfsBackup).Build(ctx)
	bootstrap := ContextBuilderFromCtx(ctx).KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx)

	assert.False(t, ApiEnabled.Value(nonBootstrap), "empty config must disable a non-bootstrap API")
	assert.True(t, ApiEnabled.Value(bootstrap), "the cloudresources bootstrap kind-group must stay enabled")
}

// TestApiEnabledRetrieverErrorFailsClosed: a retriever error (missing file) does not abort
// startup; the provider serves the secure default — non-bootstrap disabled, bootstrap
// enabled — and the load-failure signal is emitted.
func TestApiEnabledRetrieverErrorFailsClosed(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)

	var logs []map[string]any
	require.NoError(t, Initialize(ctx, capturingLogger(&logs), WithFile("testdata/doesNotExist.yaml")))

	sch := commonscheme.SkrScheme
	nonBootstrap := ContextBuilderFromCtx(ctx).Feature(types.FeatureNfsBackup).Build(ctx)
	bootstrap := ContextBuilderFromCtx(ctx).KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx)

	assert.False(t, ApiEnabled.Value(nonBootstrap), "after a retriever error a non-bootstrap API must be disabled")
	assert.True(t, ApiEnabled.Value(bootstrap), "the cloudresources bootstrap kind-group must stay enabled")
	entry := findLog(logs, SignalConfigLoadFailure)
	require.NotNil(t, entry, "a retriever error must emit the load-failure signal; got %v", msgs(logs))
	assert.Contains(t, entry["error"], "doesNotExist",
		"the retriever error must be attached to the signal (proves Initialize propagates it)")
}

// findLog returns the first captured entry whose "msg" equals msg, or nil.
func findLog(entries []map[string]any, msg string) map[string]any {
	for _, e := range entries {
		if e["msg"] == msg {
			return e
		}
	}
	return nil
}

// restoreProvider captures the package provider before a test mutates it (via Initialize)
// and restores the exact original in cleanup, so loading an empty/broken config does not
// leak that state — or change the provider's concrete type — for later package tests.
func restoreProvider(t *testing.T) {
	orig := provider
	t.Cleanup(func() { provider = orig })
}
