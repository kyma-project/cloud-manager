package cloudcontrol

import (
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
	"time"
)

var _ = Describe("Feature: KCP VpcPeering", func() {

	It("Scenario: KCP AWS VpcPeering is created", func() {
		const (
			kymaName        = "09bdb13e-8a51-4920-852d-b170433d1236"
			vpcId           = "vpc-c0c7d75db0832988d"
			vpcCidr         = "10.180.0.0/16"
			remoteVpcId     = "vpc-2c41e43fcd5340f8f"
			remoteVpcCidr   = "10.200.0.0/16"
			remoteAccountId = "444455556666"
			remoteRegion    = "eu-west1"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			scopePkg.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		By("And Given AWS VPC exists", func() {
			infra.AwsMock().AddVpc(
				"wrong1",
				"10.200.0.0/16",
				awsutil.Ec2Tags("Name", "wrong1"),
				nil,
			)
			infra.AwsMock().AddVpc(
				vpcId,
				vpcCidr,
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

		By("And Given AWS route table exists", func() {

			infra.AwsMock().AddRouteTable(
				ptr.To("rtb-c6606c725da27ff10"),
				ptr.To(vpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", scope.Spec.Scope.Aws.VpcNetwork), "1"),
				[]ec2Types.RouteTableAssociation{})

			infra.AwsMock().AddRouteTable(
				ptr.To("rtb-0c65354e2979d9c83"),
				ptr.To(vpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", scope.Spec.Scope.Aws.VpcNetwork), "1"),
				[]ec2Types.RouteTableAssociation{})

			infra.AwsMock().AddRouteTable(
				ptr.To("rtb-ae17300793a424240"),
				ptr.To(vpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", "wrong2"), "1"),
				[]ec2Types.RouteTableAssociation{})
		})

		By("And Given AWS remote VPC exists", func() {
			infra.AwsMock().AddVpc(
				remoteVpcId,
				remoteVpcCidr,
				awsutil.Ec2Tags("Name", "Remote Network Name", kymaName, kymaName),
				nil,
			)
			infra.AwsMock().AddVpc(
				"wrong3",
				"10.200.0.0/16",
				awsutil.Ec2Tags("Name", "wrong3"),
				nil,
			)
		})

		By("And Given AWS remote route table exists", func() {

			infra.AwsMock().AddRouteTable(
				ptr.To("rtb-69a1e8a929a9eb5ed"),
				ptr.To(remoteVpcId),
				awsutil.Ec2Tags(),
				[]ec2Types.RouteTableAssociation{
					{
						Main: ptr.To(true),
					},
				})

			infra.AwsMock().AddRouteTable(
				ptr.To("rtb-ae17300793a424247"),
				ptr.To(remoteVpcId),
				awsutil.Ec2Tags(),
				[]ec2Types.RouteTableAssociation{
					{
						SubnetId: ptr.To("10.200.1.0/24"),
					},
				})
		})

		obj := &cloudcontrolv1beta1.VpcPeering{}

		By("When KCP VpcPeering is created", func() {
			Eventually(CreateKcpVpcPeering).
				WithArguments(infra.Ctx(), infra.KCP().Client(), obj,
					WithName("1839c399-c52e-4b43-b156-4b51027508cd"),
					WithKcpVpcPeeringRemoteRef("skr-namespace", "skr-aws-ip-range"),
					WithKcpVpcPeeringSpecScope(kymaName),
					WithKcpVpcPeeringSpecAws(remoteVpcId, remoteAccountId, remoteRegion),
				).
				Should(Succeed())
		})

		By("Then KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), obj,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then KCP VpcPeering has status.vpcId equal to existing AWS VPC id", func() {
			Expect(obj.Status.VpcId).To(Equal(vpcId))
		})

		list, _ := infra.AwsMock().DescribeVpcPeeringConnections(infra.Ctx())

		var connection ec2Types.VpcPeeringConnection
		for _, p := range list {
			if obj.Status.Id == ptr.Deref(p.VpcPeeringConnectionId, "") {
				connection = p
			}
		}

		By("And Then found VpcPeeringConnection has AccepterVpcInfo.VpcId equals remote vpc id", func() {
			Expect(*connection.AccepterVpcInfo.VpcId).To(Equal(remoteVpcId))
		})

		By("And Then KCP VpcPeering has status.Id equal to existing AWS Connection id", func() {
			Expect(obj.Status.Id).To(Equal(ptr.Deref(connection.VpcPeeringConnectionId, "xxx")))
		})

		tables, _ := infra.AwsMock().DescribeRouteTables(infra.Ctx(), vpcId)

		By("And Then all route table has one new route with destination CIDR matching remote VPC CIDR", func() {
			Expect(routeCount(tables, *connection.VpcPeeringConnectionId, remoteVpcCidr)).To(Equal(3))
		})

		tables, _ = infra.AwsMock().DescribeRouteTables(infra.Ctx(), remoteVpcId)

		By("And Then all remote route table has one new route with destination CIDR matching VPC CIDR", func() {
			Expect(routeCount(tables, *connection.VpcPeeringConnectionId, vpcCidr)).To(Equal(2))
		})

		// DELETE

		By("When KCP VpcPeering is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), obj).
				Should(Succeed(), "failed deleting VpcPeering")
		})

		By("Then VpcPeering does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), obj).
				Should(Succeed(), "expected VpcPeering not to exist (be deleted), but it still exists")
		})

		By("And Then VpcPeeringConnection is deleted", func() {
			list, _ = infra.AwsMock().DescribeVpcPeeringConnections(infra.Ctx())

			found := pie.Any(list, func(x ec2Types.VpcPeeringConnection) bool {
				return ptr.Deref(x.VpcPeeringConnectionId, "xxx") == obj.Status.Id
			})

			Expect(found).To(Equal(false))
		})

		By("And Then all route table has no routes with destination CIDR matching remote VPC CIDR", func() {
			Expect(routeCount(tables, *connection.VpcPeeringConnectionId, remoteVpcCidr)).To(Equal(0))
		})
	})
})

func routeCount(tables []ec2Types.RouteTable, vpcPeeringConnectionId, destinationCidrBlock string) int {
	cnt := 0
	for _, t := range tables {
		for _, r := range t.Routes {
			if *r.VpcPeeringConnectionId == vpcPeeringConnectionId &&
				*r.DestinationCidrBlock == destinationCidrBlock {
				cnt++
			}
		}
	}
	return cnt
}
