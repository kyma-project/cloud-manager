package cloudcontrol

import (
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	kcpnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/network"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"time"
)

var _ = Describe("Feature: KCP VpcPeering", func() {

	It("Scenario: KCP AWS VpcPeering is created", func() {
		const (
			kymaName        = "09bdb13e-8a51-4920-852d-b170433d1236"
			kcpPeeringName  = "1839c399-c52e-4b43-b156-4b51027508cd"
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

		vpcName := scope.Spec.Scope.Aws.VpcNetwork
		remoteVpcName := "Remote Network Name"

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
				awsutil.Ec2Tags("Name", vpcName),
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
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", vpcName), "1"),
				[]ec2Types.RouteTableAssociation{})

			infra.AwsMock().AddRouteTable(
				ptr.To("rtb-0c65354e2979d9c83"),
				ptr.To(vpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", vpcName), "1"),
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
				awsutil.Ec2Tags("Name", remoteVpcName, kymaName, kymaName),
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

		localKcpNetworkName := common.KymaNetworkCommonName(scope.Name)
		remoteKcpNetworkName := scope.Name + "--remote"

		var kcpPeering *cloudcontrolv1beta1.VpcPeering

		By("When KCP VpcPeering is created", func() {
			kcpPeering = (&cloudcontrolv1beta1.VpcPeeringBuilder{}).
				WithScope(kymaName).
				WithRemoteRef("skr-namespace", "skr-aws-ip-range").
				WithDetails(localKcpNetworkName, infra.KCP().Namespace(), remoteKcpNetworkName, infra.KCP().Namespace(), "", false).
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					WithName(kcpPeeringName),
				).Should(Succeed())

		})

		By("Then KCP VpcPeering is in missing local network error state", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering, NewObjActions(),
					HavingCondition(cloudcontrolv1beta1.ConditionTypeError, metav1.ConditionTrue, cloudcontrolv1beta1.ReasonMissingDependency, "Local network not found"),
				).Should(Succeed())
		})

		var localKcpNet *cloudcontrolv1beta1.Network

		By("When local KCP Network is created", func() {
			kcpnetwork.Ignore.AddName(localKcpNetworkName)
			localKcpNet = (&cloudcontrolv1beta1.NetworkBuilder{}).
				WithScope(scope.Name).
				WithAwsRef(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region, scope.Spec.Scope.Aws.Network.VPC.Id, localKcpNetworkName).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet, WithName(localKcpNetworkName)).
				Should(Succeed())
		})

		By("Then KCP VpcPeering is in waiting local network to be ready error state", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering, NewObjActions(),
					HavingCondition(cloudcontrolv1beta1.ConditionTypeError, metav1.ConditionTrue, cloudcontrolv1beta1.ReasonWaitingDependency, "Local network not ready"),
				).Should(Succeed())
		})

		By("When local KCP Network is Ready", func() {
			kcpnetwork.Ignore.RemoveName(localKcpNetworkName)
			// trigger the reconciliation
			_, err := composed.PatchObjAddAnnotation(infra.Ctx(), "test", "1", localKcpNet, infra.KCP().Client())
			Expect(err).To(Succeed())
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet, NewObjActions(), HavingState("Ready")).
				Should(Succeed(), "expected local kcp network to become ready but it didn't")
		})

		By("Then KCP VpcPeering is in missing remote network error state", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering, NewObjActions(),
					HavingCondition(cloudcontrolv1beta1.ConditionTypeError, metav1.ConditionTrue, cloudcontrolv1beta1.ReasonMissingDependency, "Remote network not found"),
				).Should(Succeed())
		})

		var remoteKcpNet *cloudcontrolv1beta1.Network

		By("When remote KCP Network is created", func() {
			kcpnetwork.Ignore.AddName(remoteKcpNetworkName)
			remoteKcpNet = (&cloudcontrolv1beta1.NetworkBuilder{}).
				WithScope(scope.Name).
				WithAwsRef(remoteAccountId, remoteRegion, remoteVpcId, remoteVpcName).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet, WithName(remoteKcpNetworkName)).
				Should(Succeed())
		})

		By("Then KCP VpcPeering is in waiting remote network to be ready error state", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering, NewObjActions(),
					HavingCondition(cloudcontrolv1beta1.ConditionTypeError, metav1.ConditionTrue, cloudcontrolv1beta1.ReasonWaitingDependency, "Remote network not ready"),
				).Should(Succeed())
		})

		By("When remote KCP Network is Ready", func() {
			kcpnetwork.Ignore.RemoveName(remoteKcpNetworkName)
			// trigger the reconciliation
			_, err := composed.PatchObjAddAnnotation(infra.Ctx(), "test", "1", remoteKcpNet, infra.KCP().Client())
			Expect(err).
				To(Succeed())
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet, NewObjActions(), HavingState("Ready")).
				Should(Succeed(), "expected remote kcp network to become ready but it didn't")
		})

		By("Then KCP VpcPeering is initiating-request", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HaveFinalizer(cloudcontrolv1beta1.FinalizerName),
					HavingState(string(ec2Types.VpcPeeringConnectionStateReasonCodeInitiatingRequest)),
				).
				Should(Succeed())
		})

		By("When AWS VPC Peering state is Connected", func() {
			infra.AwsMock().SetVpcPeeringConnectionActive(infra.Ctx(), ptr.To(vpcId), ptr.To(remoteVpcId))
		})

		By("Then KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then KCP VpcPeering has status.vpcId equal to existing AWS VPC id", func() {
			Expect(kcpPeering.Status.VpcId).To(Equal(vpcId))
		})

		list, _ := infra.AwsMock().DescribeVpcPeeringConnections(infra.Ctx())

		var connection ec2Types.VpcPeeringConnection
		for _, p := range list {
			if kcpPeering.Status.Id == ptr.Deref(p.VpcPeeringConnectionId, "") {
				connection = p
			}
		}

		By("And Then found VpcPeeringConnection has AccepterVpcInfo.VpcId equals remote vpc id", func() {
			Expect(*connection.AccepterVpcInfo.VpcId).To(Equal(remoteVpcId))
		})

		By("And Then KCP VpcPeering has status.Id equal to existing AWS Connection id", func() {
			Expect(kcpPeering.Status.Id).To(Equal(ptr.Deref(connection.VpcPeeringConnectionId, "xxx")))
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
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "failed deleting VpcPeering")
		})

		By("Then VpcPeering does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "expected VpcPeering not to exist (be deleted), but it still exists")
		})

		By("And Then VpcPeeringConnection is deleted", func() {
			list, _ = infra.AwsMock().DescribeVpcPeeringConnections(infra.Ctx())

			found := pie.Any(list, func(x ec2Types.VpcPeeringConnection) bool {
				return ptr.Deref(x.VpcPeeringConnectionId, "xxx") == kcpPeering.Status.Id
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
