package azuremanagedredis

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redisenterprise/armredisenterprise/v3"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
)

// TestTierToSpec_AllEnumValuesCovered guards against drift between the CRD
// enum (api/cloud-resources/v1beta1/azuremanagedredis_tier.go) and the
// tier→spec mapping. Every constant declared in the API must have a row here,
// or the SKR controller will reject the value at create time even though the
// CRD validation accepted it.
func TestTierToSpec_AllEnumValuesCovered(t *testing.T) {
	allTiers := []cloudresourcesv1beta1.AzureManagedRedisTier{
		cloudresourcesv1beta1.AzureManagedRedisTierS1,
		cloudresourcesv1beta1.AzureManagedRedisTierS2,
		cloudresourcesv1beta1.AzureManagedRedisTierS3,
		cloudresourcesv1beta1.AzureManagedRedisTierS4,
		cloudresourcesv1beta1.AzureManagedRedisTierS5,
		cloudresourcesv1beta1.AzureManagedRedisTierP1,
		cloudresourcesv1beta1.AzureManagedRedisTierP2,
		cloudresourcesv1beta1.AzureManagedRedisTierP3,
		cloudresourcesv1beta1.AzureManagedRedisTierP4,
		cloudresourcesv1beta1.AzureManagedRedisTierP5,
		cloudresourcesv1beta1.AzureManagedRedisTierC3,
		cloudresourcesv1beta1.AzureManagedRedisTierC4,
		cloudresourcesv1beta1.AzureManagedRedisTierC5,
		cloudresourcesv1beta1.AzureManagedRedisTierC6,
		cloudresourcesv1beta1.AzureManagedRedisTierC7,
	}

	for _, tier := range allTiers {
		t.Run(string(tier), func(t *testing.T) {
			spec, err := TierToSpec(tier)
			if err != nil {
				t.Fatalf("TierToSpec(%q) returned error: %v", tier, err)
			}
			if spec.SKU == "" {
				t.Errorf("tier %q expanded to empty SKU", tier)
			}
			if spec.ClusteringPolicy == "" {
				t.Errorf("tier %q expanded to empty ClusteringPolicy", tier)
			}
		})
	}
}

// TestTierToSpec_FamilyInvariants asserts the structural rules described in
// docs/user/resources/04-40-32-azure-managed-redis.md:
//   - S* are non-HA, EnterpriseCluster, Balanced family.
//   - P* are HA,     EnterpriseCluster, ComputeOptimized family.
//   - C* are HA,     OSSCluster,        ComputeOptimized family.
func TestTierToSpec_FamilyInvariants(t *testing.T) {
	cases := []struct {
		tier             cloudresourcesv1beta1.AzureManagedRedisTier
		wantHA           bool
		wantPolicy       armredisenterprise.ClusteringPolicy
		wantFamilyPrefix string
	}{
		{cloudresourcesv1beta1.AzureManagedRedisTierS1, false, armredisenterprise.ClusteringPolicyEnterpriseCluster, "Balanced_"},
		{cloudresourcesv1beta1.AzureManagedRedisTierS5, false, armredisenterprise.ClusteringPolicyEnterpriseCluster, "Balanced_"},
		{cloudresourcesv1beta1.AzureManagedRedisTierP1, true, armredisenterprise.ClusteringPolicyEnterpriseCluster, "ComputeOptimized_"},
		{cloudresourcesv1beta1.AzureManagedRedisTierP5, true, armredisenterprise.ClusteringPolicyEnterpriseCluster, "ComputeOptimized_"},
		{cloudresourcesv1beta1.AzureManagedRedisTierC3, true, armredisenterprise.ClusteringPolicyOSSCluster, "ComputeOptimized_"},
		{cloudresourcesv1beta1.AzureManagedRedisTierC7, true, armredisenterprise.ClusteringPolicyOSSCluster, "ComputeOptimized_"},
	}

	for _, c := range cases {
		t.Run(string(c.tier), func(t *testing.T) {
			spec, err := TierToSpec(c.tier)
			if err != nil {
				t.Fatalf("TierToSpec(%q) error: %v", c.tier, err)
			}
			if spec.HighAvailability != c.wantHA {
				t.Errorf("tier %q: HighAvailability=%v, want %v", c.tier, spec.HighAvailability, c.wantHA)
			}
			if spec.ClusteringPolicy != c.wantPolicy {
				t.Errorf("tier %q: ClusteringPolicy=%q, want %q", c.tier, spec.ClusteringPolicy, c.wantPolicy)
			}
			if !startsWith(string(spec.SKU), c.wantFamilyPrefix) {
				t.Errorf("tier %q: SKU=%q, want prefix %q", c.tier, spec.SKU, c.wantFamilyPrefix)
			}
		})
	}
}

// TestTierToSpec_PCParity verifies that P-tier and C-tier of the same numeric
// rank share the same SKU (the ComputeOptimized capacity is identical; only
// the ClusteringPolicy differs). This is the key invariant from the pricing
// doc — P{N+2} and C{N+2} cost the same.
func TestTierToSpec_PCParity(t *testing.T) {
	pairs := []struct {
		p cloudresourcesv1beta1.AzureManagedRedisTier
		c cloudresourcesv1beta1.AzureManagedRedisTier
	}{
		{cloudresourcesv1beta1.AzureManagedRedisTierP1, cloudresourcesv1beta1.AzureManagedRedisTierC3},
		{cloudresourcesv1beta1.AzureManagedRedisTierP2, cloudresourcesv1beta1.AzureManagedRedisTierC4},
		{cloudresourcesv1beta1.AzureManagedRedisTierP3, cloudresourcesv1beta1.AzureManagedRedisTierC5},
		{cloudresourcesv1beta1.AzureManagedRedisTierP4, cloudresourcesv1beta1.AzureManagedRedisTierC6},
		{cloudresourcesv1beta1.AzureManagedRedisTierP5, cloudresourcesv1beta1.AzureManagedRedisTierC7},
	}
	for _, pair := range pairs {
		t.Run(string(pair.p)+"_vs_"+string(pair.c), func(t *testing.T) {
			pSpec, err := TierToSpec(pair.p)
			if err != nil {
				t.Fatalf("TierToSpec(%q) error: %v", pair.p, err)
			}
			cSpec, err := TierToSpec(pair.c)
			if err != nil {
				t.Fatalf("TierToSpec(%q) error: %v", pair.c, err)
			}
			if pSpec.SKU != cSpec.SKU {
				t.Errorf("P/C parity broken: %s SKU=%q, %s SKU=%q", pair.p, pSpec.SKU, pair.c, cSpec.SKU)
			}
			if pSpec.ClusteringPolicy == cSpec.ClusteringPolicy {
				t.Errorf("P/C should differ in ClusteringPolicy: both got %q", pSpec.ClusteringPolicy)
			}
		})
	}
}

func TestTierToSpec_UnknownTier(t *testing.T) {
	_, err := TierToSpec(cloudresourcesv1beta1.AzureManagedRedisTier("X42"))
	if err == nil {
		t.Fatal("expected error for unknown tier, got nil")
	}
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
