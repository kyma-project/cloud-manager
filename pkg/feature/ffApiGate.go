package feature

import (
	"context"
	"strings"

	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// BootstrapKindGroup is the kind-group for the CloudResources bootstrap object.
// It is always-on regardless of per-feature flag state.
const BootstrapKindGroup = "cloudresources.cloud-resources.kyma-project.io"

// isBootstrapKindGroup returns true when the context carries the
// cloudresources.cloud-resources.kyma-project.io kind-group. The bootstrap
// CloudResources object must always be accessible regardless of per-feature
// flag state, so the ApiEnabled guard short-circuits on it.
func isBootstrapKindGroup(ctx context.Context) bool {
	ffCtx := MustContextFromCtx(ctx)
	return strings.Contains(
		util.CastInterfaceToString(ffCtx.GetCustom()[types.KeyAllKindGroups]),
		BootstrapKindGroup,
	)
}

// isFoundationOrUnknown returns true when the context feature is FeatureFoundation or FeatureUnknown.
// These are the floor for non-rule-matched contexts; an explicit config deny rule overrides the floor.
func isFoundationOrUnknown(ctx context.Context) bool {
	f := util.CastInterfaceToString(MustContextFromCtx(ctx).GetCustom()[types.KeyFeature])
	return f == types.FeatureFoundation || f == types.FeatureUnknown
}
