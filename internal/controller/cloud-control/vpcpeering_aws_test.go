package cloudcontrol

import (
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

var _ = Describe("Feature: KCP VpcPeering", func() {

	It("Scenario: KCP AWS VpcPeering is created", func() {
		const (
			kymaName        = "09bdb13e-8a51-4920-852d-b170433d1236"
			vpcId           = "26ce833e-07d1-4493-98ee-f9d6f11a6987"
			remoteVpcId     = "6e6d1748-9912-4957-9075-b97a6fac8ac1"
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
				"10.180.0.0/16",
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

		By("And Given AWS remote VPC exists", func() {
			infra.AwsMock().AddVpc(
				remoteVpcId,
				"10.200.0.0/16",
				awsutil.Ec2Tags("Name", "Remote Network Name"),
				nil,
			)
			infra.AwsMock().AddVpc(
				"wrong3",
				"10.200.0.0/16",
				awsutil.Ec2Tags("Name", "wrong3"),
				nil,
			)
		})

		vpcpeeringName := "b76ff161-c288-44fa-a295-8df2076af6a5"
		vpcpeering := &cloudcontrolv1beta1.VpcPeering{}

		By("When KCP VpcPeering is created", func() {
			Eventually(CreateKcpVpcPeering).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					WithName(vpcpeeringName),
					WithKcpVpcPeeringRemoteRef("skr-namespace", "skr-aws-ip-range"),
					WithKcpVpcPeeringSpecScope(kymaName),
					WithKcpVpcPeeringSpecAws(remoteVpcId, remoteAccountId, remoteRegion),
				).
				Should(Succeed())
		})

		By("Then KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then KCP VpcPeering has status.vpcId equal to existing AWS VPC id", func() {
			Expect(vpcpeering.Status.VpcId).To(Equal(vpcId))
		})

		list, _ := infra.AwsMock().DescribeVpcPeeringConnections(infra.Ctx())

		var connection ec2Types.VpcPeeringConnection
		for _, p := range list {
			if vpcpeering.Status.ConnectionId == pointer.StringDeref(p.VpcPeeringConnectionId, "") {
				connection = p
			}
		}
		By("And Then found VpcPeeringConnection has AccepterVpcInfo.VpcId equals remote vpc id", func() {
			Expect(*connection.AccepterVpcInfo.VpcId).To(Equal(remoteVpcId))
		})

		By("And Then KCP VpcPeering has status.ConnectionId equal to existing AWS Connection id", func() {
			Expect(vpcpeering.Status.ConnectionId).To(Equal(pointer.StringDeref(connection.VpcPeeringConnectionId, "xxx")))
		})
	})

})
