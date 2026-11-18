package feature

import (
	"context"
	"strings"

	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// BootstrapKindGroup is the kind-group for the CloudResources bootstrap object.
// It is always-on regardless of per-feature flag state and is the only exception
// to the trial-SKR disable-all rule.
const BootstrapKindGroup = "cloudresources.cloud-resources.kyma-project.io"

// isBootstrapKindGroup returns true when the context carries the
// cloudresources.cloud-resources.kyma-project.io kind-group. The bootstrap
// CloudResources object must always be accessible regardless of per-feature
// flag state, so both ApiDisabled and ApiEnabled short-circuit on it.
func isBootstrapKindGroup(ctx context.Context) bool {
	ffCtx := MustContextFromCtx(ctx)
	return strings.Contains(
		util.CastInterfaceToString(ffCtx.GetCustom()[types.KeyAllKindGroups]),
		BootstrapKindGroup,
	)
}

// brokerPlanTrial is the broker plan of a trial SKR. Trial SKRs disable every
// non-bootstrap API (the "disable all on trial" config rule), so the always-on
// carve-out does not apply to them.
const brokerPlanTrial = "trial"

// isTrialBrokerPlan reports whether the context targets a trial SKR.
func isTrialBrokerPlan(ctx context.Context) bool {
	return util.CastInterfaceToString(MustContextFromCtx(ctx).GetCustom()[types.KeyBrokerPlan]) == brokerPlanTrial
}

// isAlwaysEnabledApi returns true when the API must always be enabled regardless
// of config. This covers the bootstrap kind-group (CloudResources), and the two
// always-on feature buckets: FeatureFoundation and FeatureUnknown. Trial SKRs are
// the one exception to the feature buckets: there the legacy "disable all on trial"
// stance still applies, so foundation/unknown fall through to the flag. The bootstrap
// kind-group remains always-on even on trial.
func isAlwaysEnabledApi(ctx context.Context) bool {
	if isBootstrapKindGroup(ctx) {
		return true
	}
	if isTrialBrokerPlan(ctx) {
		return false
	}
	f := util.CastInterfaceToString(MustContextFromCtx(ctx).GetCustom()[types.KeyFeature])
	return f == types.FeatureFoundation || f == types.FeatureUnknown
}
