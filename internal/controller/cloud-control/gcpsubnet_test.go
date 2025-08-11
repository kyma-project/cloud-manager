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
	"k8s.io/utils/ptr"
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

		By("And When GcpSubnet creation operation is started", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnet,
					NewObjActions(),
					HavingKcpGcpSubnetCreationOperationDefined(),
				).
				Should(Succeed())
		})

		By("And When GcpSubnet creation operation is done", func() {
			infra.GcpMock().SetRegionOperationDone(gcpSubnet.Status.SubnetCreationOperationName)
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
			Expect(ptr.Deref(subnet.Purpose, "")).To(Equal("PRIVATE"))
			Expect(ptr.Deref(subnet.PrivateIpGoogleAccess, false)).To(Equal(true))
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

		By("And Then GCP Private Subnet does not exist", func() {
			subnet, err := infra.GcpMock().GetSubnet(infra.Ctx(), client.GetSubnetRequest{
				ProjectId: scope.Spec.Scope.Gcp.Project,
				Name:      "cm-" + gcpSubnet.Name,
				Region:    scope.Spec.Region,
			})
			Expect(subnet).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

		By("And Then GCP Connection Policy does not exist", func() {
			policyName := fmt.Sprintf("projects/%s/locations/%s/serviceConnectionPolicies/cm-%s-%s-redis-cluster",
				scope.Spec.Scope.Gcp.Project, scope.Spec.Region, scope.Spec.Scope.Gcp.VpcNetwork, scope.Spec.Region,
			)
			connectionPolicy, err := infra.GcpMock().GetServiceConnectionPolicy(infra.Ctx(), policyName)
			Expect(connectionPolicy).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

	})

	It("Scenario: KCP GcpSubnet goes into error state when subnet operation fails", func() {
		const (
			kymaName      = "c8bdab81-c6f6-45f8-8ea0-d6f3dec18cdb"
			gcpSubnetName = "1c7a0058-5390-44cb-bd05-09f6afeb2267"
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
					WithKcpGcpSubnetSpecCidr("10.20.80.0/24"),
					WithKcpGcpSubnetPurposePrivate(),
					WithScope(kymaName),
				).
				Should(Succeed())
		})

		By("And When GcpSubnet creation operation is started", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnet,
					NewObjActions(),
					HavingKcpGcpSubnetCreationOperationDefined(),
				).
				Should(Succeed())
		})

		By("And When GcpSubnet creation operation has errors", func() {
			infra.GcpMock().SetRegionOperationError(gcpSubnet.Status.SubnetCreationOperationName)
		})

		By("Then KCP GcpSubnet has Error condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnet,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeError),
				).
				Should(Succeed())
		})

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

		By("And Then GCP Private Subnet does not exist", func() {
			subnet, err := infra.GcpMock().GetSubnet(infra.Ctx(), client.GetSubnetRequest{
				ProjectId: scope.Spec.Scope.Gcp.Project,
				Name:      "cm-" + gcpSubnet.Name,
				Region:    scope.Spec.Region,
			})
			Expect(subnet).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

		By("And Then GCP Connection Policy does not exist", func() {
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

		By("And When GcpSubnet creation operation is started", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnet,
					NewObjActions(),
					HavingKcpGcpSubnetCreationOperationDefined(),
				).
				Should(Succeed())
		})

		By("And When GcpSubnet creation operation is done", func() {
			infra.GcpMock().SetRegionOperationDone(gcpSubnet.Status.SubnetCreationOperationName)
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

	It("Scenario: All KCP GcpSubnets that use Connection Policy are deleted in parallel", func() {
		const (
			kymaName       = "0f08ca00-3e86-4c9c-8103-c38a205c09d4"
			gcpSubnetAName = "0eecbc5a-4be5-4c56-8c83-17138af340c9"
			gcpSubnetBName = "02279018-b06c-4193-aabc-27c2c714b411"
		)

		scope := &cloudcontrolv1beta1.Scope{}
		gcpSubnetA := &cloudcontrolv1beta1.GcpSubnet{}
		gcpSubnetB := &cloudcontrolv1beta1.GcpSubnet{}

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

		By("And Given KCP GcpSubnet A is created", func() {
			Eventually(CreateKcpGcpSubnet).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetA,
					WithName(gcpSubnetAName),
					WithKcpGcpSubnetRemoteRef(gcpSubnetAName),
					WithKcpGcpSubnetSpecCidr("10.20.60.0/24"),
					WithKcpGcpSubnetPurposePrivate(),
					WithScope(kymaName),
				).
				Should(Succeed())
		})

		By("And When GcpSubnet A creation operation is started", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetA,
					NewObjActions(),
					HavingKcpGcpSubnetCreationOperationDefined(),
				).
				Should(Succeed())
		})

		By("And When GcpSubnet A creation operation is done", func() {
			infra.GcpMock().SetRegionOperationDone(gcpSubnetA.Status.SubnetCreationOperationName)
		})

		By("And Given KCP GcpSubnet A has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetA,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Given KCP GcpSubnet B is created", func() {
			Eventually(CreateKcpGcpSubnet).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetB,
					WithName(gcpSubnetBName),
					WithKcpGcpSubnetRemoteRef(gcpSubnetBName),
					WithKcpGcpSubnetSpecCidr("10.20.64.0/24"),
					WithKcpGcpSubnetPurposePrivate(),
					WithScope(kymaName),
				).
				Should(Succeed())
		})

		By("And When GcpSubnet B creation operation is started", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetB,
					NewObjActions(),
					HavingKcpGcpSubnetCreationOperationDefined(),
				).
				Should(Succeed())
		})

		By("And When GcpSubnet B creation operation is done", func() {
			infra.GcpMock().SetRegionOperationDone(gcpSubnetB.Status.SubnetCreationOperationName)
		})

		By("And Given KCP GcpSubnet B has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetB,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Given GCP Connection Policy is created", func() {
			policyName := fmt.Sprintf("projects/%s/locations/%s/serviceConnectionPolicies/cm-%s-%s-rc",
				scope.Spec.Scope.Gcp.Project, scope.Spec.Region, scope.Spec.Scope.Gcp.VpcNetwork, scope.Spec.Region,
			)
			connectionPolicy, err := infra.GcpMock().GetServiceConnectionPolicy(infra.Ctx(), policyName)
			Expect(err).NotTo(HaveOccurred())
			Expect(connectionPolicy).NotTo(BeNil())
			Expect(connectionPolicy.PscConfig.Subnetworks).Should(HaveLen(2))
		})

		// Delete

		By("When KCP GcpSubnet A is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetA).
				Should(Succeed(), "failed deleting KCP GcpSubnet")
		})

		By("And When KCP GcpSubnet B is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetB).
				Should(Succeed(), "failed deleting KCP GcpSubnet")
		})

		By("Then KCP GcpSubnet A does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetA).
				Should(Succeed(), "expected KCP GcpSubnet to be deleted, but it exists")
		})

		By("And Then KCP GcpSubnet B does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetB).
				Should(Succeed(), "expected KCP GcpSubnet to be deleted, but it exists")
		})

		By("And Then GCP Private Subnet A does not exist", func() {
			subnet, err := infra.GcpMock().GetSubnet(infra.Ctx(), client.GetSubnetRequest{
				ProjectId: scope.Spec.Scope.Gcp.Project,
				Name:      "cm-" + gcpSubnetA.Name,
				Region:    scope.Spec.Region,
			})
			Expect(subnet).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

		By("And Then GCP Private Subnet B does not exist", func() {
			subnet, err := infra.GcpMock().GetSubnet(infra.Ctx(), client.GetSubnetRequest{
				ProjectId: scope.Spec.Scope.Gcp.Project,
				Name:      "cm-" + gcpSubnetB.Name,
				Region:    scope.Spec.Region,
			})
			Expect(subnet).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

		By("And Then GCP Connection Policy does not exist", func() {
			policyName := fmt.Sprintf("projects/%s/locations/%s/serviceConnectionPolicies/cm-%s-%s-redis-cluster",
				scope.Spec.Scope.Gcp.Project, scope.Spec.Region, scope.Spec.Scope.Gcp.VpcNetwork, scope.Spec.Region,
			)
			connectionPolicy, err := infra.GcpMock().GetServiceConnectionPolicy(infra.Ctx(), policyName)
			Expect(connectionPolicy).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

	})

	It("Scenario: Some KCP GcpSubnets that use Connection Policy are deleted in parallel", func() {
		const (
			kymaName       = "e9e0e0c5-a87e-424c-89e3-0e6f59529a54"
			gcpSubnetAName = "a0f95406-0716-44e1-bb94-ae5013f17e3b"
			gcpSubnetBName = "3b12c9a5-b587-440c-a476-784c7b449fa1"
			gcpSubnetCName = "c86be6e4-0e01-48c3-bbf5-2520c89b1f19"
		)

		scope := &cloudcontrolv1beta1.Scope{}
		gcpSubnetA := &cloudcontrolv1beta1.GcpSubnet{}
		gcpSubnetB := &cloudcontrolv1beta1.GcpSubnet{}
		gcpSubnetC := &cloudcontrolv1beta1.GcpSubnet{}

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

		By("And Given KCP GcpSubnet A is created", func() {
			Eventually(CreateKcpGcpSubnet).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetA,
					WithName(gcpSubnetAName),
					WithKcpGcpSubnetRemoteRef(gcpSubnetAName),
					WithKcpGcpSubnetSpecCidr("10.20.60.0/24"),
					WithKcpGcpSubnetPurposePrivate(),
					WithScope(kymaName),
				).
				Should(Succeed())
		})

		By("And When GcpSubnet A creation operation is started", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetA,
					NewObjActions(),
					HavingKcpGcpSubnetCreationOperationDefined(),
				).
				Should(Succeed())
		})

		By("And When GcpSubnet A creation operation is done", func() {
			infra.GcpMock().SetRegionOperationDone(gcpSubnetA.Status.SubnetCreationOperationName)
		})

		By("And Given KCP GcpSubnet A has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetA,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Given KCP GcpSubnet B is created", func() {
			Eventually(CreateKcpGcpSubnet).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetB,
					WithName(gcpSubnetBName),
					WithKcpGcpSubnetRemoteRef(gcpSubnetBName),
					WithKcpGcpSubnetSpecCidr("10.20.64.0/24"),
					WithKcpGcpSubnetPurposePrivate(),
					WithScope(kymaName),
				).
				Should(Succeed())
		})

		By("And When GcpSubnet B creation operation is started", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetB,
					NewObjActions(),
					HavingKcpGcpSubnetCreationOperationDefined(),
				).
				Should(Succeed())
		})

		By("And When GcpSubnet B creation operation is done", func() {
			infra.GcpMock().SetRegionOperationDone(gcpSubnetB.Status.SubnetCreationOperationName)
		})

		By("And Given KCP GcpSubnet B has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetB,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Given KCP GcpSubnet C is created", func() {
			Eventually(CreateKcpGcpSubnet).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetC,
					WithName(gcpSubnetCName),
					WithKcpGcpSubnetRemoteRef(gcpSubnetCName),
					WithKcpGcpSubnetSpecCidr("10.20.68.0/24"),
					WithKcpGcpSubnetPurposePrivate(),
					WithScope(kymaName),
				).
				Should(Succeed())
		})

		By("And When GcpSubnet C creation operation is started", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetC,
					NewObjActions(),
					HavingKcpGcpSubnetCreationOperationDefined(),
				).
				Should(Succeed())
		})

		By("And When GcpSubnet C creation operation is done", func() {
			infra.GcpMock().SetRegionOperationDone(gcpSubnetC.Status.SubnetCreationOperationName)
		})

		By("And Given KCP GcpSubnet C has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetC,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Given GCP Connection Policy is created", func() {
			policyName := fmt.Sprintf("projects/%s/locations/%s/serviceConnectionPolicies/cm-%s-%s-rc",
				scope.Spec.Scope.Gcp.Project, scope.Spec.Region, scope.Spec.Scope.Gcp.VpcNetwork, scope.Spec.Region,
			)
			connectionPolicy, err := infra.GcpMock().GetServiceConnectionPolicy(infra.Ctx(), policyName)
			Expect(err).NotTo(HaveOccurred())
			Expect(connectionPolicy).NotTo(BeNil())
			Expect(connectionPolicy.PscConfig.Subnetworks).To(HaveLen(3))
		})

		// Delete

		By("When KCP GcpSubnet A is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetA).
				Should(Succeed(), "failed deleting KCP GcpSubnet")
		})

		By("And When KCP GcpSubnet B is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetB).
				Should(Succeed(), "failed deleting KCP GcpSubnet")
		})

		By("Then KCP GcpSubnet A does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetA).
				Should(Succeed(), "expected KCP GcpSubnet to be deleted, but it exists")
		})

		By("And Then KCP GcpSubnet B does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetB).
				Should(Succeed(), "expected KCP GcpSubnet to be deleted, but it exists")
		})

		By("And Then GCP Private Subnet A does not exist", func() {
			subnet, err := infra.GcpMock().GetSubnet(infra.Ctx(), client.GetSubnetRequest{
				ProjectId: scope.Spec.Scope.Gcp.Project,
				Name:      "cm-" + gcpSubnetA.Name,
				Region:    scope.Spec.Region,
			})
			Expect(subnet).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

		By("And Then GCP Private Subnet B does not exist", func() {
			subnet, err := infra.GcpMock().GetSubnet(infra.Ctx(), client.GetSubnetRequest{
				ProjectId: scope.Spec.Scope.Gcp.Project,
				Name:      "cm-" + gcpSubnetB.Name,
				Region:    scope.Spec.Region,
			})
			Expect(subnet).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

		By("And Then GCP Connection Policy has only GcpSubnet C", func() {
			policyName := fmt.Sprintf("projects/%s/locations/%s/serviceConnectionPolicies/cm-%s-%s-rc",
				scope.Spec.Scope.Gcp.Project, scope.Spec.Region, scope.Spec.Scope.Gcp.VpcNetwork, scope.Spec.Region,
			)
			connectionPolicy, err := infra.GcpMock().GetServiceConnectionPolicy(infra.Ctx(), policyName)
			Expect(err).NotTo(HaveOccurred())
			Expect(connectionPolicy).NotTo(BeNil())
			Expect(connectionPolicy.PscConfig.Subnetworks).To(HaveLen(1))
			Expect(connectionPolicy.PscConfig.Subnetworks[0]).To(ContainSubstring(gcpSubnetCName))
		})

		By("When KCP GcpSubnet C is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetC).
				Should(Succeed(), "failed deleting KCP GcpSubnet")
		})

		By("Then KCP GcpSubnet C does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnetC).
				Should(Succeed(), "expected KCP GcpSubnet to be deleted, but it exists")
		})

		By("And Then GCP Private Subnet C does not exist", func() {
			subnet, err := infra.GcpMock().GetSubnet(infra.Ctx(), client.GetSubnetRequest{
				ProjectId: scope.Spec.Scope.Gcp.Project,
				Name:      "cm-" + gcpSubnetC.Name,
				Region:    scope.Spec.Region,
			})
			Expect(subnet).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

		By("And Then GCP Connection Policy does not exist", func() {
			policyName := fmt.Sprintf("projects/%s/locations/%s/serviceConnectionPolicies/cm-%s-%s-redis-cluster",
				scope.Spec.Scope.Gcp.Project, scope.Spec.Region, scope.Spec.Scope.Gcp.VpcNetwork, scope.Spec.Region,
			)
			connectionPolicy, err := infra.GcpMock().GetServiceConnectionPolicy(infra.Ctx(), policyName)
			Expect(connectionPolicy).To(BeNil())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
		})

	})

})
