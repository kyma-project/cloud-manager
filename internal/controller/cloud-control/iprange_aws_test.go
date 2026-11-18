package cloudcontrol

import (
	"fmt"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
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

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(kymaName)).
				Should(Succeed())
		})

		awsMock := awsAccount.Region(scope.Spec.Region)

		var theVpc *ec2types.Vpc
		By("And Given AWS VPC exists", func() {
			awsMock.AddVpc(
				"wrong1",
				"10.200.0.0/16",
				awsutil.Ec2Tags("Name", "wrong1"),
				nil,
			)
			theVpc = awsMock.AddVpc(
				vpcId,
				"10.250.0.0/22",
				awsutil.Ec2Tags("Name", scope.Spec.Scope.Aws.VpcNetwork),
				awsmock.VpcSubnetsFromScope(scope),
			)
			awsMock.AddVpc(
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
			subnets, err := awsMock.DescribeSubnets(infra.Ctx(), vpcId)
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

	It("Scenario: second KCP AWS IpRange is created in the same VPC", func() {

		const (
			kymaName     = "0d1c5e2a-8f2b-4c7e-9b3a-2d4f6a8c1e5b"
			vpcId        = "7c3e9a1d-4b6f-48a2-9c5e-1f3d7b2a6c8e"
			firstIpRange = "3a5b7c9d-1e2f-4a6b-8c0d-5e7f9a1b3c5d"
			firstCidr    = "10.181.0.0/16"
			secondRange  = "6d8f0a2b-4c6e-48a0-b2d4-7f9b1d3f5a7c"
			secondCidr   = "10.182.0.0/16"
		)

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(kymaName)).
				Should(Succeed())
		})

		awsMock := awsAccount.Region(scope.Spec.Region)

		By("And Given AWS VPC exists", func() {
			awsMock.AddVpc(
				vpcId,
				"10.250.0.0/22",
				awsutil.Ec2Tags("Name", scope.Spec.Scope.Aws.VpcNetwork),
				awsmock.VpcSubnetsFromScope(scope),
			)
		})

		By("And Given KCP Kyma Network exists in Ready state", func() {
			kcpNetworkKyma := cloudcontrolv1beta1.NewNetworkBuilder().
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

		first := &cloudcontrolv1beta1.IpRange{}

		By("And Given first KCP IpRange is Ready", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), first,
					WithName(firstIpRange),
					WithKcpIpRangeRemoteRef("skr-aws-ip-range-first"),
					WithScope(kymaName),
					WithKcpIpRangeSpecCidr(firstCidr),
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), first,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())

			Expect(first.Status.Subnets).To(HaveLen(3), "expected first IpRange to have three subnets")
		})

		var firstAwsSubnets []ec2types.Subnet
		By("And Given first KCP IpRange AWS Subnets exist", func() {
			awsSubnets, err := awsMock.DescribeSubnets(infra.Ctx(), vpcId)
			Expect(err).NotTo(HaveOccurred())
			for _, iprangeSubnet := range first.Status.Subnets {
				for _, awsSubnet := range awsSubnets {
					if iprangeSubnet.Id == ptr.Deref(awsSubnet.SubnetId, "") {
						firstAwsSubnets = append(firstAwsSubnets, awsSubnet)
						break
					}
				}
			}
			Expect(firstAwsSubnets).To(HaveLen(3), "expected three AWS subnets for first IpRange")
		})

		second := &cloudcontrolv1beta1.IpRange{}

		By("When second KCP IpRange is created in the same VPC", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), second,
					WithName(secondRange),
					WithKcpIpRangeRemoteRef("skr-aws-ip-range-second"),
					WithScope(kymaName),
					WithKcpIpRangeSpecCidr(secondCidr),
				).
				Should(Succeed())
		})

		By("Then second KCP IpRange has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), second,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then second KCP IpRange has its own three subnets in distinct zones", func() {
			Expect(second.Status.Subnets).To(HaveLen(3), "expected second IpRange to have exactly three subnets")

			seenZones := map[string]struct{}{}
			for i, subnet := range second.Status.Subnets {
				Expect(subnet.Id).NotTo(BeEmpty(), fmt.Sprintf("expected second IpRange.status.subnets[%d].id not to be empty", i))
				// Must belong to the second IpRange's own ranges, not the first's.
				Expect(second.Status.Ranges).To(
					ContainElement(subnet.Range),
					fmt.Sprintf("expected second IpRange.status.subnets[%d].range %s to be one of its own ranges %v", i, subnet.Range, second.Status.Ranges),
				)
				Expect(seenZones).NotTo(HaveKey(subnet.Zone), fmt.Sprintf("expected second IpRange.status.subnets[%d].zone %s to be distinct", i, subnet.Zone))
				seenZones[subnet.Zone] = struct{}{}
			}
			Expect(seenZones).To(HaveLen(3), "expected second IpRange subnets to span three distinct zones")
		})

		By("When second KCP IpRange is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), second).
				Should(Succeed(), "failed deleting second KCP IpRange")

			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), second).
				Should(Succeed(), "expected second KCP IpRange to be deleted, but it exists")
		})

		By("Then first KCP IpRange AWS Subnets still exist", func() {
			awsSubnets, err := awsMock.DescribeSubnets(infra.Ctx(), vpcId)
			Expect(err).NotTo(HaveOccurred())

			for _, firstSubnet := range firstAwsSubnets {
				found := false
				for _, awsSubnet := range awsSubnets {
					if ptr.Deref(firstSubnet.SubnetId, "") == ptr.Deref(awsSubnet.SubnetId, "") {
						found = true
						break
					}
				}
				Expect(found).To(
					BeTrue(),
					fmt.Sprintf(
						"expected first IpRange subnet %s/%s/%s to survive deletion of the second IpRange",
						ptr.Deref(firstSubnet.SubnetId, ""),
						ptr.Deref(firstSubnet.AvailabilityZone, ""),
						ptr.Deref(firstSubnet.CidrBlock, ""),
					),
				)
			}
		})

		By("Cleanup: delete first KCP IpRange", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), first).
				Should(Succeed(), "failed deleting first KCP IpRange")

			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), first).
				Should(Succeed(), "expected first KCP IpRange to be deleted, but it exists")
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

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(kymaName)).
				Should(Succeed())
		})

		awsMock := awsAccount.Region(scope.Spec.Region)

		var theVpc *ec2types.Vpc
		By("And Given AWS VPC exists", func() {
			awsMock.AddVpc(
				"wrong1",
				"10.200.0.0/16",
				awsutil.Ec2Tags("Name", "wrong1"),
				nil,
			)
			theVpc = awsMock.AddVpc(
				vpcId,
				vpcCidr,
				awsutil.Ec2Tags("Name", scope.Spec.Scope.Aws.VpcNetwork),
				awsmock.VpcSubnetsFromScope(scope),
			)
			Expect(theVpc).NotTo(BeNil(), "expected non nil aws vpc to be created")
			awsMock.AddVpc(
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

		var theAwsSubnets []ec2types.Subnet
		By("And Given KCP IpRange AWS Subnets are created", func() {
			awsSubnets, err := awsMock.DescribeSubnets(infra.Ctx(), vpcId)
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
			subnets, err := awsMock.DescribeSubnets(infra.Ctx(), vpcId)
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

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()

		scope := &cloudcontrolv1beta1.Scope{}
		iprange := &cloudcontrolv1beta1.IpRange{}
		var theVpc *ec2types.Vpc

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(kymaName)).
				Should(Succeed())
		})

		awsMock := awsAccount.Region(scope.Spec.Region)

		By("And Given AWS VPC exists", func() {
			theVpc = awsMock.AddVpc(
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
			subnets, err := awsMock.DescribeSubnets(infra.Ctx(), vpcId)
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

	It("Scenario: KCP AWS IpRange is deleted when AWS VPC does not exist", func() {
		const (
			kymaName    = "a3e4c7f0-9b1d-4e2a-8c6f-1d2e3f4a5b6c"
			vpcId       = "c1d2e3f4-a5b6-47c8-9d0e-1f2a3b4c5d6e"
			vpcCidr     = "10.182.0.0/16"
			iprangeName = "f6e5d4c3-b2a1-4098-8f7e-6d5c4b3a2b1a"
			iprangeCidr = "10.183.0.0/16"
		)

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(kymaName)).
				Should(Succeed())
		})

		awsMock := awsAccount.Region(scope.Spec.Region)

		By("And Given AWS VPC exists", func() {
			theVpc := awsMock.AddVpc(
				vpcId,
				vpcCidr,
				awsutil.Ec2Tags("Name", scope.Spec.Scope.Aws.VpcNetwork),
				awsmock.VpcSubnetsFromScope(scope),
			)
			Expect(theVpc).NotTo(BeNil(), "expected non nil aws vpc to be created")
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

		By("And Given KCP IpRange has status.vpcId populated", func() {
			Expect(iprange.Status.VpcId).To(Equal(vpcId))
		})

		By("When the AWS VPC no longer exists", func() {
			// Removing the VPC makes DescribeVpc return (nil, nil) - the production
			// path where a previously provisioned IpRange's VPC/shoot was deleted.
			err := awsMock.DeleteVpc(infra.Ctx(), vpcId)
			Expect(err).NotTo(HaveOccurred())
		})

		By("And When KCP IpRange is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange).
				Should(Succeed(), "failed deleting KCP IpRange")
		})

		By("Then KCP IpRange does not exist", func() {
			// The reconciler must remove the finalizer and let the object be garbage
			// collected rather than deadlooping on the missing VPC.
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange).
				Should(Succeed(), "expected KCP IpRange to be deleted, but it exists")
		})
	})

})
