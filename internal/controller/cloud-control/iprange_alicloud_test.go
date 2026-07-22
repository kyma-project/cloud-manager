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

	It("Scenario: KCP Alicloud IpRange CIDR is automatically allocated outside VPC CIDR and vSwitch is created", func() {
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

		By("And Then KCP IpRange has allocated status.cidr disjoint from VPC CIDR 10.180.0.0/16", func() {
			Expect(iprange.Status.Cidr).NotTo(BeEmpty())
			Expect(iprange.Status.Cidr).To(HaveSuffix("/22"))
			Expect(iprange.Status.Cidr).NotTo(HavePrefix("10.180."))
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

		By("And Then secondary CIDR block is associated to VPC in mock", func() {
			attr, err := alicloudRegion.IpRangeClient().DescribeVpcAttribute(infra.Ctx(), "vpc-ac-01")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.SecondaryCidrBlocks).To(ContainElement(iprange.Status.Cidr))
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
			vsw, err := alicloudRegion.IpRangeClient().DescribeVSwitchesByName(infra.Ctx(), "vpc-ac-01", "cm-"+iprangeName+"-0")
			Expect(err).NotTo(HaveOccurred())
			Expect(vsw).To(BeEmpty())
		})

		By("And Then secondary CIDR block is disassociated from VPC in mock", func() {
			attr, err := alicloudRegion.IpRangeClient().DescribeVpcAttribute(infra.Ctx(), "vpc-ac-01")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.SecondaryCidrBlocks).To(BeEmpty())
		})
	})

	It("Scenario: KCP Alicloud IpRange with explicit spec.cidr creates vSwitch", func() {
		const (
			kymaName    = "ac-ipr-explicit-01"
			iprangeName = "ac-ipr-explicit-iprange-01"
			iprangeCidr = "10.181.0.0/22"
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

		By("And Then secondary CIDR block is associated to VPC in mock", func() {
			attr, err := alicloudRegion.IpRangeClient().DescribeVpcAttribute(infra.Ctx(), "vpc-ac-02")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.SecondaryCidrBlocks).To(ContainElement(iprangeCidr))
		})
	})

	It("Scenario: KCP Alicloud IpRange with multiple zones creates one vSwitch per zone", func() {
		const (
			kymaName    = "ac-ipr-multizone-01"
			iprangeName = "ac-ipr-multizone-iprange-01"
			iprangeCidr = "10.181.0.0/22"
			region      = "ap-southeast-1"
		)

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()
		alicloudRegion := alicloudAccount.Region(region)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope with two zones exists", func() {
			kcpscope.Ignore.AddName(kymaName)

			scope.Name = kymaName
			scope.Namespace = DefaultKcpNamespace
			scope.Spec = cloudcontrolv1beta1.ScopeSpec{
				KymaName:  kymaName,
				ShootName: kymaName,
				Region:    region,
				Provider:  cloudcontrolv1beta1.ProviderAlicloud,
				Scope: cloudcontrolv1beta1.ScopeInfo{
					Alicloud: &cloudcontrolv1beta1.AlicloudScope{
						AccountId:  alicloudAccount.Credentials().AccessKeyId,
						VpcNetwork: common.GardenerVpcName(DefaultKcpNamespace, kymaName),
						Network: cloudcontrolv1beta1.AlicloudNetwork{
							Nodes:    "10.180.0.0/16",
							Pods:     "172.16.0.0/13",
							Services: "172.24.0.0/13",
							VPC: cloudcontrolv1beta1.AlicloudVPC{
								CIDR: "10.180.0.0/16",
							},
							Zones: []cloudcontrolv1beta1.AlicloudZone{
								{Name: "ap-southeast-1a", Workers: "10.180.0.0/19"},
								{Name: "ap-southeast-1b", Workers: "10.180.32.0/19"},
							},
						},
					},
				},
			}

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})

		vpcName := common.GardenerVpcName(scope.Namespace, kymaName)
		alicloudVpcName := vpcName + "-vpc"

		By("And Given Alicloud VPC with two zones exists", func() {
			alicloudRegion.AddVpc("vpc-ac-03", alicloudVpcName, "10.180.0.0/16")
			alicloudRegion.AddZone("ap-southeast-1a")
			alicloudRegion.AddZone("ap-southeast-1b")
		})

		var kcpNetworkKyma *cloudcontrolv1beta1.Network

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(kymaName).
				WithName(common.KcpNetworkKymaCommonName(kymaName)).
				WithAlicloudRef(scope.Spec.Scope.Alicloud.AccountId, region, "vpc-ac-03", scope.Spec.Scope.Alicloud.VpcNetwork).
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

		By("And Then KCP IpRange has two subnets, one per zone", func() {
			Expect(iprange.Status.Subnets).To(HaveLen(2))
			Expect(iprange.Status.Subnets[0].Zone).To(Equal("ap-southeast-1a"))
			Expect(iprange.Status.Subnets[1].Zone).To(Equal("ap-southeast-1b"))
			Expect(iprange.Status.Subnets[0].Range).NotTo(Equal(iprange.Status.Subnets[1].Range))
		})

		By("And Then secondary CIDR block is associated to VPC", func() {
			attr, err := alicloudRegion.IpRangeClient().DescribeVpcAttribute(infra.Ctx(), "vpc-ac-03")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.SecondaryCidrBlocks).To(ContainElement(iprangeCidr))
		})

		By("And Then both vSwitches exist in mock", func() {
			for _, subnet := range iprange.Status.Subnets {
				vsw, err := alicloudRegion.IpRangeClient().DescribeVSwitch(infra.Ctx(), subnet.Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(vsw).NotTo(BeNil())
			}
		})

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

		By("And Then vSwitches are deleted from mock", func() {
			vswitches, err := alicloudRegion.IpRangeClient().DescribeVSwitchesByVpcId(infra.Ctx(), "vpc-ac-03")
			Expect(err).NotTo(HaveOccurred())
			Expect(vswitches).To(BeEmpty())
		})

		By("And Then secondary CIDR block is disassociated from VPC", func() {
			attr, err := alicloudRegion.IpRangeClient().DescribeVpcAttribute(infra.Ctx(), "vpc-ac-03")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.SecondaryCidrBlocks).To(BeEmpty())
		})

		By("// cleanup: delete KCP Kyma Network", func() {
			Expect(infra.KCP().Client().Delete(infra.Ctx(), kcpNetworkKyma)).To(Succeed())
		})

		By("// cleanup: delete Scope", func() {
			Expect(infra.KCP().Client().Delete(infra.Ctx(), scope)).To(Succeed())
		})
	})

	It("Scenario: KCP Alicloud IpRange with spec.cidr overlapping VPC primary CIDR errors", func() {
		const (
			kymaName    = "ac-ipr-overlap-01"
			iprangeName = "ac-ipr-overlap-iprange-01"
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
			alicloudRegion.AddVpc("vpc-ac-04", alicloudVpcName, "10.180.0.0/16")
			alicloudRegion.AddZone("ap-southeast-1a")
		})

		var kcpNetworkKyma *cloudcontrolv1beta1.Network

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(kymaName).
				WithName(common.KcpNetworkKymaCommonName(kymaName)).
				WithAlicloudRef(scope.Spec.Scope.Alicloud.AccountId, scope.Spec.Region, "vpc-ac-04", scope.Spec.Scope.Alicloud.VpcNetwork).
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

		By("When KCP IpRange is created with overlapping spec.cidr", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					WithName(iprangeName),
					WithKcpIpRangeRemoteRef(iprangeName),
					WithScope(kymaName),
					WithKcpIpRangeSpecCidr(iprangeCidr),
				).
				Should(Succeed())
		})

		By("Then KCP IpRange has Error condition with CidrOverlap reason", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					NewObjActions(),
					HavingConditionReasonTrue(cloudcontrolv1beta1.ConditionTypeError, cloudcontrolv1beta1.ReasonCidrOverlap),
				).
				Should(Succeed())
		})

		By("And Then no secondary CIDR block is associated to VPC", func() {
			attr, err := alicloudRegion.IpRangeClient().DescribeVpcAttribute(infra.Ctx(), "vpc-ac-04")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.SecondaryCidrBlocks).To(BeEmpty())
		})

		By("// cleanup: delete KCP IpRange", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange).
				Should(Succeed())
		})

		By("// cleanup: delete KCP Kyma Network", func() {
			Expect(infra.KCP().Client().Delete(infra.Ctx(), kcpNetworkKyma)).To(Succeed())
		})
	})
})
