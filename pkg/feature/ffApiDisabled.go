package feature

import (
	"context"
)

const apiDisabledFlagName = "apiDisabled"

var ApiDisabled = &apiDisabledInfo{}

type apiDisabledInfo struct{}

func (f *apiDisabledInfo) Value(ctx context.Context) bool {
	if isAlwaysEnabledApi(ctx) {
		return false
	}
	return provider.BoolVariation(ctx, apiDisabledFlagName, false)
}
