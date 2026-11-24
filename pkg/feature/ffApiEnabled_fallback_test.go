package feature

import (
	"context"
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
)

// nonReasonAwareProvider implements types.Provider but deliberately NOT
// types.ReasonAwareProvider, so it exercises the fallback branch in
// ApiEnabled.Value taken when the installed provider cannot report whether a
// targeting rule matched. Production always installs *providerGoFF (reason-aware),
// so this branch is unreachable today — this test pins its semantics so a future
// non-reason-aware provider is not silently wrong.
type nonReasonAwareProvider struct {
	boolValue bool
}

func (p nonReasonAwareProvider) BoolVariation(_ context.Context, _ string, _ bool) bool {
	return p.boolValue
}
func (p nonReasonAwareProvider) IntVariation(_ context.Context, _ string, d int) int { return d }
func (p nonReasonAwareProvider) Float64Variation(_ context.Context, _ string, d float64) float64 {
	return d
}
func (p nonReasonAwareProvider) StringVariation(_ context.Context, _ string, d string) string {
	return d
}
func (p nonReasonAwareProvider) JSONArrayVariation(_ context.Context, _ string, d []any) []any {
	return d
}
func (p nonReasonAwareProvider) JSONVariation(_ context.Context, _ string, d map[string]any) map[string]any {
	return d
}

// double-waiver: B1 — providerGoFF holds an out-of-process file/HTTP retriever handle; a fake is
// required because the real type always implements ReasonAwareProvider, so the else-branch is
// unreachable with it.
var _ types.Provider = nonReasonAwareProvider{}

func TestApiEnabledNonReasonAwareProviderFallback(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)

	// Guard: if the fake ever starts satisfying ReasonAwareProvider, this test would
	// silently stop exercising the fallback branch it exists to pin.
	if _, ok := any(nonReasonAwareProvider{}).(types.ReasonAwareProvider); ok {
		t.Fatal("nonReasonAwareProvider must not implement types.ReasonAwareProvider")
	}

	gcp := string(cloudcontrolv1beta1.ProviderGCP)

	cases := []struct {
		name       string
		feature    string
		brokerPlan string
		boolValue  bool
		want       bool
	}{
		{"foundation floor applies when provider returns false", types.FeatureFoundation, "", false, true},
		{"unknown floor applies when provider returns false", types.FeatureUnknown, "", false, true},
		{"non-foundation feature follows provider value false", types.FeatureNfs, "", false, false},
		{"non-foundation feature follows provider value true", types.FeatureNfs, "", true, true},
		// Documents the fallback branch's limitation vs the reason-aware path: with no
		// reason signal byRule is always false, so a config deny (e.g. trial) does NOT
		// suppress the foundation floor here. The production providerGoFF path suppresses
		// it via byRule=true — that divergence is why this branch is kept out of production.
		{"foundation floor is not suppressed on trial in the fallback path", types.FeatureFoundation, "trial", false, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			provider = nonReasonAwareProvider{boolValue: tc.boolValue}
			b := ContextBuilderFromCtx(ctx).Feature(tc.feature).Provider(gcp)
			if tc.brokerPlan != "" {
				b = b.BrokerPlan(tc.brokerPlan)
			}
			assert.Equal(t, tc.want, ApiEnabled.Value(b.Build(ctx)))
		})
	}
}
