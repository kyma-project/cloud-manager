package feature

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/feature/types"
)

const apiEnabledFlagName = "apiEnabled"

var ApiEnabled = &apiEnabledInfo{}

type apiEnabledInfo struct{}

func (f *apiEnabledInfo) Value(ctx context.Context) bool {
	if isBootstrapKindGroup(ctx) {
		return true
	}
	var v, byRule bool
	if rp, ok := provider.(types.ReasonAwareProvider); ok {
		v, byRule = rp.BoolVariationReason(ctx, apiEnabledFlagName, false)
	} else {
		v = provider.BoolVariation(ctx, apiEnabledFlagName, false)
	}
	if byRule {
		return v
	}
	return v || isFoundationOrUnknown(ctx)
}
