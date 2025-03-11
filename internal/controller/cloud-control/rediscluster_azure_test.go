package cloudcontrol

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	azurecommon "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/common"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"k8s.io/utils/ptr"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	iprangePkg "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP RedisCluster", func() {

	It("Scenario: KCP Azure RedisCluster is created, modified and deleted", func() {

		name := "924a92cf-9e72-408d-a1e8-017a2fd8dcls"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			scopePkg.Ignore.AddName(name)

			Eventually(CreateScopeAzure).
				WithArguments(infra.Ctx(), infra, scope, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "ffc7ebcc-114e-4d68-948c-241405fd0cls"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		// Tell IpRange reconciler to ignore this kymaName
		iprangePkg.Ignore.AddName(kcpIpRangeName)
		By("And Given KCP IPRange exists", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithKcpIpRangeRemoteRef("some-remote-ref"),
					WithKcpIpRangeNetwork("kcpNetworkCm.Name"),
					WithScope(scope.Name),
				).
				Should(Succeed())
		})

		By("And Given KCP IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithConditions(KcpReadyCondition()),
				).WithTimeout(20*time.Second).WithPolling(200*time.Millisecond).
				Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		redisCluster := &cloudcontrolv1beta1.RedisCluster{}
		redisCapacity := 2
		shardCount := 3
		replicaCount := 5

		resourceGroupName := azurecommon.AzureCloudManagerResourceGroupName(scope.Spec.Scope.Azure.VpcNetwork)
		var redis *armredis.ResourceInfo
		azureMock := infra.AzureMock().MockConfigs(scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.TenantId)

		By("When KCP RedisCluster is created", func() {
			Eventually(CreateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithName(name),
					WithRemoteRef("skr-redis-example"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisClusterAzure(),
					WithClusterProperties(shardCount, replicaCount),
					WithClusterSKU(redisCapacity),
					WithKcpAzureRedisClusterVersion("6.0"),
				).
				Should(Succeed(), "failed creating RedisCluster")
		})

		By("Then Azure Redis is created", func() {
			Eventually(func() error {
				r, err := azureMock.GetRedisInstance(infra.Ctx(), resourceGroupName, name)
				if err != nil {
					return err
				}
				redis = r
				return nil
			}).Should(Succeed())
		})

		By("And Then Azure Redis has capacity as specified in KCP RedisCluster", func() {
			actualCapacity := ptr.Deref(redis.Properties.SKU.Capacity, int32(0))
			Expect(actualCapacity).To(Equal(int32(redisCapacity)))
		})

		By("And Then Azure Redis has family as specified in KCP RedisCluster", func() {
			actualFamily := string(ptr.Deref(redis.Properties.SKU.Family, ""))
			redisStandardFamilyTier := "P"
			Expect(actualFamily).To(Equal(redisStandardFamilyTier))

			redisStandardFamilyName := "Premium"
			actualFamilyName := string(ptr.Deref(redis.Properties.SKU.Name, ""))
			Expect(actualFamilyName).To(Equal(redisStandardFamilyName))
		})

		By("And Then Azure Redis CLuster has shard and replica count setup ", func() {
			shards := ptr.Deref(redis.Properties.ShardCount, int32(shardCount))
			Expect(shards).To(Equal(int32(shardCount)))

			replicas := ptr.Deref(redis.Properties.ReplicasPerPrimary, int32(replicaCount))
			Expect(replicas).To(Equal(int32(replicaCount)))
		})

		By("And Then Azure Redis has nonSSl port disabled ", func() {
			nonSSLPortEnabled := ptr.Deref(redis.Properties.EnableNonSSLPort, true)
			Expect(nonSSLPortEnabled).To(Equal(false))
		})

		By("When Azure Redis state is Succeeded", func() {
			err := azureMock.AzureSetRedisInstanceState(infra.Ctx(), resourceGroupName, name, armredis.ProvisioningStateSucceeded)
			Expect(err).ToNot(HaveOccurred())
		})

		By("Then KCP RedisCluster has status.id", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingRedisClusterStatusId()).
				Should(Succeed(), "expected RedisCluster to get status.id")
		})

		By("And Then KCP RedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed(), "expected RedisCluster to has Ready state, but it didn't")
		})

		By("And Then KCP RedisCluster has status state Ready", func() {
			Expect(redisCluster.Status.State).To(Equal(cloudcontrolv1beta1.StateReady))
		})

		By("And Then KCP RedisCluster has .status.primaryEndpoint set", func() {
			expected := fmt.Sprintf(
				"%s:%d",
				ptr.Deref(redis.Properties.HostName, ""),
				ptr.Deref(redis.Properties.SSLPort, 0),
			)
			Expect(redisCluster.Status.DiscoveryEndpoint).To(Equal(expected))
		})

		By("And Then KCP RedisCluster has .status.authString set", func() {
			keys, err := azureMock.GetRedisInstanceAccessKeys(infra.Ctx(), resourceGroupName, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(keys).To(HaveLen(2))
			Expect(redisCluster.Status.AuthString).To(Equal(keys[0]))
		})

		By("And Then Private End Point is created", func() {
			pep, err := azureMock.GetPrivateEndPoint(infra.Ctx(), resourceGroupName, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(pep).NotTo(BeNil())
		})

		By("And Then Private Dns Zone Group is created", func() {
			pep, err := azureMock.GetPrivateDnsZoneGroup(infra.Ctx(), resourceGroupName, name, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(pep).ToNot(BeNil())
		})

		// UPDATE

		//By("When KCP RedisCluster is updated", func() {
		//	Eventually(UpdateRedisCluster).
		//		WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
		//			WithName(name),
		//			WithRemoteRef("skr-redis-example"),
		//			WithIpRange(kcpIpRangeName),
		//			WithScope(name),
		//			WithRedisClusterAzure(),
		//			WithClusterSKU(redisCapacity),
		//			WithKcpAzureRedisClusterVersion("6.0"),
		//		).
		//		Should(Succeed(), "failed updating RedisCluster")
		//})

		// DELETE

		By("When KCP RedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "failed deleting RedisCluster")
		})

		By("Then Azure Redis is in Deleting state", func() {
			Eventually(func() error {
				r, err := azureMock.GetRedisInstance(infra.Ctx(), resourceGroupName, name)
				if err != nil {
					return err
				}
				if ptr.Deref(r.Properties.ProvisioningState, "") != armredis.ProvisioningStateDeleting {
					return fmt.Errorf("expected Azure Redis to be in Deleting state, but it was: %s", ptr.Deref(r.Properties.ProvisioningState, ""))
				}
				redis = r
				return nil
			})
		})

		By("When Azure Redis is deleted", func() {
			err := azureMock.AzureRemoveRedisInstance(infra.Ctx(), resourceGroupName, redisCluster.Name)
			Expect(err).ToNot(HaveOccurred())
		})

		By("Then KCP RedisCluster does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "expected RedisCluster not to exist (be deleted), but it still exists")
		})

		By("Then Private Dns Zone Group is deleted", func() {
			dnsZoneGroup, err := azureMock.GetPrivateDnsZoneGroup(infra.Ctx(), resourceGroupName, redisCluster.Name, redisCluster.Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(dnsZoneGroup).To(BeNil())
		})

		By("Then Private Private End Point is deleted", func() {
			privateEndPoint, err := azureMock.GetPrivateEndPoint(infra.Ctx(), resourceGroupName, redisCluster.Name)
			Expect(err).To(HaveOccurred())
			Expect(azuremeta.IsNotFound(err)).To(BeTrue())
			Expect(privateEndPoint).To(BeNil())
		})
	})

})
