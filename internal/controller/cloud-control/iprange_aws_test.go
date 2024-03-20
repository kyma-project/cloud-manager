package cloudcontrol

import (
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/pointer"
)

var _ = Describe("Feature: KCP IpRange", func() {

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
			theVpc = infra.AwsMock().AddVpc(
				vpcId,
				"10.180.0.0/16",
				awsutil.Ec2Tags("Name", scope.Spec.Scope.Aws.VpcNetwork),
				awsmock.VpcSubnetsFromScope(scope),
			)
		})

		iprangeName := "b76ff161-c288-44fa-a295-8df2076af6a5"
		iprangeCidr := "10.181.0.0/16"
		iprange := &cloudcontrolv1beta1.IpRange{}

		By("When KCP IpRange is created", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					WithName(iprangeName),
					WithKcpIpRangeRemoteRef("skr-namespace", "skr-aws-ip-range"),
					WithKcpIpRangeSpecScope(kymaName),
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
					if iprangeSubnet.Id == pointer.StringDeref(awsSubnet.SubnetId, "") {
						Expect(pointer.StringDeref(awsSubnet.AvailabilityZone, "")).To(Equal(iprangeSubnet.Zone))
						Expect(pointer.StringDeref(awsSubnet.CidrBlock, "")).To(Equal(iprangeSubnet.Range))
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
				if pointer.StringDeref(cidrBlock.CidrBlock, "") == iprange.Spec.Cidr {
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
			theVpc = infra.AwsMock().AddVpc(
				vpcId,
				vpcCidr,
				awsutil.Ec2Tags("Name", scope.Spec.Scope.Aws.VpcNetwork),
				awsmock.VpcSubnetsFromScope(scope),
			)
			Expect(theVpc).NotTo(BeNil(), "expected non nil aws vpc to be created")
		})

		iprange := &cloudcontrolv1beta1.IpRange{}

		By("And Given KCP IpRange is created", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					WithName(iprangeName),
					WithKcpIpRangeRemoteRef("skr-namespace", "skr-aws-ip-range"),
					WithKcpIpRangeSpecScope(kymaName),
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
					if iprangeSubnet.Id == pointer.StringDeref(awsSubnet.SubnetId, "") {
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
					Expect(pointer.StringDeref(deletedSubnet.SubnetId, "")).
						NotTo(
							Equal(pointer.StringDeref(existingSubnet.SubnetId, "")),
							fmt.Sprintf(
								"expected subnet %s/%s/%s to be deleted, but it still exists",
								pointer.StringDeref(deletedSubnet.SubnetId, ""),
								pointer.StringDeref(deletedSubnet.AvailabilityZone, ""),
								pointer.StringDeref(deletedSubnet.CidrBlock, ""),
							),
						)
				}
			}
		})

		By("And Then KCP IpRange Cidr block does not exist", func() {
			for _, cidrBlock := range theVpc.CidrBlockAssociationSet {
				Expect(pointer.StringDeref(cidrBlock.CidrBlock, "")).
					NotTo(Equal(iprangeCidr),
						fmt.Sprintf("expected VPC Cidr block not to exist, but it still exists"),
					)
			}
		})

	})

})
