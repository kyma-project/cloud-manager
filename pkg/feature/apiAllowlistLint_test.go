package feature

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	require.NoError(t, err, "reading testdata/%s", name)
	return b
}

func TestApiAllowlistLint_ProviderScoped_Pass(t *testing.T) {
	b := readTestdata(t, "apiEnabledProviderScoped.yaml")
	result := RenderLintResult(LintApiEnabled(b))
	assert.Equal(t, "PASS", result)
}

func TestApiAllowlistLint_InListProvider_Pass(t *testing.T) {
	b := readTestdata(t, "apiEnabledInListProvider.yaml")
	result := RenderLintResult(LintApiEnabled(b))
	assert.Equal(t, "PASS", result)
}

func TestApiAllowlistLint_BlanketRules_TwoViolations(t *testing.T) {
	b := readTestdata(t, "apiEnabledBlanket.yaml")
	violations := LintApiEnabled(b)
	result := RenderLintResult(violations)

	assert.Len(t, violations, 2)
	assert.True(t, strings.HasPrefix(result, "FAIL: 2 violations"), "result=%q", result)
	// First blanket violation.
	assert.Contains(t, result, `"Enable Blanket 1"`)
	assert.Contains(t, result, "(index 1)")
	assert.Contains(t, result, `feature in ["nfs", "redis"]`)
	// Second blanket violation must be reported too (guards second-rule detail accuracy).
	assert.Contains(t, result, `"Enable Blanket 2"`)
	assert.Contains(t, result, "(index 2)")
	assert.Contains(t, result, `feature in ["peering"]`)
}

func TestApiAllowlistLint_DefaultEnabled_Fail(t *testing.T) {
	b := readTestdata(t, "apiEnabledDefaultEnabled.yaml")
	violations := LintApiEnabled(b)
	require.Len(t, violations, 1)
	assert.Equal(t, violationDefaultNotDisabled, violations[0].Type)
	assert.Equal(t, variationEnabled, violations[0].DefaultVariation)
	result := RenderLintResult(violations)
	assert.Contains(t, result, `defaultRule.variation is "enabled"`)
}

func TestApiAllowlistLint_MissingDefaultRule_Fail(t *testing.T) {
	b := readTestdata(t, "apiEnabledMissingDefaultRule.yaml")
	violations := LintApiEnabled(b)
	require.Len(t, violations, 1)
	assert.Equal(t, violationDefaultNotDisabled, violations[0].Type)
	assert.Equal(t, "", violations[0].DefaultVariation, "a missing defaultRule renders an empty actual value")
}

func TestApiAllowlistLint_AbsentBlock_Fail(t *testing.T) {
	b := readTestdata(t, "apiEnabledNoBlock.yaml")
	violations := LintApiEnabled(b)
	require.Len(t, violations, 1)
	assert.Equal(t, violationAbsent, violations[0].Type)
	result := RenderLintResult(violations)
	assert.Contains(t, result, "FAIL:")
	assert.Contains(t, result, "absent")
}

func TestApiAllowlistLint_EmptyBlock_Fail(t *testing.T) {
	b := readTestdata(t, "apiEnabledEmptyBlock.yaml")
	violations := LintApiEnabled(b)
	assert.Len(t, violations, 1)
	assert.Equal(t, violationEmpty, violations[0].Type)
	assert.Contains(t, RenderLintResult(violations), "empty")
}

func TestApiAllowlistLint_NonTrialDisable_Pass(t *testing.T) {
	b := readTestdata(t, "apiEnabledNonTrialDisable.yaml")
	result := RenderLintResult(LintApiEnabled(b))
	assert.Equal(t, "PASS", result)
}

func TestApiAllowlistLint_FakeDisable_Fail(t *testing.T) {
	// Rule name starts with "disable" but its variation resolves to enabled: the lint must
	// flag it by resolved variation value, not by name/prefix.
	b := readTestdata(t, "apiEnabledFakeDisable.yaml")
	violations := LintApiEnabled(b)
	require.Len(t, violations, 1)
	assert.Equal(t, violationBlanketEnable, violations[0].Type)
	result := RenderLintResult(violations)
	assert.Contains(t, result, `"disable-audit"`)
	assert.Contains(t, result, "(index 0)")
	assert.Contains(t, result, `feature in ["audit"]`)
}

func TestApiAllowlistLint_AliasVariation_Fail(t *testing.T) {
	// A blanket enable rule using a variation aliased to true (not literally "enabled") must
	// still be flagged — the lint resolves the variation value rather than matching the name.
	b := readTestdata(t, "apiEnabledAliasVariation.yaml")
	violations := LintApiEnabled(b)
	require.Len(t, violations, 1)
	assert.Equal(t, violationBlanketEnable, violations[0].Type)
	assert.Contains(t, RenderLintResult(violations), `"Enable Blanket via alias"`)
}

func TestApiAllowlistLint_EdgeYaml_Skip(t *testing.T) {
	b := readTestdata(t, "../ff_edge.yaml")
	result := LintFileReport("ff_edge.yaml", b)
	assert.Contains(t, result, "SKIP")
	assert.Contains(t, result, "ff_edge.yaml")
}

func TestApiAllowlistLint_EmbeddedGA_Pass(t *testing.T) {
	result := RenderLintResult(LintApiEnabled(embeddedConfig))
	assert.Equal(t, "PASS", result, "embedded ff_ga.yaml must pass the apiEnabled allowlist lint")
}

func TestApiAllowlistLint_Malformed_ParseError(t *testing.T) {
	// A structurally invalid config must surface a parse error, not be misreported as an
	// absent apiEnabled block.
	b := readTestdata(t, "apiMalformed.yaml")
	violations := LintApiEnabled(b)
	require.Len(t, violations, 1)
	assert.Equal(t, violationParseError, violations[0].Type)
	assert.Contains(t, RenderLintResult(violations), "unparseable")
}

func TestApiAllowlistLint_UnknownVariation_Fail(t *testing.T) {
	// A rule whose variation name is absent from the variations map resolves ok=false; the
	// fail-safe treats it as checked (not exempt), so a blanket query under an unknown
	// variation is still flagged rather than slipping through.
	b := readTestdata(t, "apiEnabledUnknownVariation.yaml")
	violations := LintApiEnabled(b)
	require.Len(t, violations, 1)
	assert.Equal(t, violationBlanketEnable, violations[0].Type)
}

func TestApiAllowlistLint_OrComposedEnable_Fail(t *testing.T) {
	// An enable rule that names a provider but ORs in a feature-only branch enables every
	// provider via that branch; the provider token must not make it pass the anti-blanket check.
	b := readTestdata(t, "apiEnabledOrComposed.yaml")
	violations := LintApiEnabled(b)
	require.Len(t, violations, 1)
	assert.Equal(t, violationBlanketEnable, violations[0].Type)
}
