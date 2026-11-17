package feature

import (
	"context"
	"strings"

	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// isBootstrapKindGroup returns true when the context carries the
// cloudresources.cloud-resources.kyma-project.io kind-group. The bootstrap
// CloudResources object must always be accessible regardless of per-feature
// flag state, so both ApiDisabled and ApiEnabled short-circuit on it.
func isBootstrapKindGroup(ctx context.Context) bool {
	ffCtx := MustContextFromCtx(ctx)
	return strings.Contains(
		util.CastInterfaceToString(ffCtx.GetCustom()[types.KeyAllKindGroups]),
		"cloudresources.cloud-resources.kyma-project.io",
	)
}

// isAlwaysEnabledApi returns true when the API must always be enabled regardless
// of config. This covers the bootstrap kind-group (CloudResources), and the two
// always-on feature buckets: FeatureFoundation and FeatureUnknown.
func isAlwaysEnabledApi(ctx context.Context) bool {
	if isBootstrapKindGroup(ctx) {
		return true
	}
	f := util.CastInterfaceToString(MustContextFromCtx(ctx).GetCustom()[types.KeyFeature])
	return f == types.FeatureFoundation || f == types.FeatureUnknown
}
