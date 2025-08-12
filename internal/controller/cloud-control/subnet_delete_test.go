package cloudcontrol

import (
	"fmt"

	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	kcprediscluster "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/rediscluster"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Feature: KCP GcpSubnet deletion with dependant objects", func() {

	It("Scenario: KCP GcpSubnet is deleted with existing GcpRedisCluster", func() {
		const (
			kymaName         = "7cf8b791-f8a7-4f6f-bef3-cb046b4c2002"
			vpcId            = "051172c5-8810-4d27-b7a1-d593a71ffc3d"
			vpcCidr          = "10.180.0.0/16"
			subnetName       = "27afd89b-fd62-44ba-87df-0973afbe9922"
			subnetCidr       = "10.181.0.0/16"
			redisClusterName = "73832baa-ffb0-4015-bbcd-4d5b6a109fff"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(kymaName)

			Expect(CreateScopeGcp(infra.Ctx(), infra, scope, WithName(kymaName))).
				To(Succeed())
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

		subnet := &cloudcontrolv1beta1.GcpSubnet{}

		By("And Given KCP GcpSubnet is created", func() {
			Expect(CreateKcpGcpSubnet(infra.Ctx(), infra.KCP().Client(), subnet,
				WithName(subnetName),
				WithKcpGcpSubnetRemoteRef("skr-gcp-subnet-123"),
				WithKcpGcpSubnetSpecCidr("10.20.60.0/24"),
				WithKcpGcpSubnetPurposePrivate(),
				WithScope(kymaName),
				WithKcpGcpSubnetSpecCidr(subnetCidr),
			)).
				To(Succeed())
		})

		By("And When GcpSubnet creation operation is started", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), subnet,
					NewObjActions(),
					HavingKcpGcpSubnetCreationOperationDefined(),
				).
				Should(Succeed())
		})

		By("And When GcpSubnet creation operation is done", func() {
			infra.GcpMock().SetRegionOperationDone(subnet.Status.SubnetCreationOperationName)
		})

		By("And Given KCP GcpSubnet has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), subnet,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		redisCluster := &cloudcontrolv1beta1.GcpRedisCluster{}

		By("And Given GcpRedisCluster using KCP GcpSubnet exists", func() {
			kcprediscluster.Ignore.AddName(redisClusterName)
			Expect(CreateKcpGcpRedisCluster(infra.Ctx(), infra.KCP().Client(), redisCluster,
				WithName(redisClusterName),
				WithRemoteRef("skr-gcprediscluster-example"),
				WithGcpSubnet(subnetName),
				WithScope(kymaName),
				WithGcpSubnet(subnetName),
				WithKcpGcpRedisClusterNodeType("REDIS_SHARED_CORE_NANO"),
				WithKcpGcpRedisClusterShardCount(3),
				WithKcpGcpRedisClusterReplicasPerShard(1),
				WithKcpGcpRedisClusterConfigs(map[string]string{
					"maxmemory-policy": "allkeys-lru",
				}),
			)).
				To(Succeed())
		})

		By("When GcpSubnet is marked for deletion", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), subnet)).
				To(Succeed())
		})

		By("Then GcpSubnet has warning state", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), subnet,
					NewObjActions(),
					HavingState(string(cloudcontrolv1beta1.StateWarning)),
				).
				Should(Succeed())
		})

		By("And Then GcpSubnet has DeleteWhileUsed Warning condition", func() {
			cond := meta.FindStatusCondition(subnet.Status.Conditions, cloudcontrolv1beta1.ConditionTypeWarning)
			Expect(cond).ToNot(BeNil(), fmt.Sprintf(
				"Expected Warning condition, but found: %v",
				pie.Map(subnet.Status.Conditions, func(c metav1.Condition) string {
					return c.Type
				}),
			))
			Expect(cond.Reason).To(Equal(cloudcontrolv1beta1.ReasonDeleteWhileUsed),
				fmt.Sprintf("Expected Reason to equal %s, but found %s", cloudcontrolv1beta1.ReasonDeleteWhileUsed, cond.Reason))
			Expect(cond.Status).To(Equal(metav1.ConditionTrue), fmt.Sprintf("Expected True status, but found: %s", cond.Status))
		})

		By("When GcpRedisCluster is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), redisCluster)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed())
		})

		By("Then GcpSubnet is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), subnet).
				Should(Succeed())
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
