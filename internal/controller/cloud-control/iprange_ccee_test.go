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

		By("Given OpenStack Scope exists", func() {
			// Tell Scope reconciler to ignore this Scope
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeOpenStack).
				WithArguments(infra.Ctx(), infra, scope, WithName(name)).
				Should(Succeed(), "failed creating Scope")
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

		sapMock := infra.SapMock()

		networkId := "eb9cc44d-50a6-4405-84f8-acebb4182ef6"

		By("And Given SAP network exists", func() {
			sapMock.AddNetwork(
				"wrong1",
				"wrong1",
			)
			sapMock.AddNetwork(
				networkId,
				scope.Spec.Scope.OpenStack.VpcNetwork,
			)
			sapMock.AddNetwork(
				"wrong2",
				"wrong2",
			)

			_, err := infra.SapMock().CreateSubnet(
				infra.Ctx(),
				networkId,
				scope.Spec.Scope.OpenStack.Network.Nodes, // 10.250.0.0/16
				scope.Spec.Scope.OpenStack.VpcNetwork,
			)
			Expect(err).NotTo(HaveOccurred())
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
			subnet, err := sapMock.GetSubnetByName(infra.Ctx(), networkId, kcpIpRange.Status.Subnets[0].Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(subnet).ToNot(BeNil())
			osSubnet = subnet
		})

		By("And Then KCP IpRange Openstack Subnet range equals to IpRange cidr", func() {
			Expect(osSubnet.CIDR).To(Equal(kcpIpRange.Status.Cidr))
			Expect(osSubnet.CIDR).To(Equal(kcpIpRange.Status.Subnets[0].Range))
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
			subnet, err := sapMock.GetSubnetByName(infra.Ctx(), networkId, kcpIpRange.Status.Subnets[0].Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(subnet).To(BeNil())
		})
	})
})
