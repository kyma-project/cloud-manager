package feature

import (
	"regexp"

	"github.com/kyma-project/cloud-manager/pkg/feature/types"
)

// providerEqPattern, providerRevPattern, providerInListPattern, and providerQuotedLiteralPattern
// match provider-key constraints in GoFF targeting queries, scoped to the provider key only.
// Equality accepts both GoFF operator spellings (== and eq); in-list accepts in and IN.
var providerEqPattern = regexp.MustCompile(`(?i)(\w+)\s*(?:==|\beq\b)\s*["']([^"']+)["']`)
var providerRevPattern = regexp.MustCompile(`(?i)["']([^"']+)["']\s*(?:==|\beq\b)\s*(\w+)`)
var providerInListPattern = regexp.MustCompile(`(?i)(\w+)\s+in\s+\[([^\]]*)\]`)
var providerQuotedLiteralPattern = regexp.MustCompile(`["']([^"']+)["']`)

// quotedLiteralPattern strips quoted values before operator detection so a value such as
// "restore" cannot trip the disjunction/negation scan.
var quotedLiteralPattern = regexp.MustCompile(`["'][^"']*["']`)

// disjunctionOrNegationPattern matches a disjunction (or) or any negation (not, ne, !=, !). A rule
// composed with one can enable providers its provider clause does not bound.
var disjunctionOrNegationPattern = regexp.MustCompile(`(?i)\bor\b|\bnot\b|\bne\b|!=|!`)

// queryConstrainsProvider reports whether query bounds the whole rule to explicit providers:
// it names the provider key through an equality (provider == "x" / "x" == provider, spelled
// == or eq) or a non-empty in-list (provider in ["x", ...]), AND carries no disjunction or
// negation that could widen the rule past that clause. A feature-only constraint, an empty
// in-list, or any or/not/!=-composed query does not satisfy it — an unbounded rule is treated
// as a blanket enable (fail-safe) rather than passed.
func queryConstrainsProvider(query string) bool {
	if disjunctionOrNegationPattern.MatchString(quotedLiteralPattern.ReplaceAllString(query, "")) {
		return false
	}
	for _, m := range providerEqPattern.FindAllStringSubmatch(query, -1) {
		if m[1] == types.KeyProvider {
			return true
		}
	}
	for _, m := range providerRevPattern.FindAllStringSubmatch(query, -1) {
		if m[2] == types.KeyProvider {
			return true
		}
	}
	for _, m := range providerInListPattern.FindAllStringSubmatch(query, -1) {
		if m[1] == types.KeyProvider && providerQuotedLiteralPattern.MatchString(m[2]) {
			return true
		}
	}
	return false
}
