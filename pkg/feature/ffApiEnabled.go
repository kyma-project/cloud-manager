package feature

import (
	"context"
)

const apiEnabledFlagName = "apiEnabled"

var ApiEnabled = &apiEnabledInfo{}

type apiEnabledInfo struct{}

func (f *apiEnabledInfo) Value(ctx context.Context) bool {
	if isAlwaysEnabledApi(ctx) {
		return true
	}
	return provider.BoolVariation(ctx, apiEnabledFlagName, false)
}
