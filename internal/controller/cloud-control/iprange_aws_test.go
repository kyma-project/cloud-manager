package cloudcontrol

import (
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("Feature: KCP IpRange for AWS", func() {

	It("Scenario: KCP AWS IpRange is created", func() {

		const (
			kymaName = "d87cfa6d-ff74-47e9-a3f6-c6efc637ce2a"
			vpcId    = "b1d68fc4-1bd4-4ad6-b81c-3d86de54f4f9"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			scopePkg.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		var theVpc *ec2Types.Vpc
		By("And Given AWS VPC exists", func() {
			infra.AwsMock().AddVpc(
				"wrong1",
				"10.200.0.0/16",
				awsutil.Ec2Tags("Name", "wrong1"),
				nil,
			)
			theVpc = infra.AwsMock().AddVpc(
				vpcId,
				"10.250.0.0/22",
				awsutil.Ec2Tags("Name", scope.Spec.Scope.Aws.VpcNetwork),
				awsmock.VpcSubnetsFromScope(scope),
			)
			infra.AwsMock().AddVpc(
				"wrong2",
				"10.200.0.0/16",
				awsutil.Ec2Tags("Name", "wrong2"),
				nil,
			)
		})

		var kcpNetworkKyma *cloudcontrolv1beta1.Network

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(kymaName).
				WithName(common.KcpNetworkKymaCommonName(kymaName)).
				WithAwsRef(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region, vpcId, scope.Spec.Scope.Aws.VpcNetwork).
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

		iprangeName := "b76ff161-c288-44fa-a295-8df2076af6a5"
		iprangeCidr := "10.181.0.0/16"
		iprange := &cloudcontrolv1beta1.IpRange{}

		By("When KCP IpRange is created", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					WithName(iprangeName),
					WithKcpIpRangeRemoteRef("skr-aws-ip-range"),
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

		By("And Then KCP IpRange has status.cidr equal to spec.cidr", func() {
			Expect(iprange.Status.Cidr).To(Equal(iprange.Spec.Cidr), "expected IpRange status.cidr to be equal to spec.cidr")
		})

		By("And Then KCP IpRange has count(status.ranges) equal to Scope zones count", func() {
			Expect(iprange.Status.Ranges).To(HaveLen(3), "expected three IpRange status.ranges")
			Expect(iprange.Status.Ranges).To(ContainElement("10.181.0.0/18"), "expected IpRange status range to have 10.181.0.0/18")
			Expect(iprange.Status.Ranges).To(ContainElement("10.181.64.0/18"), "expected IpRange status range to have 10.181.64.0/18")
			Expect(iprange.Status.Ranges).To(ContainElement("10.181.128.0/18"), "expected IpRange status range to have 10.181.128.0/18")
		})

		By("And Then KCP IpRange has status.vpcId equal to existing AWS VPC id", func() {
			Expect(iprange.Status.VpcId).To(Equal(vpcId))
		})

		By("And Then KCP IpRange has status.subnets as Scope has zones", func() {
			Expect(iprange.Status.Subnets).To(HaveLen(3))

			Expect(iprange.Status.Subnets).To(HaveLen(3))
			expectedZones := map[string]struct{}{
				"eu-west-1a": {},
				"eu-west-1b": {},
				"eu-west-1c": {},
			}
			for i, subnet := range iprange.Status.Subnets {
				Expect(subnet.Id).NotTo(BeEmpty(), fmt.Sprintf("expected IpRange.status.subnets[%d].id not to be empty", i))
				Expect(iprange.Status.Ranges).To(ContainElement(subnet.Range), fmt.Sprintf("expected IpRange.status.subnets[%d].range %s to be listed in IpRange.status.ranges", i, subnet.Range))
				Expect(expectedZones).To(HaveKey(subnet.Zone), fmt.Sprintf("expected IpRange.status.subnets[%d].zone %s to be one of %v", i, subnet.Zone, expectedZones))
				delete(expectedZones, subnet.Zone)
			}
		})

		By("And Then KCP IpRange AWS Subnets are created", func() {
			subnets, err := infra.AwsMock().DescribeSubnets(infra.Ctx(), vpcId)
			Expect(err).NotTo(HaveOccurred())
			for _, iprangeSubnet := range iprange.Status.Subnets {
				found := false
				for _, awsSubnet := range subnets {
					if iprangeSubnet.Id == ptr.Deref(awsSubnet.SubnetId, "") {
						Expect(ptr.Deref(awsSubnet.AvailabilityZone, "")).To(Equal(iprangeSubnet.Zone))
						Expect(ptr.Deref(awsSubnet.CidrBlock, "")).To(Equal(iprangeSubnet.Range))
						found = true
						break
					}
				}
				Expect(found).To(BeTrue(), fmt.Sprintf("Expected AWS subnet %s %s %s to exist, but none found", iprangeSubnet.Id, iprangeSubnet.Zone, iprangeSubnet.Range))
			}
		})

		By("And Then KCP IpRange AWS VPC Cidr block is created", func() {
			found := false
			for _, cidrBlock := range theVpc.CidrBlockAssociationSet {
				if ptr.Deref(cidrBlock.CidrBlock, "") == iprange.Spec.Cidr {
					found = true
				}
			}
			Expect(found).To(BeTrue(), "expected KCP IpRange VPC Cidr block to be created, but none found")
		})
	})

	It("Scenario: KCP AWS IpRange is deleted", func() {
		const (
			kymaName    = "b46d0996-5c9b-42a4-a4f7-65528de92514"
			vpcId       = "e434aac3-2557-4fd8-8ff4-35f1fb873b6e"
			vpcCidr     = "10.180.0.0/16"
			iprangeName = "dfb7578e-8eaf-4397-bd36-2eb24727b8ca"
			iprangeCidr = "10.181.0.0/16"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			scopePkg.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		var theVpc *ec2Types.Vpc
		By("And Given AWS VPC exists", func() {
			infra.AwsMock().AddVpc(
				"wrong1",
				"10.200.0.0/16",
				awsutil.Ec2Tags("Name", "wrong1"),
				nil,
			)
			theVpc = infra.AwsMock().AddVpc(
				vpcId,
				vpcCidr,
				awsutil.Ec2Tags("Name", scope.Spec.Scope.Aws.VpcNetwork),
				awsmock.VpcSubnetsFromScope(scope),
			)
			Expect(theVpc).NotTo(BeNil(), "expected non nil aws vpc to be created")
			infra.AwsMock().AddVpc(
				"wrong2",
				"10.200.0.0/16",
				awsutil.Ec2Tags("Name", "wrong2"),
				nil,
			)
		})

		var kcpNetworkKyma *cloudcontrolv1beta1.Network

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(kymaName).
				WithName(common.KcpNetworkKymaCommonName(kymaName)).
				WithAwsRef(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region, vpcId, scope.Spec.Scope.Aws.VpcNetwork).
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

		By("And Given KCP IpRange is created", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					WithName(iprangeName),
					WithKcpIpRangeRemoteRef("skr-aws-ip-range"),
					WithScope(kymaName),
					WithKcpIpRangeSpecCidr(iprangeCidr),
				).
				Should(Succeed())
		})

		By("And Given KCP IpRange has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		var theAwsSubnets []ec2Types.Subnet
		By("And Given KCP IpRange AWS Subnets are created", func() {
			awsSubnets, err := infra.AwsMock().DescribeSubnets(infra.Ctx(), vpcId)
			Expect(err).NotTo(HaveOccurred())
			for _, iprangeSubnet := range iprange.Status.Subnets {
				for _, awsSubnet := range awsSubnets {
					if iprangeSubnet.Id == ptr.Deref(awsSubnet.SubnetId, "") {
						theAwsSubnets = append(theAwsSubnets, awsSubnet)
						break
					}
				}
			}
		})

		By("When KCP IpRange is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange).
				Should(Succeed(), "failed deleting KCP IpRange")
		})

		By("Then KCP IpRange does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange).
				Should(Succeed(), "expected KCP IpRange to be deleted, but it exists")
		})

		By("And Then KCP IpRange VPC Subnets do not exist", func() {
			subnets, err := infra.AwsMock().DescribeSubnets(infra.Ctx(), vpcId)
			Expect(err).NotTo(HaveOccurred())

			for _, deletedSubnet := range theAwsSubnets {
				for _, existingSubnet := range subnets {
					Expect(ptr.Deref(deletedSubnet.SubnetId, "")).
						NotTo(
							Equal(ptr.Deref(existingSubnet.SubnetId, "")),
							fmt.Sprintf(
								"expected subnet %s/%s/%s to be deleted, but it still exists",
								ptr.Deref(deletedSubnet.SubnetId, ""),
								ptr.Deref(deletedSubnet.AvailabilityZone, ""),
								ptr.Deref(deletedSubnet.CidrBlock, ""),
							),
						)
				}
			}
		})

		By("And Then KCP IpRange Cidr block does not exist", func() {
			for _, cidrBlock := range theVpc.CidrBlockAssociationSet {
				Expect(ptr.Deref(cidrBlock.CidrBlock, "")).
					NotTo(Equal(iprangeCidr), "expected VPC Cidr block not to exist, but it still exists")
			}
		})

	})

	It("Scenario: KCP AWS IpRange CIDR is automatically allocated", func() {
		const (
			kymaName    = "446dda67-dd5b-4af4-9860-70557a5a1160"
			vpcId       = "f79a4690-a580-43c9-bb02-b71f4e03a346"
			iprangeName = "d968c855-7de8-4e33-84b2-15510d9865a7"
		)

		scope := &cloudcontrolv1beta1.Scope{}
		iprange := &cloudcontrolv1beta1.IpRange{}
		var theVpc *ec2Types.Vpc

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			scopePkg.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		By("And Given AWS VPC exists", func() {
			theVpc = infra.AwsMock().AddVpc(
				vpcId,
				"10.250.0.0/22",
				awsutil.Ec2Tags("Name", scope.Spec.Scope.Aws.VpcNetwork),
				awsmock.VpcSubnetsFromScope(scope),
			)
		})

		var kcpNetworkKyma *cloudcontrolv1beta1.Network

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(kymaName).
				WithName(common.KcpNetworkKymaCommonName(kymaName)).
				WithAwsRef(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region, vpcId, scope.Spec.Scope.Aws.VpcNetwork).
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

		By("When KCP IpRange is created", func() {
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

		By("And Then KCP IpRange has empty spec.cidr", func() {
			Expect(iprange.Spec.Cidr).To(Equal(""), "expected IpRange spec.cidr to be empty")
		})

		By("And Then KCP IpRange has allocated status.cidr", func() {
			Expect(iprange.Status.Cidr).To(Equal("10.250.4.0/22"), "expected IpRange status.cidr to be allocated")
		})

		By("And Then KCP IpRange has count(status.ranges) equal to Scope zones count", func() {
			Expect(iprange.Status.Ranges).To(HaveLen(3), "expected three IpRange status.ranges")
			Expect(iprange.Status.Ranges).To(ContainElement("10.250.4.0/24"), "expected IpRange status range to have 10.181.0.0/18")
			Expect(iprange.Status.Ranges).To(ContainElement("10.250.5.0/24"), "expected IpRange status range to have 10.181.64.0/18")
			Expect(iprange.Status.Ranges).To(ContainElement("10.250.6.0/24"), "expected IpRange status range to have 10.181.128.0/18")
		})

		By("And Then KCP IpRange has status.vpcId equal to existing AWS VPC id", func() {
			Expect(iprange.Status.VpcId).To(Equal(vpcId))
		})

		By("And Then KCP IpRange has status.subnets as Scope has zones", func() {
			Expect(iprange.Status.Subnets).To(HaveLen(3))

			Expect(iprange.Status.Subnets).To(HaveLen(3))
			expectedZones := map[string]struct{}{
				"eu-west-1a": {},
				"eu-west-1b": {},
				"eu-west-1c": {},
			}
			for i, subnet := range iprange.Status.Subnets {
				Expect(subnet.Id).NotTo(BeEmpty(), fmt.Sprintf("expected IpRange.status.subnets[%d].id not to be empty", i))
				Expect(iprange.Status.Ranges).To(ContainElement(subnet.Range), fmt.Sprintf("expected IpRange.status.subnets[%d].range %s to be listed in IpRange.status.ranges", i, subnet.Range))
				Expect(expectedZones).To(HaveKey(subnet.Zone), fmt.Sprintf("expected IpRange.status.subnets[%d].zone %s to be one of %v", i, subnet.Zone, expectedZones))
				delete(expectedZones, subnet.Zone)
			}
		})

		By("And Then KCP IpRange AWS Subnets are created", func() {
			subnets, err := infra.AwsMock().DescribeSubnets(infra.Ctx(), vpcId)
			Expect(err).NotTo(HaveOccurred())
			for _, iprangeSubnet := range iprange.Status.Subnets {
				found := false
				for _, awsSubnet := range subnets {
					if iprangeSubnet.Id == ptr.Deref(awsSubnet.SubnetId, "") {
						Expect(ptr.Deref(awsSubnet.AvailabilityZone, "")).To(Equal(iprangeSubnet.Zone))
						Expect(ptr.Deref(awsSubnet.CidrBlock, "")).To(Equal(iprangeSubnet.Range))
						found = true
						break
					}
				}
				Expect(found).To(BeTrue(), fmt.Sprintf("Expected AWS subnet %s %s %s to exist, but none found", iprangeSubnet.Id, iprangeSubnet.Zone, iprangeSubnet.Range))
			}
		})

		By("And Then KCP IpRange AWS VPC Cidr block is created", func() {
			found := false
			var allBlocks []string
			for _, cidrBlock := range theVpc.CidrBlockAssociationSet {
				cidr := ptr.Deref(cidrBlock.CidrBlock, "")
				allBlocks = append(allBlocks, cidr)
				if cidr == iprange.Status.Cidr {
					found = true
				}
			}
			Expect(found).To(BeTrue(), fmt.Sprintf("expected KCP IpRange VPC Cidr block to be created, but none found: %#v", allBlocks))
		})
	})

})
