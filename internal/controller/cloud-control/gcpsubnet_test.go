package cloudcontrol

import (
	"fmt"

	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	kcpgcprediscluster "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/rediscluster"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet/client"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Feature: KCP GcpSubnet is created", func() {

	It("Scenario: KCP GcpSubnet with specified CIDR is created and deleted", func() {
		const (
			kymaName      = "c8ca95d7-26ff-41da-8954-e532feb7151e"
			gcpSubnetName = "ab88f23c-07d6-49a9-8e41-4494ba68d486"
		)

		scope := &cloudcontrolv1beta1.Scope{}
		gcpSubnet := &cloudcontrolv1beta1.GcpSubnet{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeGcp).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		var kcpNetworkKyma *cloudcontrolv1beta1.Network

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(kymaName).
				WithName(common.KcpNetworkKymaCommonName(kymaName)).
				WithGcpRef(scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork).
				WithType(cloudcontrolv1beta1.NetworkTypeKyma).
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNetworkKyma).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNetworkKyma, NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed())
		})

		By("When KCP GcpSubnet is created", func() {
			Eventually(CreateKcpGcpSubnet).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnet,
					WithName(gcpSubnetName),
					WithKcpGcpSubnetRemoteRef(gcpSubnetName),
					WithKcpGcpSubnetSpecCidr("10.20.60.0/24"),
					WithKcpGcpSubnetPurposePrivate(),
					WithScope(kymaName),
				).
				Should(Succeed())
		})

		By("Then KCP GcpSubnet has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnet,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then KCP GcpSubnet has status cidr equals to spec cidr", func() {
			Expect(gcpSubnet.Status.Cidr).To(Equal(gcpSubnet.Spec.Cidr))
		})

		By("And Then GCP Private Subnet is created", func() {
			subnet, err := infra.GcpMock().GetSubnet(infra.Ctx(), client.GetSubnetRequest{
				ProjectId: scope.Spec.Scope.Gcp.Project,
				Name:      "cm-" + gcpSubnet.Name,
				Region:    scope.Spec.Region,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(subnet).NotTo(BeNil())
		})

		By("And Then GCP Connection Policy is created", func() {
			policyName := fmt.Sprintf("projects/%s/locations/%s/serviceConnectionPolicies/cm-%s-%s-rc",
				scope.Spec.Scope.Gcp.Project, scope.Spec.Region, scope.Spec.Scope.Gcp.VpcNetwork, scope.Spec.Region,
			)
			connectionPolicy, err := infra.GcpMock().GetServiceConnectionPolicy(infra.Ctx(), policyName)
			Expect(err).NotTo(HaveOccurred())
			Expect(connectionPolicy).NotTo(BeNil())
		})

		// Delete

		By("When KCP GcpSubnet is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnet).
				Should(Succeed(), "failed deleting KCP GcpSubnet")
		})

		By("Then KCP GcpSubnet does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnet).
				Should(Succeed(), "expected KCP GcpSubnet to be deleted, but it exists")
		})

		By("And Then GCP Connection Policy does not exist", func() {
			subnet, err := infra.GcpMock().GetSubnet(infra.Ctx(), client.GetSubnetRequest{
				ProjectId: scope.Spec.Scope.Gcp.Project,
				Name:      "cm-" + gcpSubnet.Name,
				Region:    scope.Spec.Region,
			})
			Expect(subnet).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

		By("And Then GCP Private Subnet does not exist", func() {
			policyName := fmt.Sprintf("projects/%s/locations/%s/serviceConnectionPolicies/cm-%s-%s-redis-cluster",
				scope.Spec.Scope.Gcp.Project, scope.Spec.Region, scope.Spec.Scope.Gcp.VpcNetwork, scope.Spec.Region,
			)
			connectionPolicy, err := infra.GcpMock().GetServiceConnectionPolicy(infra.Ctx(), policyName)
			Expect(connectionPolicy).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

	})

	It("Scenario: KCP GcpSubnet is deleted with existing GcpRedisCluster using it", func() {
		const (
			kymaName      = "5a54377a-44a7-45cd-8991-2ea6b90fb4bc"
			gcpSubnetName = "b39da7ab-fdbc-4ff1-9c76-f027909e0d64"
		)

		scope := &cloudcontrolv1beta1.Scope{}
		gcpSubnet := &cloudcontrolv1beta1.GcpSubnet{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeGcp).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		var kcpNetworkKyma *cloudcontrolv1beta1.Network

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(kymaName).
				WithName(common.KcpNetworkKymaCommonName(kymaName)).
				WithGcpRef(scope.Spec.Scope.Gcp.Project, scope.Spec.Scope.Gcp.VpcNetwork).
				WithType(cloudcontrolv1beta1.NetworkTypeKyma).
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNetworkKyma).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNetworkKyma, NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed())
		})

		By("And Given KCP GcpSubnet is created", func() {
			Eventually(CreateKcpGcpSubnet).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnet,
					WithName(gcpSubnetName),
					WithKcpGcpSubnetRemoteRef(gcpSubnetName),
					WithKcpGcpSubnetSpecCidr("10.20.60.0/24"),
					WithKcpGcpSubnetPurposePrivate(),
					WithScope(kymaName),
				).
				Should(Succeed())
		})

		By("And Given KCP GcpSubnet has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnet,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		gcpRedisClusterName := "06a66d67-b597-4cb2-a79f-fc1cb57160b1"
		gcpRedisCluster := &cloudcontrolv1beta1.GcpRedisCluster{}

		By("And Given GcpRedisCluster using KCP GcpSubnet exists", func() {
			kcpgcprediscluster.Ignore.AddName(gcpRedisClusterName)
			Expect(CreateKcpGcpRedisCluster(infra.Ctx(), infra.KCP().Client(), gcpRedisCluster,
				WithName(gcpRedisClusterName),
				WithRemoteRef("foo"),
				WithScope(kymaName),
				WithGcpSubnet(gcpSubnetName),
				WithKcpGcpRedisClusterNodeType("REDIS_SHARED_CORE_NANO"),
				WithKcpGcpRedisClusterShardCount(3),
				WithKcpGcpRedisClusterReplicasPerShard(1),
				WithKcpGcpRedisClusterConfigs(map[string]string{
					"maxmemory-policy": "allkeys-lru",
				}),
			)).
				To(Succeed())
		})

		// Delete

		By("When KCP GcpSubnet is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnet).
				Should(Succeed(), "failed deleting KCP GcpSubnet")
		})

		By("Then GcpSubnet has warning state", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnet,
					NewObjActions(),
					HavingState(string(cloudcontrolv1beta1.StateWarning)),
				).
				Should(Succeed())
		})

		By("And Then GcpSubnet has DeleteWhileUsed Warning condition", func() {
			cond := meta.FindStatusCondition(gcpSubnet.Status.Conditions, cloudcontrolv1beta1.ConditionTypeWarning)
			Expect(cond).ToNot(BeNil(), fmt.Sprintf(
				"Expected Warning condition, but found: %v",
				pie.Map(gcpSubnet.Status.Conditions, func(c metav1.Condition) string {
					return c.Type
				}),
			))
			Expect(cond.Reason).To(Equal(cloudcontrolv1beta1.ReasonDeleteWhileUsed),
				fmt.Sprintf("Expected Reason to equal %s, but found %s", cloudcontrolv1beta1.ReasonDeleteWhileUsed, cond.Reason))
			Expect(cond.Status).To(Equal(metav1.ConditionTrue), fmt.Sprintf("Expected True status, but found: %s", cond.Status))
		})

		By("When RedisCluster is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), gcpRedisCluster)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpRedisCluster).
				Should(Succeed())
		})

		By("Then KCP GcpSubnet is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnet).
				Should(Succeed(), "expected KCP GcpSubnet to be deleted, but it exists")
		})

		By("// cleanup: delete KCP Kyma Network", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), kcpNetworkKyma)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNetworkKyma).
				Should(Succeed())
		})

		By("// cleanup: delete Scope", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), scope)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})

	})

})
