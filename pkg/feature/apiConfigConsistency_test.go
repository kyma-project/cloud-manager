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

func logMessages(entries []map[string]any) []any {
	var out []any
	for _, e := range entries {
		out = append(out, e["msg"])
	}
	return out
}

// TestReportLoadFailure_RetrieverError: the provider did not load and the re-read
// also errors — the load-failure signal carries the retriever error.
func TestReportLoadFailure_RetrieverError(t *testing.T) {
	var logs []map[string]any
	retrieveErr := errors.New("open /does/not/exist: no such file")

	reportLoadFailure(capturingLogger(&logs), nil, retrieveErr)

	require.Len(t, logs, 1)
	assert.Equal(t, SignalConfigLoadFailure, logs[0]["msg"])
	assert.Contains(t, logs[0]["error"], "no such file",
		"the retriever error must be attached to the signal")
}

// TestReportLoadFailure_ParseFailure: the provider did not load and the re-read
// returns present-but-unparseable bytes — the parse-failure signal fires, not load-failure.
func TestReportLoadFailure_ParseFailure(t *testing.T) {
	var logs []map[string]any
	malformed, err := os.ReadFile("testdata/apiMalformed.yaml")
	require.NoError(t, err)

	reportLoadFailure(capturingLogger(&logs), malformed, nil)

	require.Len(t, logs, 1)
	assert.Equal(t, SignalConfigParseFailure, logs[0]["msg"])
	assert.Contains(t, logs[0]["error"], "yaml:", "parse-failure signal must carry the YAML parse error")
}

// TestReportLoadFailure_ProviderNotLoadedButParseable: the provider did not load even
// though the config is fetchable and parseable to us (e.g. GoFF schema rejection) — this is
// still a fallback, so the load-failure signal fires.
func TestReportLoadFailure_ProviderNotLoadedButParseable(t *testing.T) {
	var logs []map[string]any
	cfg, err := os.ReadFile("testdata/apiEnabled.yaml")
	require.NoError(t, err)

	reportLoadFailure(capturingLogger(&logs), cfg, nil)

	require.Len(t, logs, 1)
	assert.Equal(t, SignalConfigLoadFailure, logs[0]["msg"],
		"a provider that did not load a parseable config has still fallen back")
}

// TestReportApiEnabledAbsent_NoBlock: a loaded config with no apiEnabled block
// emits the missing signal with ErrApiEnabledAbsent attached.
func TestReportApiEnabledAbsent_NoBlock(t *testing.T) {
	var logs []map[string]any
	cfg, err := os.ReadFile("testdata/apiEnabledNoBlock.yaml")
	require.NoError(t, err)

	reportApiEnabledAbsent(capturingLogger(&logs), cfg, nil)

	require.Len(t, logs, 1)
	assert.Equal(t, SignalApiEnabledAbsent, logs[0]["msg"])
	assert.Contains(t, logs[0]["error"], ErrApiEnabledAbsent.Error())
}

// TestReportApiEnabledAbsent_CleanConfig: a loaded config with an apiEnabled block emits no signal.
func TestReportApiEnabledAbsent_CleanConfig(t *testing.T) {
	var logs []map[string]any
	cfg, err := os.ReadFile("testdata/apiEnabled.yaml")
	require.NoError(t, err)

	reportApiEnabledAbsent(capturingLogger(&logs), cfg, nil)

	assert.Empty(t, logs, "a config with apiEnabled block must not signal: %v", logMessages(logs))
}

// TestReportApiEnabledAbsent_TransientRereadError: the provider loaded fine,
// but the best-effort re-read blips (e.g. a transient HTTP error). No signal must fire — the
// provider is serving the loaded config, so this is not a fallback.
func TestReportApiEnabledAbsent_TransientRereadError(t *testing.T) {
	var logs []map[string]any
	reportApiEnabledAbsent(capturingLogger(&logs), nil, errors.New("transient network blip"))
	assert.Empty(t, logs, "a re-read blip while the provider is loaded must not raise a false alarm: %v", logMessages(logs))
}

// TestReportApiEnabledAbsent_MalformedRereadStaysSilent: the provider loaded fine, but the
// best-effort re-read returns unparseable bytes. No signal must fire — a malformed re-read is
// not a fallback (the provider is still serving the loaded config). Armors the silent-return
// contract on the providerLoaded=true + retrieveErr=nil + malformed-bytes path.
func TestReportApiEnabledAbsent_MalformedRereadStaysSilent(t *testing.T) {
	var logs []map[string]any
	malformed, err := os.ReadFile("testdata/apiMalformed.yaml")
	require.NoError(t, err)

	reportApiEnabledAbsent(capturingLogger(&logs), malformed, nil)
	assert.Empty(t, logs, "a malformed re-read while the provider is loaded must stay silent: %v", logMessages(logs))
}

// TestReportApiEnabledAbsent_EmptyConfig: a loaded empty {} config has no
// apiEnabled block so the missing signal fires (fail-safe: ops is alerted that nothing is
// explicitly enabled).
func TestReportApiEnabledAbsent_EmptyConfig(t *testing.T) {
	var logs []map[string]any
	reportApiEnabledAbsent(capturingLogger(&logs), []byte("{}"), nil)
	require.Len(t, logs, 1)
	assert.Equal(t, SignalApiEnabledAbsent, logs[0]["msg"])
	assert.Contains(t, logs[0]["error"], ErrApiEnabledAbsent.Error())
}

// TestReportApiEnabledAbsent_EmptyBlock: a present-but-empty `apiEnabled:`
// key (carries no rule) is still flagged — untagged APIs go dark just as if apiEnabled
// were omitted entirely.
func TestReportApiEnabledAbsent_EmptyBlock(t *testing.T) {
	var logs []map[string]any
	cfg := []byte("apiEnabled:\n")

	reportApiEnabledAbsent(capturingLogger(&logs), cfg, nil)

	require.Len(t, logs, 1)
	assert.Equal(t, SignalApiEnabledAbsent, logs[0]["msg"],
		"a present-but-empty apiEnabled block must still be flagged")
	assert.Contains(t, logs[0]["error"], ErrApiEnabledAbsent.Error(),
		"a present-but-empty apiEnabled block must attach ErrApiEnabledAbsent")
}

// TestApiEnabledEmptyConfigFailsClosed: loading an empty {} config disables non-bootstrap
// APIs while the cloudresources bootstrap kind-group, FeatureFoundation, and FeatureUnknown
// stay enabled (always-on hardcode). The SignalApiEnabledAbsent signal is emitted.
func TestApiEnabledEmptyConfigFailsClosed(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)

	var logs []map[string]any
	require.NoError(t, Initialize(ctx, capturingLogger(&logs), WithFile("testdata/apiEmpty.yaml")))

	sch := commonscheme.SkrScheme
	nonBootstrap := ContextBuilderFromCtx(ctx).Feature(types.FeatureNfsBackup).Build(ctx)
	bootstrap := ContextBuilderFromCtx(ctx).KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx)
	foundation := ContextBuilderFromCtx(ctx).Feature(types.FeatureFoundation).Build(ctx)
	unknown := ContextBuilderFromCtx(ctx).Feature(types.FeatureUnknown).Build(ctx)

	assert.False(t, ApiEnabled.Value(nonBootstrap), "empty config must disable a non-bootstrap API")
	assert.True(t, ApiEnabled.Value(bootstrap), "the cloudresources bootstrap kind-group must stay enabled")
	assert.True(t, ApiEnabled.Value(foundation), "FeatureFoundation must stay always-on even with no apiEnabled block")
	assert.True(t, ApiEnabled.Value(unknown), "FeatureUnknown must stay always-on even with no apiEnabled block")
	require.NotNil(t, findLog(logs, SignalApiEnabledAbsent),
		"an empty config must emit the missing signal; got %v", logMessages(logs))
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
	require.NotNil(t, entry, "a retriever error must emit the load-failure signal; got %v", logMessages(logs))
	assert.Contains(t, entry["error"], "doesNotExist",
		"the retriever error must be attached to the signal (proves Initialize propagates it)")
}

// TestApiEnabledMalformedConfigFailsClosed: a present-but-unparseable config does not abort
// startup; the provider serves the secure default — non-bootstrap disabled, bootstrap enabled —
// and the parse-failure signal is emitted. This closes the gap between "parse-failure was
// signalled" and "the gate is actually closed" for a malformed config.
func TestApiEnabledMalformedConfigFailsClosed(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)

	var logs []map[string]any
	require.NoError(t, Initialize(ctx, capturingLogger(&logs), WithFile("testdata/apiMalformed.yaml")))

	sch := commonscheme.SkrScheme
	nonBootstrap := ContextBuilderFromCtx(ctx).Feature(types.FeatureNfsBackup).Build(ctx)
	bootstrap := ContextBuilderFromCtx(ctx).KindsFromObject(&cloudresourcesv1beta1.CloudResources{}, sch).Build(ctx)

	assert.False(t, ApiEnabled.Value(nonBootstrap), "a malformed config must disable a non-bootstrap API")
	assert.True(t, ApiEnabled.Value(bootstrap), "the cloudresources bootstrap kind-group must stay enabled")
	require.NotNil(t, findLog(logs, SignalConfigParseFailure),
		"a malformed config must emit the parse-failure signal; got %v", logMessages(logs))
}

// TestApiEnabledValidConfigEmitsNoFallbackSignal: a successfully-loaded, well-formed config
// emits none of the fail-closed signals — proving the provider's non-zero cache-refresh date
// after a good load is not mistaken for a load failure (the success direction the error-path
// tests above do not cover).
func TestApiEnabledValidConfigEmitsNoFallbackSignal(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)

	var logs []map[string]any
	require.NoError(t, Initialize(ctx, capturingLogger(&logs), WithFile("testdata/apiEnabled.yaml")))

	assert.Nil(t, findLog(logs, SignalConfigLoadFailure), "a valid config must not signal load failure: %v", logMessages(logs))
	assert.Nil(t, findLog(logs, SignalConfigParseFailure), "a valid config must not signal parse failure: %v", logMessages(logs))
	assert.Nil(t, findLog(logs, SignalApiEnabledAbsent), "a valid config with apiEnabled block must not signal missing: %v", logMessages(logs))
}

// TestSignalConstantsPairwiseDistinct verifies that the three fail-closed signal message
// constants are all distinct so ops alerting rules that key on exact message text cannot
// accidentally conflate them.
//
// Oracle: each signal maps to a distinct ops alert (a collapsed pair would mis-route the
// underlying condition), so distinctness is an external operational requirement — not merely
// a restatement of the constant definitions.
func TestSignalConstantsPairwiseDistinct(t *testing.T) {
	assert.NotEqual(t, SignalConfigLoadFailure, SignalConfigParseFailure,
		"SignalConfigLoadFailure and SignalConfigParseFailure must be distinct")
	assert.NotEqual(t, SignalConfigLoadFailure, SignalApiEnabledAbsent,
		"SignalConfigLoadFailure and SignalApiEnabledAbsent must be distinct")
	assert.NotEqual(t, SignalConfigParseFailure, SignalApiEnabledAbsent,
		"SignalConfigParseFailure and SignalApiEnabledAbsent must be distinct")
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
