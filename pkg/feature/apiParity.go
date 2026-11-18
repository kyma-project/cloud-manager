package feature

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"gopkg.in/yaml.v3"
)

// Divergence records a context where ApiEnabled.Value(ctx) != !ApiDisabled.Value(ctx).
type Divergence struct {
	Feature        string
	Provider       string
	BrokerPlan     string
	Landscape      string
	GlobalAccount  string
	Shoot          string
	FlagValue      bool
	LegacyDecision bool // LegacyDecision == !ApiDisabled decision
}

func (d Divergence) String() string {
	return fmt.Sprintf(
		"feature=%q provider=%q brokerPlan=%q landscape=%q globalAccount=%q shoot=%q apiEnabled=%v expected=%v",
		d.Feature, d.Provider, d.BrokerPlan, d.Landscape, d.GlobalAccount, d.Shoot,
		d.FlagValue, d.LegacyDecision,
	)
}

// parityCorpusDimensions holds the seeded dimension slices used to build the cross-product corpus.
type parityCorpusDimensions struct {
	Features       []string
	Providers      []string
	BrokerPlans    []string
	Landscapes     []string
	GlobalAccounts []string
	Shoots         []string
}

// parityCorpusEntry carries both the built context and its dimension values so
// divergence reports are self-documenting without re-reading the context.
type parityCorpusEntry struct {
	ctx           context.Context
	Feature       string
	Provider      string
	BrokerPlan    string
	Landscape     string
	GlobalAccount string
	Shoot         string
}

// CheckApiParity reads the feature-flag config at configPath, initialises the
// package-global feature provider, and returns the set of evaluation contexts
// where ApiEnabled.Value(ctx) disagrees with !ApiDisabled.Value(ctx).
//
// It is the primary entrypoint for parity validation, used by the ffparity CLI.
// Note the side effect: it reinitialises the package-global provider, so it is
// not safe to call from code already serving the live provider.
func CheckApiParity(ctx context.Context, configPath string) ([]Divergence, error) {
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", configPath, err)
	}

	if err := Initialize(ctx, logr.Discard(), WithFile(configPath)); err != nil {
		return nil, fmt.Errorf("initialising provider from %s: %w", configPath, err)
	}

	dims, err := buildParityDimensions(configBytes)
	if err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", configPath, err)
	}
	corpus := buildParityCorpus(ctx, dims)
	return evaluateParity(corpus), nil
}

// buildParityDimensions constructs dimension slices for the corpus, including
// globalAccount and shoot values self-seeded from the config's targeting queries.
// FeatureFoundation and FeatureUnknown are hardcoded always-on in code (independent
// of config); the corpus includes them so the checker confirms parity trivially holds
// for them on every run — they will never produce a divergence.
// A parse failure in the seed-extraction step is returned as an error so callers
// treat a malformed config as a hard error rather than a zero-divergence pass.
func buildParityDimensions(configBytes []byte) (parityCorpusDimensions, error) {
	features := types.AllFeatureNames()

	providers := []string{
		"", // no provider set
		string(cloudcontrolv1beta1.ProviderAws),
		string(cloudcontrolv1beta1.ProviderGCP),
		string(cloudcontrolv1beta1.ProviderAzure),
		string(cloudcontrolv1beta1.ProviderOpenStack),
		string(cloudcontrolv1beta1.ProviderAlicloud),
	}

	brokerPlans := []string{
		brokerPlanTrial,
		"", // non-trial sentinel: empty string means standard/non-trial plan
	}

	landscapes := []string{
		types.LandscapeDev,
		types.LandscapeStage,
		types.LandscapeProd,
	}

	seededAccounts, seededShoots, err := extractParitySeedValues(configBytes)
	if err != nil {
		return parityCorpusDimensions{}, fmt.Errorf("parsing config: %w", err)
	}

	return parityCorpusDimensions{
		Features:    features,
		Providers:   providers,
		BrokerPlans: brokerPlans,
		Landscapes:  landscapes,
		// Always lead with "" so no-account/no-shoot rows are always present.
		GlobalAccounts: dedupeStrings("", seededAccounts...),
		Shoots:         dedupeStrings("", seededShoots...),
	}, nil
}

// buildParityCorpus returns the cross-product of all dimensions as a slice of
// parityCorpusEntry values. Corpus growth requires only changes to this function.
func buildParityCorpus(baseCtx context.Context, dims parityCorpusDimensions) []parityCorpusEntry {
	var corpus []parityCorpusEntry
	for _, f := range dims.Features {
		for _, p := range dims.Providers {
			for _, bp := range dims.BrokerPlans {
				for _, l := range dims.Landscapes {
					for _, ga := range dims.GlobalAccounts {
						for _, sh := range dims.Shoots {
							corpus = append(corpus, buildCorpusEntry(baseCtx, f, p, bp, l, ga, sh))
						}
					}
				}
			}
		}
	}
	return corpus
}

// buildCorpusEntry constructs a single parityCorpusEntry from the given dimension values,
// omitting empty optional dimensions so no-provider/no-plan/no-account/no-shoot contexts
// match the flag's "not set" default.
func buildCorpusEntry(baseCtx context.Context, f, p, bp, l, ga, sh string) parityCorpusEntry {
	b := ContextBuilderFromCtx(baseCtx).Feature(f).Landscape(l)
	if p != "" {
		b = b.Provider(p)
	}
	if bp != "" {
		b = b.BrokerPlan(bp)
	}
	if ga != "" {
		b = b.GlobalAccount(ga)
	}
	if sh != "" {
		b = b.Shoot(sh)
	}
	return parityCorpusEntry{
		ctx:           b.Build(baseCtx),
		Feature:       f,
		Provider:      p,
		BrokerPlan:    bp,
		Landscape:     l,
		GlobalAccount: ga,
		Shoot:         sh,
	}
}

// evaluateParity compares ApiEnabled and !ApiDisabled for each corpus entry and
// collects entries where they disagree. Separated from corpus generation so each
// concern is independently testable and growable.
func evaluateParity(corpus []parityCorpusEntry) []Divergence {
	var divergences []Divergence
	for _, e := range corpus {
		enabled := ApiEnabled.Value(e.ctx)
		expected := !ApiDisabled.Value(e.ctx)
		if enabled != expected {
			divergences = append(divergences, Divergence{
				Feature:        e.Feature,
				Provider:       e.Provider,
				BrokerPlan:     e.BrokerPlan,
				Landscape:      e.Landscape,
				GlobalAccount:  e.GlobalAccount,
				Shoot:          e.Shoot,
				FlagValue:      enabled,
				LegacyDecision: expected,
			})
		}
	}
	return divergences
}

// extractParitySeedValues parses the raw YAML config and returns every distinct
// globalAccount and shoot string literal found in targeting-rule queries. This
// ensures that account- and shoot-targeted rules are exercised by the corpus:
// a rule targeting an account absent from the corpus would never be hit.
// An unmarshal failure is returned as an error so callers can distinguish a
// malformed config from a valid-but-empty one.
func extractParitySeedValues(configBytes []byte) (globalAccounts []string, shoots []string, err error) {
	var root map[string]interface{}
	if unmarshalErr := yaml.Unmarshal(configBytes, &root); unmarshalErr != nil {
		return nil, nil, unmarshalErr
	}

	gaSet := map[string]struct{}{}
	shSet := map[string]struct{}{}

	for _, flagVal := range root {
		flagMap, ok := flagVal.(map[string]interface{})
		if !ok {
			continue
		}
		targeting, ok := flagMap["targeting"]
		if !ok {
			continue
		}
		rules, ok := targeting.([]interface{})
		if !ok {
			continue
		}
		for _, rule := range rules {
			ruleMap, ok := rule.(map[string]interface{})
			if !ok {
				continue
			}
			query, ok := ruleMap["query"].(string)
			if !ok {
				continue
			}
			for _, v := range extractQueryLiterals(query, "globalAccount") {
				gaSet[v] = struct{}{}
			}
			for _, v := range extractQueryLiterals(query, "shoot") {
				shSet[v] = struct{}{}
			}
		}
	}

	for ga := range gaSet {
		globalAccounts = append(globalAccounts, ga)
	}
	for sh := range shSet {
		shoots = append(shoots, sh)
	}
	return globalAccounts, shoots, nil
}

// quotedLiteral matches a single- or double-quoted string literal, capturing its value.
var quotedLiteral = regexp.MustCompile(`["']([^"']+)["']`)

// eqPattern matches `word == "value"` or `word == 'value'`, capturing the word and value.
var eqPattern = regexp.MustCompile(`(\w+)\s*==\s*["']([^"']+)["']`)

// revPattern matches `"value" == word` or `'value' == word`, capturing the value and word.
var revPattern = regexp.MustCompile(`["']([^"']+)["']\s*==\s*(\w+)`)

// inListPattern matches `word in [...]`, capturing the word and the bracketed list content.
var inListPattern = regexp.MustCompile(`(\w+)\s+in\s+\[([^\]]*)\]`)

// extractQueryLiterals returns every string literal a targeting-rule query associates
// with key, across the query syntaxes GoFF accepts: `key == "v"`, `key == 'v'`, the
// reversed `"v" == key`, and set membership `key in ["v1", "v2"]`. Matching any of these
// keeps the seeded corpus from missing an account/shoot a rule actually targets.
func extractQueryLiterals(query, key string) []string {
	var out []string
	for _, m := range eqPattern.FindAllStringSubmatch(query, -1) {
		if m[1] == key {
			out = append(out, m[2])
		}
	}
	for _, m := range revPattern.FindAllStringSubmatch(query, -1) {
		if m[2] == key {
			out = append(out, m[1])
		}
	}
	for _, m := range inListPattern.FindAllStringSubmatch(query, -1) {
		if m[1] == key {
			for _, lit := range quotedLiteral.FindAllStringSubmatch(m[2], -1) {
				out = append(out, lit[1])
			}
		}
	}
	return out
}

// dedupeStrings returns a deduplicated slice starting with seed followed by
// any extra values not already present.
func dedupeStrings(seed string, extra ...string) []string {
	seen := map[string]struct{}{seed: {}}
	result := []string{seed}
	for _, s := range extra {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result
}
