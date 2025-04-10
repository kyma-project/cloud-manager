package cloudcontrol

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	azurecommon "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/common"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"k8s.io/utils/ptr"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP RedisInstance", func() {

	It("Scenario: KCP Azure RedisInstance is created and deleted", func() {

		name := "924a92cf-9e72-408d-a1e8-017a2fd8d42e"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeAzure).
				WithArguments(infra.Ctx(), infra, scope, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "ffc7ebcc-114e-4d68-948c-241405fd01b6"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		// Tell IpRange reconciler to ignore this kymaName
		kcpiprange.Ignore.AddName(kcpIpRangeName)
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

		redisInstance := &cloudcontrolv1beta1.RedisInstance{}
		redisCapacity := 2
		redisFamily := "S"

		resourceGroupName := azurecommon.AzureCloudManagerResourceGroupName(scope.Spec.Scope.Azure.VpcNetwork)
		var redis *armredis.ResourceInfo
		azureMock := infra.AzureMock().MockConfigs(scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.TenantId)

		By("When KCP RedisInstance is created", func() {
			Eventually(CreateRedisInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					WithName(name),
					WithRemoteRef("skr-redis-example"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisInstanceAzure(),
					WithSKU(redisCapacity, redisFamily),
					WithKcpAzureRedisVersion("6.0"),
				).
				Should(Succeed(), "failed creating RedisInstance")
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

		By("And Then Azure Redis has capacity as specified in KCP RedisInstance", func() {
			actualCapacity := ptr.Deref(redis.Properties.SKU.Capacity, int32(0))
			Expect(actualCapacity).To(Equal(int32(redisCapacity)))
		})

		By("And Then Azure Redis has family as specified in KCP RedisInstance", func() {
			actualFamily := string(ptr.Deref(redis.Properties.SKU.Family, ""))
			redisStandardFamilyTier := "C"
			Expect(actualFamily).To(Equal(redisStandardFamilyTier))

			redisStandardFamilyName := "Basic"
			actualFamilyName := string(ptr.Deref(redis.Properties.SKU.Name, ""))
			Expect(actualFamilyName).To(Equal(redisStandardFamilyName))
		})

		By("And Then Azure Redis has nonSSl port disabled ", func() {
			nonSSLPortEnabled := ptr.Deref(redis.Properties.EnableNonSSLPort, true)
			Expect(nonSSLPortEnabled).To(Equal(false))
		})

		By("When Azure Redis state is Succeeded", func() {
			err := azureMock.AzureSetRedisInstanceState(infra.Ctx(), resourceGroupName, name, armredis.ProvisioningStateSucceeded)
			Expect(err).ToNot(HaveOccurred())
		})

		By("Then KCP RedisInstance has status.id", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingRedisInstanceStatusId()).
				Should(Succeed(), "expected RedisInstance to get status.id")
		})

		By("And Then KCP RedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed(), "expected RedisInstance to has Ready state, but it didn't")
		})

		By("And Then KCP RedisInstance has status state Ready", func() {
			Expect(redisInstance.Status.State).To(Equal(cloudcontrolv1beta1.StateReady))
		})

		By("And Then KCP RedisInstance has .status.primaryEndpoint set", func() {
			expected := fmt.Sprintf(
				"%s:%d",
				ptr.Deref(redis.Properties.HostName, ""),
				ptr.Deref(redis.Properties.SSLPort, 0),
			)
			Expect(redisInstance.Status.PrimaryEndpoint).To(Equal(expected))
		})

		By("And Then KCP RedisInstance has .status.authString set", func() {
			keys, err := azureMock.GetRedisInstanceAccessKeys(infra.Ctx(), resourceGroupName, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(keys).To(HaveLen(2))
			Expect(redisInstance.Status.AuthString).To(Equal(keys[0]))
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

		// DELETE

		By("When KCP RedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "failed deleting RedisInstance")
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
			err := azureMock.AzureRemoveRedisInstance(infra.Ctx(), resourceGroupName, redisInstance.Name)
			Expect(err).ToNot(HaveOccurred())
		})

		By("Then KCP RedisInstance does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "expected RedisInstance not to exist (be deleted), but it still exists")
		})

		By("Then Private Dns Zone Group is deleted", func() {
			dnsZoneGroup, err := azureMock.GetPrivateDnsZoneGroup(infra.Ctx(), resourceGroupName, redisInstance.Name, redisInstance.Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(dnsZoneGroup).To(BeNil())
		})

		By("Then Private Private End Point is deleted", func() {
			privateEndPoint, err := azureMock.GetPrivateEndPoint(infra.Ctx(), resourceGroupName, redisInstance.Name)
			Expect(err).To(HaveOccurred())
			Expect(azuremeta.IsNotFound(err)).To(BeTrue())
			Expect(privateEndPoint).To(BeNil())
		})
	})

})
