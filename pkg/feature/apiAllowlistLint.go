package feature

import (
	"fmt"
	"strings"
)

// LintViolationType identifies the category of an apiEnabled allowlist violation.
type LintViolationType int

const (
	// violationAbsent: apiEnabled block absent from the config file.
	violationAbsent LintViolationType = iota
	// violationEmpty: apiEnabled block is present but carries no rules or defaultRule.
	violationEmpty
	// violationBlanketEnable: an "enabled" targeting rule does not constrain the provider key.
	violationBlanketEnable
	// violationDefaultNotDisabled: defaultRule.variation is not "disabled".
	violationDefaultNotDisabled
	// violationParseError: the config bytes are not valid YAML.
	violationParseError
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

// LintViolation describes one invariant failure in the apiEnabled block.
type LintViolation struct {
	Type             LintViolationType
	RuleName         string
	RuleIndex        int    // 0-based index in targeting; -1 for block-level violations
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
func LintApiEnabled(configBytes []byte) []LintViolation {
	root, err := parseConfigBytes(configBytes)
	if err != nil {
		return []LintViolation{{Type: violationParseError, RuleIndex: -1, Detail: err.Error()}}
	}
	if root == nil {
		return []LintViolation{{Type: violationAbsent, RuleIndex: -1}}
	}

	rawBlock, exists := root[apiEnabledFlagName]
	if !exists || rawBlock == nil {
		return []LintViolation{{Type: violationAbsent, RuleIndex: -1}}
	}
	if !blockHasContent(rawBlock) {
		return []LintViolation{{Type: violationEmpty, RuleIndex: -1}}
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
					Type:      violationBlanketEnable,
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
				Type:             violationDefaultNotDisabled,
				RuleIndex:        -1,
				DefaultVariation: variation,
			})
		}
	} else {
		// defaultRule missing entirely — also a violation.
		violations = append(violations, LintViolation{
			Type:      violationDefaultNotDisabled,
			RuleIndex: -1,
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
	case violationAbsent:
		return "apiEnabled block absent"
	case violationEmpty:
		return "apiEnabled block present but empty"
	case violationBlanketEnable:
		return fmt.Sprintf(
			"rule %q (index %d): no provider constraint in query '%s' — add e.g. 'provider == \"gcp\" and ...' (valid providers: gcp, aws, azure, openstack, alicloud)",
			v.RuleName, v.RuleIndex, v.Query,
		)
	case violationDefaultNotDisabled:
		return fmt.Sprintf("defaultRule.variation is %q — must be %q", v.DefaultVariation, variationDisabled)
	case violationParseError:
		return fmt.Sprintf("config unparseable: %s", v.Detail)
	default:
		return fmt.Sprintf("unknown violation type %d", v.Type)
	}
}
