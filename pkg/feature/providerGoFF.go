package feature

import (
	"context"
	ffclient "github.com/thomaspoignant/go-feature-flag"
)

// GoFF v1.56.1 resolution-reason strings (OpenFeature-standard).
const (
	reasonTargetingMatch      = "TARGETING_MATCH"
	reasonTargetingMatchSplit = "TARGETING_MATCH_SPLIT"
)

type providerGoFF struct {
	ff *ffclient.GoFeatureFlag
}

func (p *providerGoFF) BoolVariation(ctx context.Context, flagKey string, defaultValue bool) bool {
	ffCtx := MustContextFromCtx(ctx)
	res, err := p.ff.BoolVariation(flagKey, ffCtx, defaultValue)
	if err != nil {
		return defaultValue
	}
	return res
}

func (p *providerGoFF) IntVariation(ctx context.Context, flagKey string, defaultValue int) int {
	ffCtx := MustContextFromCtx(ctx)
	res, err := p.ff.IntVariation(flagKey, ffCtx, defaultValue)
	if err != nil {
		return defaultValue
	}
	return res
}

func (p *providerGoFF) Float64Variation(ctx context.Context, flagKey string, defaultValue float64) float64 {
	ffCtx := MustContextFromCtx(ctx)
	res, err := p.ff.Float64Variation(flagKey, ffCtx, defaultValue)
	if err != nil {
		return defaultValue
	}
	return res
}

func (p *providerGoFF) StringVariation(ctx context.Context, flagKey string, defaultValue string) string {
	ffCtx := MustContextFromCtx(ctx)
	res, err := p.ff.StringVariation(flagKey, ffCtx, defaultValue)
	if err != nil {
		return defaultValue
	}
	return res
}

func (p *providerGoFF) JSONArrayVariation(ctx context.Context, flagKey string, defaultValue []any) []any {
	ffCtx := MustContextFromCtx(ctx)
	res, err := p.ff.JSONArrayVariation(flagKey, ffCtx, defaultValue)
	if err != nil {
		return defaultValue
	}
	return res
}

func (p *providerGoFF) JSONVariation(ctx context.Context, flagKey string, defaultValue map[string]any) map[string]any {
	ffCtx := MustContextFromCtx(ctx)
	res, err := p.ff.JSONVariation(flagKey, ffCtx, defaultValue)
	if err != nil {
		return defaultValue
	}
	return res
}

func (p *providerGoFF) BoolVariationReason(ctx context.Context, flagKey string, defaultValue bool) (value bool, byRule bool) {
	ffCtx := MustContextFromCtx(ctx)
	res, err := p.ff.BoolVariationDetails(flagKey, ffCtx, defaultValue)
	if err != nil {
		return defaultValue, false
	}
	return res.Value, reasonMatchesByRule(string(res.Reason))
}

// reasonMatchesByRule reports whether a GoFF resolution reason indicates a targeting rule was matched.
func reasonMatchesByRule(reason string) bool {
	return reason == reasonTargetingMatch || reason == reasonTargetingMatchSplit
}
