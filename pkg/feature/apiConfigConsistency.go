package feature

import (
	"errors"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v3"
)

// Fail-closed signal messages. Kept stable and greppable — ops alerting keys on them.
const (
	SignalConfigLoadFailure  = "feature-flag config load failed; falling back to secure default (non-bootstrap APIs disabled)"
	SignalConfigParseFailure = "feature-flag config present but unparseable; falling back to secure default (non-bootstrap APIs disabled)"
	SignalApiEnabledMissing  = "feature-flag config carries apiDisabled rules but no apiEnabled block; secure-by-default disables every API lacking an apiEnabled rule"
)

// ErrApiConfigInconsistent is the cause attached to the apiEnabled-missing signal so
// the error-level log carries a non-nil error even though there is no underlying failure.
var ErrApiConfigInconsistent = errors.New("feature-flag config inconsistent")

// reportApiConfigFallback emits an error-level signal for each fail-closed condition
// detected while loading the feature-flag config: a retriever error, an unparseable
// config, or a config that carries apiDisabled rules but no apiEnabled block. It never
// aborts startup — the caller keeps the secure-default provider either way.
func reportApiConfigFallback(logger logr.Logger, configBytes []byte, retrieveErr error) {
	if retrieveErr != nil {
		logger.Error(retrieveErr, SignalConfigLoadFailure)
		return
	}
	var root map[string]any
	if err := yaml.Unmarshal(configBytes, &root); err != nil {
		logger.Error(err, SignalConfigParseFailure)
		return
	}
	if apiDisabledOnly(root) {
		logger.Error(ErrApiConfigInconsistent, SignalApiEnabledMissing)
	}
}

// apiDisabledOnly reports whether the parsed config carries an apiDisabled block with
// content while the apiEnabled block is absent or empty — the secure-by-default divergent
// case where every untagged API silently goes dark. A present-but-empty apiEnabled key
// (e.g. a partially-written config) counts as absent: it carries no rule, so untagged
// APIs still go dark and the operator should be warned.
func apiDisabledOnly(root map[string]any) bool {
	if blockHasContent(root["apiEnabled"]) {
		return false
	}
	return blockHasContent(root["apiDisabled"])
}

// blockHasContent reports whether a flag block carries targeting rules or a defaultRule.
// An empty block is not treated as an inconsistency.
func blockHasContent(block any) bool {
	m, ok := block.(map[string]any)
	if !ok {
		return false
	}
	if t, ok := m["targeting"].([]any); ok && len(t) > 0 {
		return true
	}
	_, hasDefault := m["defaultRule"]
	return hasDefault
}
