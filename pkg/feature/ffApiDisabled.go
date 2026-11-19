package feature

import (
	"context"
)

const apiDisabledFlagName = "apiDisabled"

var ApiDisabled = &apiDisabledInfo{}

type apiDisabledInfo struct{}

func (f *apiDisabledInfo) Value(ctx context.Context) bool {
	// Legacy flag, frozen during the apiEnabled migration: it short-circuits only
	// on the bootstrap kind-group and otherwise reads config, so it stays the stable
	// oracle the parity checker validates apiEnabled against. The foundation/unknown
	// always-on policy lives on apiEnabled, not here.
	if isBootstrapKindGroup(ctx) {
		return false
	}
	return provider.BoolVariation(ctx, apiDisabledFlagName, false)
}
