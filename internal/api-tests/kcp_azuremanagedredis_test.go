package api_tests

import (
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testKcpAzureManagedRedisBuilder struct {
	instance cloudcontrolv1beta1.AzureManagedRedis
}

func newTestKcpAzureManagedRedisBuilder() *testKcpAzureManagedRedisBuilder {
	return &testKcpAzureManagedRedisBuilder{
		instance: cloudcontrolv1beta1.AzureManagedRedis{
			Spec: cloudcontrolv1beta1.AzureManagedRedisSpec{
				RemoteRef: cloudcontrolv1beta1.RemoteRef{
					Name:      uuid.NewString(),
					Namespace: "default",
				},
				Scope: cloudcontrolv1beta1.ScopeRef{
					Name: uuid.NewString(),
				},
				SKU:              "Balanced_B5",
				ClusteringPolicy: "EnterpriseCluster",
			},
		},
	}
}

func (b *testKcpAzureManagedRedisBuilder) Build() *cloudcontrolv1beta1.AzureManagedRedis {
	return &b.instance
}

func (b *testKcpAzureManagedRedisBuilder) WithSKU(sku string) *testKcpAzureManagedRedisBuilder {
	b.instance.Spec.SKU = sku
	return b
}

func (b *testKcpAzureManagedRedisBuilder) WithClusteringPolicy(policy string) *testKcpAzureManagedRedisBuilder {
	b.instance.Spec.ClusteringPolicy = policy
	return b
}

func (b *testKcpAzureManagedRedisBuilder) WithHighAvailability(ha bool) *testKcpAzureManagedRedisBuilder {
	b.instance.Spec.HighAvailability = ha
	return b
}

func (b *testKcpAzureManagedRedisBuilder) WithIpRangeName(name string) *testKcpAzureManagedRedisBuilder {
	b.instance.Spec.IpRange.Name = name
	return b
}

func (b *testKcpAzureManagedRedisBuilder) WithRemoteRef(name, namespace string) *testKcpAzureManagedRedisBuilder {
	b.instance.Spec.RemoteRef.Name = name
	b.instance.Spec.RemoteRef.Namespace = namespace
	return b
}

var _ = Describe("Feature: KCP AzureManagedRedis", Ordered, func() {

	Context("Scenario: SKU enum validation", func() {

		// A representative sample of valid SKUs across all four families.
		// We don't enumerate all 53 — the validator is a simple Enum check.
		validSamples := []string{
			"Balanced_B0",
			"Balanced_B5",
			"ComputeOptimized_X3",
			"ComputeOptimized_X100",
			"MemoryOptimized_M10",
			"MemoryOptimized_M2000",
			"FlashOptimized_A250",
			"FlashOptimized_A4500",
			"Enterprise_E1",
			"Enterprise_E400",
			"EnterpriseFlash_F300",
			"EnterpriseFlash_F1500",
		}
		for _, sku := range validSamples {
			canCreateKcp(
				"AzureManagedRedis with sku="+sku+" is accepted",
				newTestKcpAzureManagedRedisBuilder().WithSKU(sku),
			)
		}

		canNotCreateKcp(
			"AzureManagedRedis with unknown sku is rejected",
			newTestKcpAzureManagedRedisBuilder().WithSKU("NoSuchSKU_Z9"),
			"",
		)

		canNotCreateKcp(
			"AzureManagedRedis with empty sku is rejected",
			newTestKcpAzureManagedRedisBuilder().WithSKU(""),
			"",
		)
	})

	Context("Scenario: SKU immutability", func() {

		canNotChangeKcp(
			"AzureManagedRedis sku cannot be changed",
			newTestKcpAzureManagedRedisBuilder().WithSKU("Balanced_B5"),
			func(b Builder[*cloudcontrolv1beta1.AzureManagedRedis]) {
				b.(*testKcpAzureManagedRedisBuilder).WithSKU("Balanced_B10")
			},
			"sku is immutable",
		)
	})

	Context("Scenario: ClusteringPolicy enum validation", func() {

		for _, policy := range []string{"EnterpriseCluster", "NoCluster", "OSSCluster"} {
			canCreateKcp(
				"AzureManagedRedis with clusteringPolicy="+policy+" is accepted",
				newTestKcpAzureManagedRedisBuilder().WithClusteringPolicy(policy),
			)
		}

		canNotCreateKcp(
			"AzureManagedRedis with unknown clusteringPolicy is rejected",
			newTestKcpAzureManagedRedisBuilder().WithClusteringPolicy("Invalid"),
			"",
		)

		canNotCreateKcp(
			"AzureManagedRedis with empty clusteringPolicy is rejected",
			newTestKcpAzureManagedRedisBuilder().WithClusteringPolicy(""),
			"",
		)
	})

	Context("Scenario: ClusteringPolicy immutability", func() {

		canNotChangeKcp(
			"AzureManagedRedis clusteringPolicy cannot be changed",
			newTestKcpAzureManagedRedisBuilder().WithClusteringPolicy("EnterpriseCluster"),
			func(b Builder[*cloudcontrolv1beta1.AzureManagedRedis]) {
				b.(*testKcpAzureManagedRedisBuilder).WithClusteringPolicy("OSSCluster")
			},
			"ClusteringPolicy is immutable",
		)
	})

	Context("Scenario: HighAvailability immutability", func() {

		canCreateKcp(
			"AzureManagedRedis with highAvailability=true is accepted",
			newTestKcpAzureManagedRedisBuilder().WithHighAvailability(true),
		)

		canCreateKcp(
			"AzureManagedRedis with highAvailability=false is accepted",
			newTestKcpAzureManagedRedisBuilder().WithHighAvailability(false),
		)

		canNotChangeKcp(
			"AzureManagedRedis highAvailability cannot be flipped on",
			newTestKcpAzureManagedRedisBuilder().WithHighAvailability(false),
			func(b Builder[*cloudcontrolv1beta1.AzureManagedRedis]) {
				b.(*testKcpAzureManagedRedisBuilder).WithHighAvailability(true)
			},
			"HighAvailability is immutable",
		)
	})

	Context("Scenario: RemoteRef immutability", func() {

		canNotChangeKcp(
			"AzureManagedRedis remoteRef cannot be changed",
			newTestKcpAzureManagedRedisBuilder().WithRemoteRef("orig-name", "orig-ns"),
			func(b Builder[*cloudcontrolv1beta1.AzureManagedRedis]) {
				b.(*testKcpAzureManagedRedisBuilder).WithRemoteRef("new-name", "orig-ns")
			},
			"RemoteRef is immutable",
		)
	})

	Context("Scenario: IpRange optional + immutable", func() {

		canCreateKcp(
			"AzureManagedRedis can be created without ipRange",
			newTestKcpAzureManagedRedisBuilder(),
		)

		canCreateKcp(
			"AzureManagedRedis can be created with ipRange",
			newTestKcpAzureManagedRedisBuilder().WithIpRangeName("my-iprange"),
		)

		canNotChangeKcp(
			"AzureManagedRedis ipRange cannot be changed once set",
			newTestKcpAzureManagedRedisBuilder().WithIpRangeName("orig-iprange"),
			func(b Builder[*cloudcontrolv1beta1.AzureManagedRedis]) {
				b.(*testKcpAzureManagedRedisBuilder).WithIpRangeName("new-iprange")
			},
			"IpRange is immutable",
		)

		canNotChangeKcp(
			"AzureManagedRedis ipRange cannot be added after creation",
			newTestKcpAzureManagedRedisBuilder(),
			func(b Builder[*cloudcontrolv1beta1.AzureManagedRedis]) {
				b.(*testKcpAzureManagedRedisBuilder).WithIpRangeName("late-iprange")
			},
			"IpRange is immutable",
		)
	})
})
