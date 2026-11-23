package feature

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedApiEnabledMatrix returns the authoritative expected value for ApiEnabled.Value
// under the provider-scoped GA allowlist loaded from the embedded ff_ga.yaml.
//
// Matrix:
//   - trial → false for everything (trial disables all, including foundation/unknown)
//   - non-trial, foundation/unknown → true (always-on hardcode, independent of config)
//   - non-trial, nfs:          gcp, aws
//   - non-trial, peering:      gcp, aws, azure
//   - non-trial, redis:        gcp, aws, azure
//   - everything else → false
func expectedApiEnabledMatrix(feature, provider, brokerPlan string) bool {
	if brokerPlan == brokerPlanTrial {
		return false
	}
	switch feature {
	case types.FeatureFoundation, types.FeatureUnknown:
		return true
	case types.FeatureNfs:
		return provider == string(cloudcontrolv1beta1.ProviderGCP) ||
			provider == string(cloudcontrolv1beta1.ProviderAws)
	case types.FeaturePeering, types.FeatureRedis:
		return provider == string(cloudcontrolv1beta1.ProviderGCP) ||
			provider == string(cloudcontrolv1beta1.ProviderAws) ||
			provider == string(cloudcontrolv1beta1.ProviderAzure)
	default:
		return false
	}
}

// TestApiEnabledMatrix is the authoritative exact-matrix test for the provider-scoped
// GA allowlist in ff_ga.yaml. It drives every combination of AllFeatureNames ×
// AllProviders × {non-trial, trial} (= 110 rows) against the embedded config and
// asserts ApiEnabled.Value(ctx) matches expectedApiEnabledMatrix.
//
// This test does not depend on the removed parity machinery: it is the new oracle
// for the provider-scoped allowlist shape.
func TestApiEnabledMatrix(t *testing.T) {
	ctx := t.Context()
	restoreProvider(t)
	require.NoError(t, Initialize(ctx, logr.Discard(), WithEmbeddedRetriever()))

	features := types.AllFeatureNames()
	providers := cloudcontrolv1beta1.AllProviders()
	plans := []string{"", brokerPlanTrial} // non-trial first, then trial

	var corpusCount int
	for _, f := range features {
		for _, p := range providers {
			for _, bp := range plans {
				f, p, bp := f, string(p), bp
				corpusCount++

				label := bp
				if label == "" {
					label = "non-trial"
				}
				t.Run(fmt.Sprintf("%s/%s/%s", f, p, label), func(t *testing.T) {
					b := ContextBuilderFromCtx(ctx).Feature(f).Provider(p)
					if bp == brokerPlanTrial {
						b = b.BrokerPlan(bp)
					}
					c := b.Build(ctx)

					want := expectedApiEnabledMatrix(f, p, bp)
					assert.Equal(t, want, ApiEnabled.Value(c),
						"feature=%q provider=%q brokerPlan=%q: want %v", f, p, bp, want)
				})
			}
		}
	}

	// Corpus-size guard: a future shrink of AllFeatureNames or AllProviders changes
	// the total and forces an intentional update here (11 features × 5 providers × 2 plans = 110).
	assert.Equal(t, len(types.AllFeatureNames())*len(cloudcontrolv1beta1.AllProviders())*2, corpusCount,
		"corpus size changed — update expectation if AllFeatureNames or AllProviders grew/shrank")
	// Pinned total: catches a list shrink the dynamic guard above cannot (both sides move together).
	assert.Equal(t, 110, corpusCount,
		"expected 110 cells; update this constant and the oracle when a feature or provider is added/removed")
}
