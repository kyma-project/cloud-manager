package feature

import (
	"context"
	"encoding/json"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/tidwall/gjson"
	"strconv"
)

type providerConfig struct {
	json string
}

func NewProviderConfig(env abstractions.Environment) types.Provider {
	cfg := config.NewConfig(env)
	cfg.SourceEnv("", "FF_")
	cfg.Read()
	js := cfg.Json()

	return &providerConfig{json: js}
}

func (p *providerConfig) BoolVariation(ctx context.Context, flagKey string, defaultValue bool) bool {
	res := gjson.Get(p.json, flagKey)
	if res.Type == gjson.Null {
		return defaultValue
	}
	v, err := strconv.ParseBool(res.String())
	if err != nil {
		return defaultValue
	}
	return v
}

func (p *providerConfig) IntVariation(ctx context.Context, flagKey string, defaultValue int) int {
	res := gjson.Get(p.json, flagKey)
	if res.Type == gjson.Null {
		return defaultValue
	}
	v, err := strconv.ParseInt(res.String(), 10, 64)
	if err != nil {
		return defaultValue
	}
	return int(v)
}

func (p *providerConfig) Float64Variation(ctx context.Context, flagKey string, defaultValue float64) float64 {
	res := gjson.Get(p.json, flagKey)
	if res.Type == gjson.Null {
		return defaultValue
	}
	v, err := strconv.ParseFloat(res.String(), 64)
	if err != nil {
		return defaultValue
	}
	return v
}

func (p *providerConfig) StringVariation(ctx context.Context, flagKey string, defaultValue string) string {
	res := gjson.Get(p.json, flagKey)
	if res.Type == gjson.Null {
		return defaultValue
	}
	return res.String()
}

func (p *providerConfig) JSONArrayVariation(ctx context.Context, flagKey string, defaultValue []interface{}) []interface{} {
	res := gjson.Get(p.json, flagKey)
	if res.Type == gjson.Null {
		return defaultValue
	}
	arr := []interface{}{}
	err := json.Unmarshal([]byte(res.String()), &arr)
	if err != nil {
		return defaultValue
	}
	return arr
}

func (p *providerConfig) JSONVariation(ctx context.Context, flagKey string, defaultValue map[string]interface{}) map[string]interface{} {
	res := gjson.Get(p.json, flagKey)
	if res.Type == gjson.Null {
		return defaultValue
	}
	obj := map[string]interface{}{}
	err := json.Unmarshal([]byte(res.String()), &obj)
	if err != nil {
		return defaultValue
	}
	return obj
}
