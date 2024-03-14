package cloudcontrol

import (
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("KCP IpRange", func() {

	const (
		kymaName = "d87cfa6d-ff74-47e9-a3f6-c6efc637ce2a"
		vpcId    = "b1d68fc4-1bd4-4ad6-b81c-3d86de54f4f9"
	)

	scope := &cloudcontrolv1beta1.Scope{}

	It("IpRange AWS", func() {

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			scopePkg.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		By("And Given AWS VPC exists", func() {
			infra.AwsMock().AddVpc(
				vpcId,
				"10.180.0.0/16",
				awsutil.Ec2Tags("Name", scope.Spec.Scope.Aws.VpcNetwork),
				awsmock.VpcSubnetsFromScope(scope),
			)
		})

		iprangeName := "some-aws-ip-range"
		iprangeCidr := "10.181.0.0/16"
		iprange := &cloudcontrolv1beta1.IpRange{}

		By("When IpRange is created", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					WithName(iprangeName),
					WithKcpIpRangeRemoteRef("skr-namespace", "skr-aws-ip-range"),
					WithKcpIpRangeSpecScope(kymaName),
					WithKcpIpRangeSpecCidr(iprangeCidr),
				).
				Should(Succeed())
		})

		By("Then IpRange will have Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), iprange,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And has status.cidr equal to spec.cidr", func() {
			Expect(iprange.Status.Cidr).To(Equal(iprange.Spec.Cidr), "expected IpRange status.cidr to be equal to spec.cidr")
		})

		By("And has count(status.ranges) equal to Scope zones count", func() {
			Expect(iprange.Status.Ranges).To(HaveLen(3), "expected three IpRange status.ranges")
			Expect(iprange.Status.Ranges).To(ContainElement("10.181.0.0/18"), "expected IpRange status range to have 10.181.0.0/18")
			Expect(iprange.Status.Ranges).To(ContainElement("10.181.64.0/18"), "expected IpRange status range to have 10.181.64.0/18")
			Expect(iprange.Status.Ranges).To(ContainElement("10.181.128.0/18"), "expected IpRange status range to have 10.181.128.0/18")
		})

		By("And has status.vpcId equal to existing AWS VPC id", func() {
			Expect(iprange.Status.VpcId).To(Equal(vpcId))
		})

		By("And has status.subnets as Scope has zones", func() {
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
	})

})
