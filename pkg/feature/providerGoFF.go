package feature

import (
	"context"
	ffclient "github.com/thomaspoignant/go-feature-flag"
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

func (p *providerGoFF) JSONArrayVariation(ctx context.Context, flagKey string, defaultValue []interface{}) []interface{} {
	ffCtx := MustContextFromCtx(ctx)
	res, err := p.ff.JSONArrayVariation(flagKey, ffCtx, defaultValue)
	if err != nil {
		return defaultValue
	}
	return res
}

func (p *providerGoFF) JSONVariation(ctx context.Context, flagKey string, defaultValue map[string]interface{}) map[string]interface{} {
	ffCtx := MustContextFromCtx(ctx)
	res, err := p.ff.JSONVariation(flagKey, ffCtx, defaultValue)
	if err != nil {
		return defaultValue
	}
	return res
}
