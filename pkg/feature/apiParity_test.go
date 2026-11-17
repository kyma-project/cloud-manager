package feature

import (
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParityHolds asserts zero divergence against testdata/apiEnabled.yaml,
// which carries structurally-mirrored apiDisabled and apiEnabled blocks.
func TestParityHolds(t *testing.T) {
	divergences, err := CheckApiParity(t.Context(), "testdata/apiEnabled.yaml")
	require.NoError(t, err)
	assert.Empty(t, divergences,
		"expected zero divergences against apiEnabled.yaml but got %d:\n%v",
		len(divergences), divergences)
}

// TestParityEquivalentToLegacy iterates the full corpus derived from
// testdata/apiEnabled.yaml and asserts !ApiDisabled == ApiEnabled for every
// entry (AC7: both-code-paths-equivalent property).
func TestParityEquivalentToLegacy(t *testing.T) {
	ctx := t.Context()

	err := Initialize(ctx, logr.Discard(), WithFile("testdata/apiEnabled.yaml"))
	require.NoError(t, err)

	configBytes, err := os.ReadFile("testdata/apiEnabled.yaml")
	require.NoError(t, err)

	dims, err := buildParityDimensions(configBytes)
	require.NoError(t, err)
	corpus := buildParityCorpus(ctx, dims)

	for _, e := range corpus {
		e := e
		t.Run(e.Feature+"/"+e.Provider+"/"+e.BrokerPlan+"/"+e.Landscape+"/"+e.GlobalAccount, func(t *testing.T) {
			enabled := ApiEnabled.Value(e.ctx)
			notDisabled := !ApiDisabled.Value(e.ctx)
			assert.Equal(t, notDisabled, enabled,
				"AC7: !ApiDisabled must equal ApiEnabled for feature=%q provider=%q brokerPlan=%q landscape=%q globalAccount=%q shoot=%q",
				e.Feature, e.Provider, e.BrokerPlan, e.Landscape, e.GlobalAccount, e.Shoot)
		})
	}
}

// TestParityDivergentDetected asserts that CheckApiParity returns a non-empty
// divergence set for testdata/apiParityDivergent.yaml, and specifically that
// a configurable feature appears in the divergence set — the divergent fixture
// omits the stage-exemption for the 6034adc9 global account from apiEnabled,
// causing peering/nfsBackup/redis on stage with that account to diverge.
// Foundation and unknown are hardcoded always-on and will never appear in divergences.
func TestParityDivergentDetected(t *testing.T) {
	divergences, err := CheckApiParity(t.Context(), "testdata/apiParityDivergent.yaml")
	require.NoError(t, err)
	require.NotEmpty(t, divergences,
		"expected divergences to be detected in apiParityDivergent.yaml but got none")

	const exemptAccount = "6034adc9-3441-494c-b438-a0eba4c5045a"
	var foundConfigurableInContext bool
	for _, d := range divergences {
		switch d.Feature {
		case types.FeatureNfsBackup, types.FeaturePeering, types.FeatureRedis:
			if d.Landscape == types.LandscapeStage && d.GlobalAccount == exemptAccount {
				foundConfigurableInContext = true
			}
		}
	}
	assert.True(t, foundConfigurableInContext,
		"divergence for a configurable GA feature (nfsBackup/peering/redis) on stage with account "+
			exemptAccount+" not detected — the removed stage-exemption in apiEnabled must diverge in exactly that context")
}

// TestParityGlobalAccountSelfSeeding verifies two things:
//  1. The 6034adc9 account from apiEnabled.yaml's stage-exemption query is
//     extracted by self-seeding and appears in the corpus dimensions.
//  2. CheckApiParity on the divergent fixture (which removes the stage-exemption
//     from apiEnabled but keeps it in apiDisabled) detects divergences that
//     involve that specific globalAccount on stage, proving the corpus covered it.
func TestParityGlobalAccountSelfSeeding(t *testing.T) {
	const targetAccount = "6034adc9-3441-494c-b438-a0eba4c5045a"

	// Part 1: seed extraction from apiEnabled.yaml
	configBytes, err := os.ReadFile("testdata/apiEnabled.yaml")
	require.NoError(t, err)

	dims, err := buildParityDimensions(configBytes)
	require.NoError(t, err)

	var accountInCorpus bool
	for _, ga := range dims.GlobalAccounts {
		if ga == targetAccount {
			accountInCorpus = true
			break
		}
	}
	assert.True(t, accountInCorpus,
		"self-seeding must include %s from apiEnabled.yaml's stage-exemption query", targetAccount)

	// Part 2: divergence detection proves corpus covered the account.
	// The divergent fixture removes the stage-exemption from apiEnabled, so
	// contexts with globalAccount=6034adc9 on stage disagree between the two flags.
	// This divergence is only detectable if the account is in the corpus.
	divergences, err := CheckApiParity(t.Context(), "testdata/apiParityDivergent.yaml")
	require.NoError(t, err)
	require.NotEmpty(t, divergences, "expected divergences in divergent fixture")

	var foundAccountDivergence bool
	for _, d := range divergences {
		if d.GlobalAccount == targetAccount && d.Landscape == types.LandscapeStage {
			foundAccountDivergence = true
			break
		}
	}
	assert.True(t, foundAccountDivergence,
		"self-seeding must cause the corpus to cover account=%s on stage so the missing stage-exemption is detected",
		targetAccount)
}

// TestParityHoldsForEmbeddedGA is the in-repo parity gate: the shipped ff_ga.yaml
// must have zero apiEnabled/apiDisabled divergence over the full corpus. A regression
// introduced directly into ff_ga.yaml fails here.
func TestParityHoldsForEmbeddedGA(t *testing.T) {
	divergences, err := CheckApiParity(t.Context(), "ff_ga.yaml")
	require.NoError(t, err)
	assert.Empty(t, divergences, "ff_ga.yaml must have zero parity divergences: %v", divergences)
}

// TestParityHoldsForEmbeddedEdge is the in-repo parity gate for ff_edge.yaml.
func TestParityHoldsForEmbeddedEdge(t *testing.T) {
	divergences, err := CheckApiParity(t.Context(), "ff_edge.yaml")
	require.NoError(t, err)
	assert.Empty(t, divergences, "ff_edge.yaml must have zero parity divergences: %v", divergences)
}
