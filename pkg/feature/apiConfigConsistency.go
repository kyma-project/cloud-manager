package feature

import (
	"errors"

	"github.com/go-logr/logr"
)

// Fail-closed signal messages. Kept stable and greppable — ops alerting keys on them.
const (
	SignalConfigLoadFailure  = "feature-flag config load failed; falling back to secure default (non-bootstrap APIs disabled)"
	SignalConfigParseFailure = "feature-flag config present but unparseable; falling back to secure default (non-bootstrap APIs disabled)"
	SignalApiEnabledMissing  = "feature-flag config apiEnabled block absent or empty; all non-bootstrap APIs disabled as fail-safe"
)

// ErrApiEnabledAbsent is the cause attached to the apiEnabled-missing signal so
// the error-level log carries a non-nil error even though there is no underlying failure.
var ErrApiEnabledAbsent = errors.New("feature-flag config: apiEnabled block absent or empty")

// reportApiEnabledAbsent flags an absent or empty apiEnabled block in a config the provider
// loaded successfully. configBytes/retrieveErr are a best-effort re-read used only to inspect
// the live config; a re-read failure is not a fallback and must stay silent (the provider is
// serving the loaded config either way). It never aborts startup.
func reportApiEnabledAbsent(logger logr.Logger, configBytes []byte, retrieveErr error) {
	if retrieveErr != nil {
		return
	}
	root, err := parseConfigBytes(configBytes)
	if err == nil && !blockHasContent(root[apiEnabledFlagName]) {
		logger.Error(ErrApiEnabledAbsent, SignalApiEnabledMissing)
	}
}

// reportLoadFailure classifies why the provider fell back to the secure default (retriever
// error vs. parse error vs. provider-side rejection of a config we can parse) and emits the
// matching error-level signal. It never aborts startup — the caller keeps the secure-default
// provider either way.
func reportLoadFailure(logger logr.Logger, configBytes []byte, retrieveErr error) {
	if retrieveErr != nil {
		logger.Error(retrieveErr, SignalConfigLoadFailure)
		return
	}
	if _, err := parseConfigBytes(configBytes); err != nil {
		logger.Error(err, SignalConfigParseFailure)
		return
	}
	// Config is fetchable and parseable to us, yet the provider rejected it (e.g. GoFF
	// schema validation) — still a fallback to the secure default.
	logger.Error(ErrApiEnabledAbsent, SignalConfigLoadFailure)
}
