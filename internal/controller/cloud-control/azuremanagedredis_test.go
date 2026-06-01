package cloudcontrol

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redisenterprise/armredisenterprise/v3"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	azurecommon "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/common"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	kcpsubscription "github.com/kyma-project/cloud-manager/pkg/kcp/subscription"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP AzureManagedRedis", func() {

	It("Scenario: KCP AzureManagedRedis is created and deleted", func() {

		Skip("fix me, I', flaky....")

		name := "a3b1c2d4-e5f6-7a8b-9c0d-e1f2a3b4c5d6"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeAzure).
				WithArguments(infra.Ctx(), infra, scope, WithName(name)).
				Should(Succeed())
		})

		azureMock := infra.AzureMock().MockConfigs(scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.TenantId)
		resourceGroupName := azurecommon.AzureCloudManagerResourceGroupName(scope.Spec.Scope.Azure.VpcNetwork)

		subscription := &cloudcontrolv1beta1.Subscription{}

		By("And Given Azure Subscription exists", func() {
			kcpsubscription.Ignore.AddName(name)
			Expect(
				CreateSubscription(infra.Ctx(), infra, subscription,
					WithName(name),
					WithSubscriptionSpecGarden("binding-name")),
			).To(Succeed())

			Expect(
				SubscriptionPatchStatusReadyAzure(infra.Ctx(), infra, subscription,
					scope.Spec.Scope.Azure.TenantId, scope.Spec.Scope.Azure.SubscriptionId),
			).To(Succeed())
		})

		vpcNetwork := &cloudcontrolv1beta1.VpcNetwork{}

		By("And Given KCP VpcNetwork exists in Ready state", func() {
			vpcNetworkName := scope.Spec.Scope.Azure.VpcNetwork
			vpcNetwork = cloudcontrolv1beta1.NewVpcNetworkBuilder().
				WithName(name).
				WithVpcNetworkName(&vpcNetworkName).
				WithRegion(scope.Spec.Region).
				WithSubscription(name).
				WithCidrBlocks("10.250.0.0/22").
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcNetwork).
				Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcNetwork,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed(), "Expected KCP VpcNetwork to become ready")
		})

		kcpIpRangeName := "5b9d8b8a-2b8a-4f0e-9c2a-9b1c1d1e1f1a"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		By("And Given KCP IpRange exists in Ready state", func() {
			kcpiprange.Ignore.AddName(kcpIpRangeName)
			Eventually(CreateKcpIpRange).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithKcpIpRangeRemoteRef("amr-iprange"),
					WithKcpIpRangeNetwork(name),
					WithScope(name),
				).
				Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		azureManagedRedis := &cloudcontrolv1beta1.AzureManagedRedis{}

		By("When KCP AzureManagedRedis is created", func() {
			Eventually(CreateKcpAzureManagedRedis).
				WithArguments(infra.Ctx(), infra.KCP().Client(), azureManagedRedis,
					WithName(name),
					WithRemoteRef("skr-amr-example"),
					WithKcpAzureManagedRedisVpcNetwork(name),
					WithIpRange(kcpIpRangeName),
					WithKcpAzureManagedRedisSKU(armredisenterprise.SKUNameBalancedB5),
					WithKcpAzureManagedRedisClusteringPolicy(armredisenterprise.ClusteringPolicyEnterpriseCluster),
					WithKcpAzureManagedRedisHighAvailability(true),
				).
				Should(Succeed(), "failed creating AzureManagedRedis")
		})

		By("Then KCP AzureManagedRedis has status.id", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), azureManagedRedis,
					NewObjActions(),
					HavingFieldSet("status", "id")).
				Should(Succeed(), "expected AzureManagedRedis to get status.id")
		})

		By("And Then KCP AzureManagedRedis has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), azureManagedRedis,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed(), "expected AzureManagedRedis to have Ready condition")
		})

		By("And Then KCP AzureManagedRedis has status state Ready", func() {
			Expect(azureManagedRedis.Status.State).To(Equal(string(cloudcontrolv1beta1.StateReady)))
		})

		By("And Then KCP AzureManagedRedis has .status.primaryEndpoint set", func() {
			Expect(azureManagedRedis.Status.PrimaryEndpoint).NotTo(BeEmpty())
		})

		By("And Then KCP AzureManagedRedis has .status.port set to 10000", func() {
			Expect(azureManagedRedis.Status.Port).To(Equal(int32(10000)))
		})

		By("And Then KCP AzureManagedRedis has .status.authString set", func() {
			Expect(azureManagedRedis.Status.AuthString).NotTo(BeEmpty())
		})

		By("And Then Private End Point is created", func() {
			pep, err := azureMock.GetPrivateEndPoint(infra.Ctx(), resourceGroupName, name+"-pe")
			Expect(err).ToNot(HaveOccurred())
			Expect(pep).NotTo(BeNil())
		})

		By("And Then Private Dns Zone Group is created", func() {
			dzg, err := azureMock.GetPrivateDnsZoneGroup(infra.Ctx(), resourceGroupName, name+"-pe", name+"-dzg")
			Expect(err).ToNot(HaveOccurred())
			Expect(dzg).ToNot(BeNil())
		})

		// DELETE

		By("When KCP AzureManagedRedis is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), azureManagedRedis).
				Should(Succeed(), "failed deleting AzureManagedRedis")
		})

		By("Then KCP AzureManagedRedis does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), azureManagedRedis).
				Should(Succeed(), "expected AzureManagedRedis not to exist (be deleted), but it still exists")
		})
	})

})
