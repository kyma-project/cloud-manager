package feature

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustRAP returns the package provider as ReasonAwareProvider, failing the test if the
// assertion fails. Called after Initialize so the provider is *providerGoFF.
func mustRAP(t *testing.T) types.ReasonAwareProvider {
	t.Helper()
	rap, ok := provider.(types.ReasonAwareProvider)
	require.True(t, ok, "provider must implement ReasonAwareProvider after Initialize")
	return rap
}

// TestBoolVariationReason_TargetingMatchDisable: trial rule (rule 1 in apiEnabled.yaml)
// is a pure targeting rule → reason TARGETING_MATCH → byRule true, value false.
func TestBoolVariationReason_TargetingMatchDisable(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)
	require.NoError(t, Initialize(ctx, logr.Discard(), WithFile("testdata/apiEnabled.yaml")))

	rap := mustRAP(t)
	evalCtx := ContextBuilderFromCtx(ctx).BrokerPlan("trial").Feature(types.FeatureNfs).Build(ctx)
	value, byRule := rap.BoolVariationReason(evalCtx, "apiEnabled", true)

	assert.False(t, value, "trial rule disables the flag")
	assert.True(t, byRule, "a targeting rule match must set byRule=true")
}

// TestBoolVariationReason_TargetingMatchEnable: globalAccount exemption rule (rule 2)
// matches → reason TARGETING_MATCH → byRule true, value true.
func TestBoolVariationReason_TargetingMatchEnable(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)
	require.NoError(t, Initialize(ctx, logr.Discard(), WithFile("testdata/apiEnabled.yaml")))

	rap := mustRAP(t)
	evalCtx := ContextBuilderFromCtx(ctx).
		Landscape("stage").
		GlobalAccount("6034adc9-3441-494c-b438-a0eba4c5045a").
		Feature(types.FeatureNfsBackup).
		Build(ctx)
	value, byRule := rap.BoolVariationReason(evalCtx, "apiEnabled", false)

	assert.True(t, value, "globalAccount exemption enables the flag")
	assert.True(t, byRule, "a targeting rule match must set byRule=true")
}

// TestBoolVariationReason_DefaultRule: no targeting rule matches → reason DEFAULT →
// byRule false, value equals the default rule variation (enabled = true).
func TestBoolVariationReason_DefaultRule(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)
	require.NoError(t, Initialize(ctx, logr.Discard(), WithFile("testdata/apiEnabled.yaml")))

	rap := mustRAP(t)
	// landscape=dev, feature=nfs: none of the three targeting rules match
	evalCtx := ContextBuilderFromCtx(ctx).Landscape("dev").Feature(types.FeatureNfs).Build(ctx)
	value, byRule := rap.BoolVariationReason(evalCtx, "apiEnabled", false)

	assert.True(t, value, "default rule variation is enabled (true)")
	assert.False(t, byRule, "DEFAULT reason must not set byRule")
}

// TestBoolVariationReason_UnknownFlagError: flag key absent from loaded config →
// GoFF returns an error → method returns (defaultValue, false).
func TestBoolVariationReason_UnknownFlagError(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)
	require.NoError(t, Initialize(ctx, logr.Discard(), WithFile("testdata/apiEnabled.yaml")))

	rap := mustRAP(t)
	evalCtx := ContextBuilderFromCtx(ctx).Build(ctx)
	value, byRule := rap.BoolVariationReason(evalCtx, "nonExistentFlag_xyzzy", true)

	assert.True(t, value, "error path must return the caller-supplied defaultValue")
	assert.False(t, byRule, "error path must return byRule=false")
}

// TestBoolVariationReason_TargetingMatchSplit: progressive-rollout targeting rule
// (providerGoFF_split.yaml) with end date in the past → all users get "on" → reason
// TARGETING_MATCH_SPLIT → byRule true, value true.
func TestBoolVariationReason_TargetingMatchSplit(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)
	require.NoError(t, Initialize(ctx, logr.Discard(), WithFile("testdata/providerGoFF_split.yaml")))

	rap := mustRAP(t)
	// feature=nfs triggers the targeting rule
	evalCtx := ContextBuilderFromCtx(ctx).Feature(types.FeatureNfs).Build(ctx)
	value, byRule := rap.BoolVariationReason(evalCtx, "splitFlag", false)

	// Progressive rollout ended in 2021 → 100% of users receive "on" (true)
	assert.True(t, value, "post-rollout progressive rule delivers the end variation")
	assert.True(t, byRule, "TARGETING_MATCH_SPLIT must set byRule=true")
}
