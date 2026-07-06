package cloudcontrol

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP IpRange for Alicloud", func() {

	It("Scenario: KCP Alicloud IpRange CIDR is automatically allocated within VPC and vSwitch is created", func() {
		const (
			kymaName    = "ac-ipr-alloc-01"
			iprangeName = "ac-ipr-alloc-iprange-01"
			region      = "ap-southeast-1"
		)

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()
		alicloudRegion := alicloudAccount.Region(region)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(kymaName)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, scope, alicloudAccount.Credentials().AccessKeyId, WithName(kymaName)).
				Should(Succeed())
		})

		// Seed the mock: VPC named "<vpcNetwork>-vpc" as Gardener would create it
		vpcName := common.GardenerVpcName(scope.Namespace, kymaName)
		alicloudVpcName := vpcName + "-vpc"

		By("And Given Alicloud VPC exists in mock (Gardener naming convention)", func() {
			alicloudRegion.AddVpc("vpc-ac-01", alicloudVpcName, "10.180.0.0/16")
			alicloudRegion.AddVSwitch("vpc-ac-01", "vsw-workers-01", "workers", "ap-southeast-1a", "10.180.0.0/18")
			alicloudRegion.AddZone("ap-southeast-1a")
		})

		var kcpNetworkKyma *cloudcontrolv1beta1.Network

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(kymaName).
				WithName(common.KcpNetworkKymaCommonName(kymaName)).
				WithAlicloudRef(scope.Spec.Scope.Alicloud.AccountId, scope.Spec.Region, "vpc-ac-01", scope.Spec.Scope.Alicloud.VpcNetwork).
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

		iprange := &cloudcontrolv1beta1.IpRange{}

		By("When KCP IpRange is created without spec.cidr", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					WithName(iprangeName),
					WithKcpIpRangeRemoteRef(iprangeName),
					WithScope(kymaName),
				).
				Should(Succeed())
		})

		By("Then KCP IpRange has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then KCP IpRange has allocated status.cidr within VPC CIDR 10.180.0.0/16", func() {
			Expect(iprange.Status.Cidr).NotTo(BeEmpty())
			// Must be a /22 within 10.180.0.0/16 but not overlap workers zone 10.180.0.0/18
			Expect(iprange.Status.Cidr).To(HavePrefix("10.180."))
			Expect(iprange.Status.Cidr).To(HaveSuffix("/22"))
			// Must not be in the workers range 10.180.0.0/18 (10.180.0.0 - 10.180.63.255)
			Expect(iprange.Status.Cidr).NotTo(Equal("10.180.0.0/22"))
		})

		By("And Then KCP IpRange has status.vpcId", func() {
			Expect(iprange.Status.VpcId).To(Equal("vpc-ac-01"))
		})

		By("And Then KCP IpRange has status.subnets with one vSwitch", func() {
			Expect(iprange.Status.Subnets).To(HaveLen(1))
			Expect(iprange.Status.Subnets[0].Id).NotTo(BeEmpty())
			Expect(iprange.Status.Subnets[0].Zone).To(Equal("ap-southeast-1a"))
			Expect(iprange.Status.Subnets[0].Range).To(Equal(iprange.Status.Cidr))
		})

		By("And Then vSwitch exists in mock", func() {
			vsw, err := alicloudRegion.IpRangeClient().DescribeVSwitch(infra.Ctx(), iprange.Status.Subnets[0].Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(vsw).NotTo(BeNil())
			Expect(vsw.CidrBlock).To(Equal(iprange.Status.Cidr))
			Expect(vsw.ZoneId).To(Equal("ap-southeast-1a"))
		})

		// DELETE ======================================================

		By("When KCP IpRange is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange).
				Should(Succeed())
		})

		By("Then KCP IpRange does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange).
				Should(Succeed())
		})

		By("And Then vSwitch is deleted from mock", func() {
			vsw, err := alicloudRegion.IpRangeClient().DescribeVSwitchesByName(infra.Ctx(), "vpc-ac-01", "cm-"+iprangeName)
			Expect(err).NotTo(HaveOccurred())
			Expect(vsw).To(BeEmpty())
		})
	})

	It("Scenario: KCP Alicloud IpRange with explicit spec.cidr creates vSwitch", func() {
		const (
			kymaName    = "ac-ipr-explicit-01"
			iprangeName = "ac-ipr-explicit-iprange-01"
			iprangeCidr = "10.180.64.0/22"
			region      = "ap-southeast-1"
		)

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()
		alicloudRegion := alicloudAccount.Region(region)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(kymaName)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, scope, alicloudAccount.Credentials().AccessKeyId, WithName(kymaName)).
				Should(Succeed())
		})

		vpcName := common.GardenerVpcName(scope.Namespace, kymaName)
		alicloudVpcName := vpcName + "-vpc"

		By("And Given Alicloud VPC exists", func() {
			alicloudRegion.AddVpc("vpc-ac-02", alicloudVpcName, "10.180.0.0/16")
			alicloudRegion.AddZone("ap-southeast-1a")
		})

		var kcpNetworkKyma *cloudcontrolv1beta1.Network

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(kymaName).
				WithName(common.KcpNetworkKymaCommonName(kymaName)).
				WithAlicloudRef(scope.Spec.Scope.Alicloud.AccountId, scope.Spec.Region, "vpc-ac-02", scope.Spec.Scope.Alicloud.VpcNetwork).
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

		iprange := &cloudcontrolv1beta1.IpRange{}

		By("When KCP IpRange is created with explicit spec.cidr", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					WithName(iprangeName),
					WithKcpIpRangeRemoteRef(iprangeName),
					WithScope(kymaName),
					WithKcpIpRangeSpecCidr(iprangeCidr),
				).
				Should(Succeed())
		})

		By("Then KCP IpRange has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then KCP IpRange status.cidr equals spec.cidr", func() {
			Expect(iprange.Status.Cidr).To(Equal(iprangeCidr))
		})

		By("And Then vSwitch is created with correct CIDR", func() {
			Expect(iprange.Status.Subnets).To(HaveLen(1))
			vsw, err := alicloudRegion.IpRangeClient().DescribeVSwitch(infra.Ctx(), iprange.Status.Subnets[0].Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(vsw).NotTo(BeNil())
			Expect(vsw.CidrBlock).To(Equal(iprangeCidr))
		})
	})
})
