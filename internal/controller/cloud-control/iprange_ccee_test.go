package cloudcontrol

import (
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP IpRange SAP", func() {

	It("Scenario: KCP SAP IpRange is created and deleted", func() {
		name := "dc01bde6-3012-4336-92a2-54deec85c0c6"
		scope := &cloudcontrolv1beta1.Scope{}

		sapMock := infra.SapMock().NewProject()

		By("Given OpenStack Scope exists", func() {
			// Tell Scope reconciler to ignore this Scope
			kcpscope.Ignore.AddName(name)

			Expect(CreateScopeOpenStack(infra.Ctx(), infra, scope, sapMock.ProviderParams(), WithName(name))).
				To(Succeed(), "failed creating Scope")
		})

		var kcpNetworkKyma *cloudcontrolv1beta1.Network

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(name).
				WithName(common.KcpNetworkKymaCommonName(name)).
				WithOpenStackRef(scope.Spec.Scope.OpenStack.DomainName, scope.Spec.Scope.OpenStack.TenantName, "", scope.Spec.Scope.OpenStack.VpcNetwork).
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

		var sapGardenInfra *SapGardenerInfra

		By("And Given SAP infra exists", func() {
			sapMock.AddNetwork(
				"wrong1-"+name,
				"wrong1-"+name,
			)

			sgi, err := CreateSapGardenerResources(infra.Ctx(), sapMock, infra.Garden().Namespace(), scope.Spec.ShootName, "10.250.0.0/16")
			Expect(err).NotTo(HaveOccurred())
			sapGardenInfra = sgi

			sapMock.AddNetwork(
				"wrong2-"+name,
				"wrong2-"+name,
			)
		})

		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		By("When KCP IpRange is created", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(name),
					WithKcpIpRangeRemoteRef("some-remote-ref"),
					WithKcpIpRangeNetwork(kcpNetworkKyma.Name),
					WithScope(name),
				).
				Should(Succeed())
		})

		By("Then KCP IpRange has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then KCP IpRange has allocated CIDR in status", func() {
			Expect(kcpIpRange.Status.Cidr).To(Equal("10.251.0.0/22"))
		})

		By("And Then KCP IpRange status has one subnet", func() {
			Expect(kcpIpRange.Status.Subnets).To(HaveLen(1))
		})

		var osSubnet *subnets.Subnet

		By("And Then KCP IpRange Openstack Subnet is created", func() {
			subnet, err := sapMock.GetSubnetByName(infra.Ctx(), sapGardenInfra.VPC.ID, kcpIpRange.Status.Subnets[0].Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(subnet).ToNot(BeNil())
			osSubnet = subnet
		})

		By("And Then KCP IpRange Openstack Subnet range equals to IpRange cidr", func() {
			Expect(osSubnet.CIDR).To(Equal(kcpIpRange.Status.Cidr))
			Expect(osSubnet.CIDR).To(Equal(kcpIpRange.Status.Subnets[0].Range))
		})

		By("And Then KCP IpRange Openstack Subnet is added to router", func() {
			arr, err := sapMock.ListRouterSubnetInterfaces(infra.Ctx(), sapGardenInfra.Router.ID)
			Expect(err).NotTo(HaveOccurred())

			found := false
			for _, ii := range arr {
				if ii.SubnetID == osSubnet.ID {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})

		// DELETE ==========================================================================

		By("When KCP IpRange is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), kcpIpRange)).
				To(Succeed())
		})

		By("Then KCP IpRange does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange).
				Should(Succeed())
		})

		By("And Then KCP IpRange Openstack Subnet does not exist", func() {
			subnet, err := sapMock.GetSubnetByName(infra.Ctx(), sapGardenInfra.VPC.ID, kcpIpRange.Status.Subnets[0].Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(subnet).To(BeNil())
		})

		By("And Then KCP IpRange Openstack Subnet is removed from the router", func() {
			arr, err := sapMock.ListRouterSubnetInterfaces(infra.Ctx(), sapGardenInfra.Router.ID)
			Expect(err).NotTo(HaveOccurred())

			found := false
			for _, ii := range arr {
				if ii.SubnetID == osSubnet.ID {
					found = true
					break
				}
			}
			Expect(found).To(BeFalse())
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
