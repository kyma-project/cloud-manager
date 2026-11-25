package feature

import (
	"fmt"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
)

// validProvidersHint lists the provider keys accepted in an apiEnabled targeting query,
// derived from the canonical provider set so the lint diagnostic cannot drift from it.
var validProvidersHint = func() string {
	providers := cloudcontrolv1beta1.AllProviders()
	names := make([]string, len(providers))
	for i, p := range providers {
		names[i] = string(p)
	}
	return strings.Join(names, ", ")
}()

// LintViolationType identifies the category of an apiEnabled allowlist violation.
type LintViolationType int

const (
	// ViolationAbsent: apiEnabled block absent from the config file.
	ViolationAbsent LintViolationType = iota
	// ViolationEmpty: apiEnabled block is present but carries no rules or defaultRule.
	ViolationEmpty
	// ViolationBlanketEnable: an "enabled" targeting rule does not constrain the provider key.
	ViolationBlanketEnable
	// ViolationDefaultNotDisabled: defaultRule.variation is not "disabled".
	ViolationDefaultNotDisabled
	// ViolationParseError: the config bytes are not valid YAML.
	ViolationParseError
)

const (
	variationEnabled  = "enabled"
	variationDisabled = "disabled"
)

// exemptFiles lists config base filenames that are exempt from the apiEnabled allowlist lint.
// Add an entry here only for files with a legitimately open default (e.g. e2e clusters).
var exemptFiles = map[string]struct{}{
	"ff_edge.yaml": {}, // e2e: open default is intentional for integration-test clusters
}

// blockLevelRuleIndex is the RuleIndex sentinel for violations not tied to a single
// targeting rule (absent/empty block, non-disabled default, parse error).
const blockLevelRuleIndex = -1

// LintViolation describes one invariant failure in the apiEnabled block.
type LintViolation struct {
	Type             LintViolationType
	RuleName         string
	RuleIndex        int    // 0-based index in targeting; blockLevelRuleIndex for block-level violations
	Query            string // populated for violationBlanketEnable
	DefaultVariation string // populated for violationDefaultNotDisabled
	Detail           string // populated for violationParseError
}

// LintApiEnabled checks the three invariants of the apiEnabled allowlist in configBytes:
//  1. apiEnabled block must be present and non-empty.
//  2. Every targeting rule whose variation resolves to enabled must constrain the provider key.
//  3. defaultRule.variation must be "disabled".
//
// A nil or empty slice return means the config is clean. On a block-level violation
// (absent or empty) it returns early with a single violation and skips rule-level checks.
//
// Scope: guards only the embedded ff_ga.yaml GA baseline; it is reached only from tests and
// never at runtime, and validates no externally supplied config. It is provider-only by
// design — intentionally stricter than a general allowlist validator, and correct only for
// the embedded GA matrix, which uses provider-scoped rules exclusively.
func LintApiEnabled(configBytes []byte) []LintViolation {
	root, err := parseConfigBytes(configBytes)
	if err != nil {
		return []LintViolation{{Type: ViolationParseError, RuleIndex: blockLevelRuleIndex, Detail: err.Error()}}
	}
	if root == nil {
		return []LintViolation{{Type: ViolationAbsent, RuleIndex: blockLevelRuleIndex}}
	}

	rawBlock, exists := root[apiEnabledFlagName]
	if !exists || rawBlock == nil {
		return []LintViolation{{Type: ViolationAbsent, RuleIndex: blockLevelRuleIndex}}
	}
	if !blockHasContent(rawBlock) {
		return []LintViolation{{Type: ViolationEmpty, RuleIndex: blockLevelRuleIndex}}
	}

	block, _ := rawBlock.(map[string]any)
	var violations []LintViolation

	// Invariant 2: each rule that resolves to enabled must constrain the provider.
	// Exemption is keyed on the resolved variation value (via the variations table),
	// not the variation name — a rule named "disable-*" that resolves to enabled is
	// still an enable rule, and a rule with an unresolvable variation is checked
	// (fail-safe: a blanket rule cannot evade the lint by using an alias variation).
	if targeting, ok := block["targeting"].([]any); ok {
		for i, raw := range targeting {
			rule, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			variationName, _ := rule["variation"].(string)
			if resolved, ok := resolveVariation(block, variationName); ok && !resolved {
				continue // rule resolves to disabled → exempt from the provider-constraint check
			}
			query, _ := rule["query"].(string)
			if !queryConstrainsProvider(query) {
				name, _ := rule["name"].(string)
				violations = append(violations, LintViolation{
					Type:      ViolationBlanketEnable,
					RuleName:  name,
					RuleIndex: i,
					Query:     query,
				})
			}
		}
	}

	// Invariant 3: defaultRule.variation must be "disabled".
	if defaultRule, ok := block["defaultRule"].(map[string]any); ok {
		variation, _ := defaultRule["variation"].(string)
		if variation != variationDisabled {
			violations = append(violations, LintViolation{
				Type:             ViolationDefaultNotDisabled,
				RuleIndex:        blockLevelRuleIndex,
				DefaultVariation: variation,
			})
		}
	} else {
		// defaultRule missing entirely — also a violation.
		violations = append(violations, LintViolation{
			Type:      ViolationDefaultNotDisabled,
			RuleIndex: blockLevelRuleIndex,
		})
	}

	return violations
}

// resolveVariation looks up a rule's variation name in the block's variations table and
// returns the boolean it maps to. ok is false when the block has no variations map or the
// name is absent/non-boolean, so callers can fail safe rather than assume "disabled".
func resolveVariation(block map[string]any, variationName string) (value bool, ok bool) {
	variations, isMap := block["variations"].(map[string]any)
	if !isMap {
		return false, false
	}
	raw, found := variations[variationName]
	if !found {
		return false, false
	}
	b, isBool := raw.(bool)
	return b, isBool
}

// LintFileReport runs LintApiEnabled on configBytes after checking the exemptFiles list.
// It returns a human-readable result string: "SKIP: ...", "PASS", or "FAIL: N violations\n...".
func LintFileReport(basename string, configBytes []byte) string {
	if _, exempt := exemptFiles[basename]; exempt {
		return fmt.Sprintf("SKIP: %s — exempt (e2e open default permitted)", basename)
	}
	return RenderLintResult(LintApiEnabled(configBytes))
}

// RenderLintResult converts violations to a human-readable string.
// Returns "PASS" when violations is empty, otherwise "FAIL: N violations\n  <detail>\n...".
func RenderLintResult(violations []LintViolation) string {
	if len(violations) == 0 {
		return "PASS"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "FAIL: %d violation", len(violations))
	if len(violations) != 1 {
		sb.WriteString("s")
	}
	for _, v := range violations {
		sb.WriteString("\n  ")
		sb.WriteString(renderViolation(v))
	}
	return sb.String()
}

func renderViolation(v LintViolation) string {
	switch v.Type {
	case ViolationAbsent:
		return "apiEnabled block absent"
	case ViolationEmpty:
		return "apiEnabled block present but empty"
	case ViolationBlanketEnable:
		return fmt.Sprintf(
			"rule %q (index %d): no provider constraint in query '%s' — add e.g. 'provider == \"gcp\" and ...' (valid providers: %s)",
			v.RuleName, v.RuleIndex, v.Query, validProvidersHint,
		)
	case ViolationDefaultNotDisabled:
		return fmt.Sprintf("defaultRule.variation is %q — must be %q", v.DefaultVariation, variationDisabled)
	case ViolationParseError:
		return fmt.Sprintf("config unparseable: %s", v.Detail)
	default:
		return fmt.Sprintf("unknown violation type %d", v.Type)
	}
}
