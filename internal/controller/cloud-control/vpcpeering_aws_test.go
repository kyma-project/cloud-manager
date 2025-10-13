package cloudcontrol

import (
	"errors"
	"fmt"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/feature"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	kcpnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/network"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("Feature: KCP VpcPeering", func() {

	It("Scenario: KCP AWS VpcPeering is created and deleted", func() {
		const (
			kymaName             = "09bdb13e-8a51-4920-852d-b170433d1236"
			kcpPeeringName       = "1839c399-c52e-4b43-b156-4b51027508cd"
			localVpcId           = "vpc-c0c7d75db0832988d"
			localVpcCidr         = "10.180.0.0/16"
			localVpcCidr2        = "10.182.0.0/16"
			remoteVpcId          = "vpc-2c41e43fcd5340f8f"
			remoteVpcCidr        = "10.200.0.0/16"
			remoteVpcCidr2       = "10.201.0.0/16"
			remoteAccountId      = "444455556666"
			remoteRegion         = "eu-west1"
			localMainRouteTable  = "rtb-c6606c725da27ff10"
			localRouteTable      = "rtb-0c65354e2979d9c83"
			remoteMainRouteTable = "rtb-69a1e8a929a9eb5ed"
			remoteRouteTable     = "rtb-ae17300793a424248"
			wrong1VpcId          = "wrong1"
			wrong1Cidr           = "10.200.0.0/16"
			wrong2VpcId          = "wrong2"
			wrong2Cidr           = "10.200.0.0/16"
			wrong2RouteTable     = "rtb-ae17300793a424240"
			wrong3VpcId          = "wrong3"
			wrong3Cidr           = "10.200.0.0/16"
			wrong3RouteTable     = "rtb-ae17300793a424247"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		vpcName := scope.Spec.Scope.Aws.VpcNetwork
		remoteVpcName := "Remote Network Name"

		awsMockLocal := infra.AwsMock().MockConfigs(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region)
		awsMockRemote := infra.AwsMock().MockConfigs(remoteAccountId, remoteRegion)

		By("And Given AWS VPC exists", func() {
			awsMockLocal.AddVpc(
				wrong1VpcId,
				wrong1Cidr,
				awsutil.Ec2Tags("Name", "wrong1"),
				nil,
			)
			awsMockLocal.AddVpc(
				localVpcId,
				localVpcCidr,
				awsutil.Ec2Tags("Name", vpcName),
				awsmock.VpcSubnetsFromScope(scope),
			)
			awsMockLocal.AddVpc(
				wrong2VpcId,
				wrong2Cidr,
				awsutil.Ec2Tags("Name", "wrong2"),
				nil,
			)
		})

		By("And Given AWS VPC additional CIDR exists", func() {
			_, err := awsMockLocal.AssociateVpcCidrBlock(infra.Ctx(), localVpcId, localVpcCidr2)
			Expect(err).ToNot(HaveOccurred())
		})

		By("And Given AWS route table exists", func() {

			awsMockLocal.AddRouteTable(
				ptr.To(localMainRouteTable),
				ptr.To(localVpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", vpcName), "1"),
				[]ec2types.RouteTableAssociation{
					{
						Main: ptr.To(true),
					},
				})

			awsMockLocal.AddRouteTable(
				ptr.To(wrong2RouteTable),
				ptr.To(wrong2VpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", wrong2VpcId), "1"),
				[]ec2types.RouteTableAssociation{})

			awsMockLocal.AddRouteTable(
				ptr.To(localRouteTable),
				ptr.To(localVpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", vpcName), "1"),
				[]ec2types.RouteTableAssociation{})
		})

		By("And Given AWS remote VPC exists", func() {
			awsMockRemote.AddVpc(
				remoteVpcId,
				remoteVpcCidr,
				awsutil.Ec2Tags("Name", remoteVpcName, kymaName, kymaName),
				nil,
			)
			awsMockRemote.AddVpc(
				wrong3VpcId,
				wrong3Cidr,
				awsutil.Ec2Tags("Name", "wrong3"),
				nil,
			)
		})

		By("And Given AWS remote VPC additional cidr exists", func() {
			_, err := awsMockRemote.AssociateVpcCidrBlock(infra.Ctx(), remoteVpcId, remoteVpcCidr2)
			Expect(err).ToNot(HaveOccurred())
		})

		By("And Given AWS remote route table exists", func() {

			awsMockRemote.AddRouteTable(
				ptr.To(remoteMainRouteTable),
				ptr.To(remoteVpcId),
				awsutil.Ec2Tags(),
				[]ec2types.RouteTableAssociation{
					{
						Main: ptr.To(true),
					},
				})

			awsMockRemote.AddRouteTable(
				ptr.To(wrong3RouteTable),
				ptr.To(wrong3VpcId),
				awsutil.Ec2Tags(),
				[]ec2types.RouteTableAssociation{})

			awsMockRemote.AddRouteTable(
				ptr.To(remoteRouteTable),
				ptr.To(remoteVpcId),
				awsutil.Ec2Tags(),
				[]ec2types.RouteTableAssociation{})
		})

		localKcpNetworkName := common.KcpNetworkKymaCommonName(scope.Name)
		remoteKcpNetworkName := scope.Name + "--remote"

		var kcpPeering *cloudcontrolv1beta1.VpcPeering

		By("When KCP VpcPeering is created", func() {
			kcpPeering = (&cloudcontrolv1beta1.VpcPeeringBuilder{}).
				WithScope(kymaName).
				WithRemoteRef("skr-namespace", "skr-aws-ip-range").
				WithDetails(localKcpNetworkName, infra.KCP().Namespace(), remoteKcpNetworkName, infra.KCP().Namespace(), "", false, true).
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					WithName(kcpPeeringName),
				).Should(Succeed())

		})

		var localKcpNet *cloudcontrolv1beta1.Network

		By("When local KCP Network is created", func() {
			kcpnetwork.Ignore.AddName(localKcpNetworkName)
			localKcpNet = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(scope.Name).
				WithAwsRef(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region, scope.Spec.Scope.Aws.Network.VPC.Id, localKcpNetworkName).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet, WithName(localKcpNetworkName)).
				Should(Succeed())
		})

		By("When local KCP Network is ready", func() {
			kcpnetwork.Ignore.RemoveName(localKcpNetworkName)
			// trigger the reconciliation
			_, err := composed.PatchObjMergeAnnotation(infra.Ctx(), "test", "1", localKcpNet, infra.KCP().Client())
			Expect(err).To(Succeed())
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet, NewObjActions(), HavingState(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed(), "expected local kcp network to become ready but it didn't")
		})

		var remoteKcpNet *cloudcontrolv1beta1.Network

		By("When remote KCP Network is created", func() {
			kcpnetwork.Ignore.AddName(remoteKcpNetworkName)
			remoteKcpNet = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(scope.Name).
				WithAwsRef(remoteAccountId, remoteRegion, remoteVpcId, remoteVpcName).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet, WithName(remoteKcpNetworkName)).
				Should(Succeed())
		})

		By("When remote KCP Network is ready", func() {
			kcpnetwork.Ignore.RemoveName(remoteKcpNetworkName)
			// trigger the reconciliation
			_, err := composed.PatchObjMergeAnnotation(infra.Ctx(), "test", "1", remoteKcpNet, infra.KCP().Client())
			Expect(err).
				To(Succeed())
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet, NewObjActions(), HavingState(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed(), "expected remote kcp network to become ready but it didn't")
		})

		By("Then KCP VpcPeering have status id", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingKcpVpcPeeringStatusIdNotEmpty(),
				).Should(Succeed())
		})

		By("When remote VpcPeeringConnection is initiated", func() {
			awsMockRemote.InitiateVpcPeeringConnection(kcpPeering.Status.Id, localVpcId, remoteVpcId)
		})

		By("When AWS VPC Peering state is active", func() {
			Expect(
				awsMockLocal.SetVpcPeeringConnectionStatusCode(localVpcId, remoteVpcId, ec2types.VpcPeeringConnectionStateReasonCodeActive),
			).NotTo(HaveOccurred())

			Expect(
				awsMockRemote.SetVpcPeeringConnectionStatusCode(localVpcId, remoteVpcId, ec2types.VpcPeeringConnectionStateReasonCodeActive),
			).NotTo(HaveOccurred())
		})

		By("Then KCP VpcPeering is ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then KCP VpcPeering status id equals to status remoteId", func() {
			Expect(kcpPeering.Status.Id).To(Equal(kcpPeering.Status.RemoteId))
		})

		By("And Then KCP VpcPeering status vpcId equals to existing AWS VPC id", func() {
			Expect(kcpPeering.Status.VpcId).To(Equal(localVpcId))
		})

		By("And Then found local VpcPeeringConnection AccepterVpcInfo VpcId equals to remote VPC id", func() {
			localPeering, _ := awsMockLocal.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(*localPeering.AccepterVpcInfo.VpcId).To(Equal(remoteVpcId))
		})

		By("And Then local route tables have peering routes to remote VPC CIDRs", func() {

			Expect(awsMockLocal.GetRoute(localVpcId, localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				NotTo(BeNil(), fmt.Sprintf("Local main route table should have peering route to %s", remoteVpcCidr))

			Expect(awsMockLocal.GetRoute(localVpcId, localRouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				NotTo(BeNil(), fmt.Sprintf("Local route table should have peering route to %s", remoteVpcCidr))

			Expect(awsMockLocal.GetRoute(wrong2VpcId, wrong2RouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				To(BeNil(), fmt.Sprintf("Route table %s should not have route with target %s and destination %s",
					wrong2RouteTable, kcpPeering.Status.Id, remoteVpcCidr))
		})

		By("And Then local route tables have peering routes to remote VPC additional CIDRs", func() {
			// Additional CIDR blocks
			Expect(awsMockLocal.GetRoute(localVpcId, localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr2)).
				NotTo(BeNil(), fmt.Sprintf("Local route table %s should have peering route to %s",
					localMainRouteTable, remoteVpcCidr2))

			Expect(awsMockLocal.GetRoute(localVpcId, localRouteTable, kcpPeering.Status.Id, remoteVpcCidr2)).
				NotTo(BeNil(), fmt.Sprintf("Local route table %s should have peering route to %s",
					localRouteTable, remoteVpcCidr2))

			Expect(awsMockLocal.GetRoute(wrong2VpcId, wrong2RouteTable, kcpPeering.Status.Id, remoteVpcCidr2)).
				To(BeNil(), fmt.Sprintf("Route table should not have route with target %s", remoteVpcCidr))

		})

		By("And Then remote route tables have peering routes to local VPC CIDR", func() {
			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteMainRouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				NotTo(BeNil(), fmt.Sprintf("Remote route table should have peering route to %s", remoteVpcCidr))

			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteRouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				NotTo(BeNil(), fmt.Sprintf("Remote route table should have peering route %s", remoteVpcCidr))

			Expect(awsMockRemote.GetRoute(wrong3VpcId, wrong3RouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				To(BeNil(), fmt.Sprintf("Route table %s should not be modified", wrong2RouteTable))
		})

		By("And Then remote route tables have no peering routes to local VPC additional CIDRs", func() {
			// Additional CIDR blocks
			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteMainRouteTable, kcpPeering.Status.RemoteId, localVpcCidr2)).
				To(BeNil(), fmt.Sprintf("Remote route table should not have peering route to %s", localVpcCidr2))

			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteRouteTable, kcpPeering.Status.RemoteId, localVpcCidr2)).
				To(BeNil(), fmt.Sprintf("Remote route table should not have peering route to %s", localVpcCidr2))

			Expect(awsMockRemote.GetRoute(wrong3VpcId, wrong3RouteTable, kcpPeering.Status.RemoteId, localVpcCidr2)).
				To(BeNil(), fmt.Sprintf("Route table %s should not be modified", wrong2RouteTable))
		})

		// DELETE

		By("When KCP VpcPeering is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "failed deleting VpcPeering")
		})

		By("Then VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "expected VpcPeering not to exist (be deleted), but it still exists")
		})

		By("And Then local VpcPeeringConnection is deleted", func() {
			localPeering, _ := awsMockLocal.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(localPeering).To(BeNil())
		})

		By("And Then local route tables have no peering routes to remote VPC CIDRs", func() {
			Expect(awsMockLocal.GetRoute(localVpcId, localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				To(BeNil())

			Expect(awsMockLocal.GetRoute(localVpcId, localRouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				To(BeNil())

			Expect(awsMockLocal.GetRoute(localVpcId, localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr2)).
				To(BeNil())

			Expect(awsMockLocal.GetRoute(localVpcId, localRouteTable, kcpPeering.Status.Id, remoteVpcCidr2)).
				To(BeNil())
		})

		By("And Then remote VpcPeeringConnection is deleted", func() {
			remotePeering, _ := awsMockRemote.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(remotePeering).To(BeNil())
		})

		By("And Then remote route tables have no peering routes to local VPC CIDRs", func() {
			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteMainRouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				To(BeNil())

			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteRouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				To(BeNil())
		})
	})

	// When prevent deletion of KCP Network while used by VpcPeering is implemented, this test case
	// is obsolete, but keeping it just in case, but with Network reconciler ignoring the created
	// networks, so they can be deleted while used by VpcPeering
	It("Scenario: KCP AWS VpcPeering is deleted when local and remote networks are missing", func() {
		const (
			kymaName             = "76f1dec7-c7d3-4129-9730-478f4cba241a"
			kcpPeeringName       = "f658c189-0f09-4c4b-8da6-49b3db61546d"
			localVpcId           = "vpc-7e9d1ce03b49ae18d"
			localVpcCidr         = "10.180.0.0/16"
			remoteVpcId          = "vpc-3a1cdc66b2778658e"
			remoteVpcCidr        = "10.200.0.0/16"
			remoteAccountId      = "777755556666"
			remoteRegion         = "eu-west1"
			localMainRouteTable  = "rtb-007a6396ac2021245"
			localRouteTable      = "rtb-c44da7a78dbf49bde"
			remoteMainRouteTable = "rtb-c0b83bb46e6d208b9"
			remoteRouteTable     = "rtb-30b3c0b6d895ed2d0"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		vpcName := scope.Spec.Scope.Aws.VpcNetwork
		remoteVpcName := "Remote Network Name"

		awsMockLocal := infra.AwsMock().MockConfigs(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region)
		awsMockRemote := infra.AwsMock().MockConfigs(remoteAccountId, remoteRegion)

		By("And Given AWS VPC exists", func() {
			awsMockLocal.AddVpc(
				localVpcId,
				localVpcCidr,
				awsutil.Ec2Tags("Name", vpcName),
				awsmock.VpcSubnetsFromScope(scope),
			)
		})

		By("And Given AWS route table exists", func() {
			awsMockLocal.AddRouteTable(
				ptr.To(localMainRouteTable),
				ptr.To(localVpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", vpcName), "1"),
				[]ec2types.RouteTableAssociation{
					{
						Main: ptr.To(true),
					},
				})

			awsMockLocal.AddRouteTable(
				ptr.To(localRouteTable),
				ptr.To(localVpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", vpcName), "1"),
				[]ec2types.RouteTableAssociation{})
		})

		By("And Given AWS remote VPC exists", func() {
			awsMockRemote.AddVpc(
				remoteVpcId,
				remoteVpcCidr,
				awsutil.Ec2Tags("Name", remoteVpcName, kymaName, kymaName),
				nil,
			)
		})

		By("And Given AWS remote route table exists", func() {

			awsMockRemote.AddRouteTable(
				ptr.To(remoteMainRouteTable),
				ptr.To(remoteVpcId),
				awsutil.Ec2Tags(),
				[]ec2types.RouteTableAssociation{
					{
						Main: ptr.To(true),
					},
				})

			awsMockRemote.AddRouteTable(
				ptr.To(remoteRouteTable),
				ptr.To(remoteVpcId),
				awsutil.Ec2Tags(),
				[]ec2types.RouteTableAssociation{})
		})

		localKcpNetworkName := common.KcpNetworkKymaCommonName(scope.Name)
		remoteKcpNetworkName := scope.Name + "--remote"

		var localKcpNet *cloudcontrolv1beta1.Network

		By("And Given local KCP Network exists", func() {
			// must tell reconciler to ignore it, since it would prevent deletion when used by peering
			kcpnetwork.Ignore.AddName(localKcpNetworkName)
			localKcpNet = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(scope.Name).
				WithAwsRef(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region, scope.Spec.Scope.Aws.Network.VPC.Id, localKcpNetworkName).
				Build()
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), localKcpNet, WithName(localKcpNetworkName))).
				To(Succeed())

			localKcpNet.Status.Network = localKcpNet.Spec.Network.Reference.DeepCopy()
			localKcpNet.Status.State = string(cloudcontrolv1beta1.StateReady)
			meta.SetStatusCondition(&localKcpNet.Status.Conditions, metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonReady,
				Message: cloudcontrolv1beta1.ReasonReady,
			})
			Expect(composed.PatchObjStatus(infra.Ctx(), localKcpNet, infra.KCP().Client())).
				To(Succeed())
		})

		var remoteKcpNet *cloudcontrolv1beta1.Network

		By("And Given remote KCP Network exists", func() {
			// must tell reconciler to ignore it, since it would prevent deletion when used by peering
			kcpnetwork.Ignore.AddName(remoteKcpNetworkName)
			remoteKcpNet = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(scope.Name).
				WithAwsRef(remoteAccountId, remoteRegion, remoteVpcId, remoteVpcName).
				Build()
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), remoteKcpNet, WithName(remoteKcpNetworkName))).
				Should(Succeed())

			remoteKcpNet.Status.Network = remoteKcpNet.Spec.Network.Reference.DeepCopy()
			remoteKcpNet.Status.State = string(cloudcontrolv1beta1.StateReady)
			meta.SetStatusCondition(&remoteKcpNet.Status.Conditions, metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonReady,
				Message: cloudcontrolv1beta1.ReasonReady,
			})
			Expect(composed.PatchObjStatus(infra.Ctx(), remoteKcpNet, infra.KCP().Client())).
				To(Succeed())
		})

		var kcpPeering *cloudcontrolv1beta1.VpcPeering

		By("When KCP VpcPeering is created", func() {
			kcpPeering = (&cloudcontrolv1beta1.VpcPeeringBuilder{}).
				WithScope(kymaName).
				WithRemoteRef("skr-namespace", "skr-aws-ip-range").
				WithDetails(localKcpNetworkName, infra.KCP().Namespace(), remoteKcpNetworkName, infra.KCP().Namespace(), "", false, true).
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					WithName(kcpPeeringName),
				).Should(Succeed())

		})

		By("Then KCP VpcPeering has status id", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingKcpVpcPeeringStatusIdNotEmpty(),
				).Should(Succeed())
		})

		By("When remote VpcPeeringConnection is initiated", func() {
			awsMockRemote.InitiateVpcPeeringConnection(kcpPeering.Status.Id, localVpcId, remoteVpcId)
		})

		By("When AWS VPC Peering state is active", func() {
			Expect(
				awsMockLocal.SetVpcPeeringConnectionStatusCode(localVpcId, remoteVpcId, ec2types.VpcPeeringConnectionStateReasonCodeActive),
			).NotTo(HaveOccurred())

			Expect(
				awsMockRemote.SetVpcPeeringConnectionStatusCode(localVpcId, remoteVpcId, ec2types.VpcPeeringConnectionStateReasonCodeActive),
			).NotTo(HaveOccurred())
		})

		By("Then KCP VpcPeering is ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then KCP VpcPeering status id equals to remoteId", func() {
			Expect(kcpPeering.Status.Id).To(Equal(kcpPeering.Status.RemoteId))
		})

		By("And Then KCP VpcPeering status vpcId equals to existing AWS VPC id", func() {
			Expect(kcpPeering.Status.VpcId).To(Equal(localVpcId))
		})

		By("And Then found local VpcPeeringConnection AccepterVpcInfo VpcId equals to remote vpc id", func() {
			localPeering, _ := awsMockLocal.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(*localPeering.AccepterVpcInfo.VpcId).To(Equal(remoteVpcId))
		})

		By("And Then local route tables have peering route to remote VPC CIDR", func() {
			Expect(awsMockLocal.GetRoute(localVpcId, localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				NotTo(BeNil(), fmt.Sprintf("Local main route table should have peering route to %s", remoteVpcCidr))

			Expect(awsMockLocal.GetRoute(localVpcId, localRouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				ToNot(BeNil(), fmt.Sprintf("Local route table should have peering route to %s", remoteVpcCidr))
		})

		By("And Then remote route tables have peering route to VPC CIDR", func() {
			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteMainRouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				NotTo(BeNil(), fmt.Sprintf("Remote route table should have peering route to %s", remoteVpcCidr))

			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteRouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				ToNot(BeNil(), fmt.Sprintf("Remote route table should have peering route to %s", remoteVpcCidr))
		})

		// Deleting KCP remote Network before VpcPeering deletion
		By("When KCP local Network is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet).
				Should(Succeed(), "failed deleting local KCP Network")
		})

		By("Then KCP local Network does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet).
				Should(Succeed(), "expected KCP local Network not to exist (be deleted), but it still exists")
		})

		// Deleting KCP remote Network before VpcPeering deletion
		By("When KCP remote Network is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet).
				Should(Succeed(), "failed deleting remote KCP Network")
		})

		By("Then KCP remote Network does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet).
				Should(Succeed(), "expected KCP remote Network not to exist (be deleted), but it still exists")
		})

		// DELETE

		By("When KCP VpcPeering is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "failed deleting VpcPeering")
		})

		By("Then VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "expected VpcPeering not to exist (be deleted), but it still exists")
		})

		By("And Then local VpcPeeringConnection is deleted", func() {
			localPeering, err := awsMockLocal.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(err).To(HaveOccurred())
			Expect(localPeering).To(BeNil())
		})

		By("And Then local route tables have no peering routes to remote VPC CIDRs", func() {
			Expect(awsMockLocal.GetRoute(localVpcId, localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				To(BeNil(), fmt.Sprintf("Local main route table should not have peering route to %s", remoteVpcCidr))

			Expect(awsMockLocal.GetRoute(localVpcId, localRouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				To(BeNil(), fmt.Sprintf("Local route table should not have peering route to %s", remoteVpcCidr))

		})

		// VpcPeeringConnection and Routes are not deleted since KCP remote Network is deleted previously
		By("And Then remote VpcPeeringConnection is not deleted", func() {
			remotePeering, err := awsMockRemote.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(remotePeering).NotTo(BeNil())
		})

		By("And Then remote route tables have peering routes to local VPC CIDR", func() {
			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteMainRouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				NotTo(BeNil(), fmt.Sprintf("Remote main route table should have peering route to %s", localVpcCidr))

			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteRouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				NotTo(BeNil(), fmt.Sprintf("Remote route table should have peering route to %s", localVpcCidr))

		})
	})

	It("Scenario: KCP AWS VpcPeering can be deleted when remote VPC Network authorization is revoked", func() {
		const (
			kymaName             = "50de99f8-0b35-4ac2-900e-793091f1a853"
			kcpPeeringName       = "b6689354-72cc-41ba-ae48-572fa7815a6c"
			localVpcId           = "vpc-1fe57eb9ec4b4d389"
			localVpcCidr         = "10.180.0.0/16"
			remoteVpcId          = "vpc-5a11d50637164d01a"
			remoteVpcCidr        = "10.200.0.0/16"
			remoteAccountId      = "777755556666"
			remoteRegion         = "eu-west1"
			localMainRouteTable  = "rtb-bb6743e182614c539"
			localRouteTable      = "rtb-d6d5d9e2492449b38"
			remoteMainRouteTable = "rtb-713e94a6caa54b27a"
			remoteRouteTable     = "rtb-14ff90610fc54a4cb"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		vpcName := scope.Spec.Scope.Aws.VpcNetwork
		remoteVpcName := "Remote Network Name"

		awsMockLocal := infra.AwsMock().MockConfigs(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region)
		awsMockRemote := infra.AwsMock().MockConfigs(remoteAccountId, remoteRegion)

		By("And Given AWS VPC exists", func() {
			awsMockLocal.AddVpc(
				localVpcId,
				localVpcCidr,
				awsutil.Ec2Tags("Name", vpcName),
				awsmock.VpcSubnetsFromScope(scope),
			)
		})

		By("And Given AWS route table exists", func() {
			awsMockLocal.AddRouteTable(
				ptr.To(localMainRouteTable),
				ptr.To(localVpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", vpcName), "1"),
				[]ec2types.RouteTableAssociation{
					{
						Main: ptr.To(true),
					},
				})

			awsMockLocal.AddRouteTable(
				ptr.To(localRouteTable),
				ptr.To(localVpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", vpcName), "1"),
				[]ec2types.RouteTableAssociation{})
		})

		By("And Given AWS remote VPC exists", func() {
			awsMockRemote.AddVpc(
				remoteVpcId,
				remoteVpcCidr,
				awsutil.Ec2Tags("Name", remoteVpcName, kymaName, kymaName),
				nil,
			)
		})

		By("And Given AWS remote route table exists", func() {

			awsMockRemote.AddRouteTable(
				ptr.To(remoteMainRouteTable),
				ptr.To(remoteVpcId),
				awsutil.Ec2Tags(),
				[]ec2types.RouteTableAssociation{
					{
						Main: ptr.To(true),
					},
				})

			awsMockRemote.AddRouteTable(
				ptr.To(remoteRouteTable),
				ptr.To(remoteVpcId),
				awsutil.Ec2Tags(),
				[]ec2types.RouteTableAssociation{})
		})

		localKcpNetworkName := common.KcpNetworkKymaCommonName(scope.Name)
		remoteKcpNetworkName := scope.Name + "--remote"

		var localKcpNet *cloudcontrolv1beta1.Network

		By("And Given local KCP Network exists", func() {
			localKcpNet = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(scope.Name).
				WithAwsRef(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region, scope.Spec.Scope.Aws.Network.VPC.Id, localKcpNetworkName).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet, WithName(localKcpNetworkName)).
				Should(Succeed())
		})

		By("And Given remote KCP Network is ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).Should(Succeed())
		})

		var remoteKcpNet *cloudcontrolv1beta1.Network

		By("And Given remote KCP Network exists", func() {
			remoteKcpNet = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(scope.Name).
				WithAwsRef(remoteAccountId, remoteRegion, remoteVpcId, remoteVpcName).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet, WithName(remoteKcpNetworkName)).
				Should(Succeed())
		})

		By("And Given remote KCP Network is ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).Should(Succeed())
		})

		var kcpPeering *cloudcontrolv1beta1.VpcPeering

		By("And Given KCP VpcPeering is created", func() {
			kcpPeering = (&cloudcontrolv1beta1.VpcPeeringBuilder{}).
				WithScope(kymaName).
				WithRemoteRef("skr-namespace", "skr-aws-ip-range").
				WithDetails(localKcpNetworkName, infra.KCP().Namespace(), remoteKcpNetworkName, infra.KCP().Namespace(), "", false, true).
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					WithName(kcpPeeringName),
				).Should(Succeed())

		})

		By("And Given KCP VpcPeering has status id", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingKcpVpcPeeringStatusIdNotEmpty(),
				).Should(Succeed())
		})

		By("And Given AWS VpcPeeringConnections are active", func() {

			// initiate remote vpc peering connection
			awsMockRemote.InitiateVpcPeeringConnection(kcpPeering.Status.Id, localVpcId, remoteVpcId)

			// change local vpc peering status to pending-acceptance (not necessary but leaving it for the clarity)

			Expect(
				awsMockLocal.SetVpcPeeringConnectionStatusCode(localVpcId, remoteVpcId, ec2types.VpcPeeringConnectionStateReasonCodePendingAcceptance),
			).NotTo(HaveOccurred())

			// sets vpc peering connections active
			Expect(
				awsMockLocal.SetVpcPeeringConnectionStatusCode(localVpcId, remoteVpcId, ec2types.VpcPeeringConnectionStateReasonCodeActive),
			).NotTo(HaveOccurred())

			Expect(
				awsMockRemote.SetVpcPeeringConnectionStatusCode(localVpcId, remoteVpcId, ec2types.VpcPeeringConnectionStateReasonCodeActive),
			).NotTo(HaveOccurred())
		})

		By("And Given VpcPeering is ready", func() {
			Eventually(LoadAndCheck, "2s").
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed())
		})

		By("And Given remote VpcPeeringConnection exists", func() {
			remotePeering, err := awsMockRemote.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(remotePeering).NotTo(BeNil())
		})

		By("And Given remote permissions are removed", func() {
			awsMockRemote.SetVpcPeeringConnectionError(kcpPeering.Status.Id, errors.New("peering error"))
			awsMockRemote.SetVpcError(remoteVpcId, errors.New("vpc error"))
		})

		// DELETE

		By("When KCP VpcPeering is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "failed deleting VpcPeering")
		})

		By("Then VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "expected VpcPeering not to exist (be deleted), but it still exists")
		})

		By("And Then local VpcPeeringConnection is deleted", func() {
			localPeering, _ := awsMockLocal.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(localPeering).To(BeNil())
		})

		// remove peering connection error from mock so that we could verify it
		awsMockRemote.SetVpcPeeringConnectionError(kcpPeering.Status.Id, nil)

		By("And Then remote VpcPeeringConnection is not deleted", func() {
			remotePeering, _ := awsMockRemote.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(remotePeering).NotTo(BeNil())
		})

		By("And Then remote route tables have routes to local VPC CIDR", func() {
			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteMainRouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				NotTo(BeNil(), fmt.Sprintf("Remote route table have peering route to %s", localVpcCidr))
		})
	})

	It("Scenario: KCP AWS VpcPeering remote route table update strategy MATCHED", func() {
		const (
			kymaName               = "612573aa-58be-4670-b7b2-ca4c60fb8b99"
			kcpPeeringName         = "0cb1ecb4-de6d-4146-93c7-df2a71e6f83e"
			localVpcId             = "vpc-5b6b20ad2b85f945b"
			localVpcCidr           = "10.180.0.0/16"
			remoteVpcId            = "vpc-9c7b1757ffb17f3db"
			remoteVpcCidr          = "10.200.0.0/16"
			remoteAccountId        = "777755557777"
			remoteRegion           = "eu-west1"
			localMainRouteTable    = "rtb-7ce283587d14d4517"
			localRouteTable        = "rtb-1c690daffb668e1cc"
			remoteMainRouteTable   = "rtb-d1605c3e2153551ee"
			remoteRouteTable       = "rtb-e03bfb82225944cdf"
			remoteRouteTableTagged = "rtb-04214f38752ba4e85"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		vpcName := scope.Spec.Scope.Aws.VpcNetwork
		remoteVpcName := "Remote Network Name"

		awsMockLocal := infra.AwsMock().MockConfigs(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region)
		awsMockRemote := infra.AwsMock().MockConfigs(remoteAccountId, remoteRegion)

		By("And Given AWS VPC exists", func() {
			awsMockLocal.AddVpc(
				localVpcId,
				localVpcCidr,
				awsutil.Ec2Tags("Name", vpcName),
				awsmock.VpcSubnetsFromScope(scope),
			)
		})

		By("And Given AWS route table exists", func() {
			awsMockLocal.AddRouteTable(
				ptr.To(localMainRouteTable),
				ptr.To(localVpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", vpcName), "1"),
				[]ec2types.RouteTableAssociation{
					{
						Main: ptr.To(true),
					},
				})

			awsMockLocal.AddRouteTable(
				ptr.To(localRouteTable),
				ptr.To(localVpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", vpcName), "1"),
				[]ec2types.RouteTableAssociation{})
		})

		By("And Given AWS remote VPC exists", func() {
			awsMockRemote.AddVpc(
				remoteVpcId,
				remoteVpcCidr,
				awsutil.Ec2Tags("Name", remoteVpcName, kymaName, kymaName),
				nil,
			)
		})

		By("And Given AWS remote route table exists", func() {

			awsMockRemote.AddRouteTable(
				ptr.To(remoteMainRouteTable),
				ptr.To(remoteVpcId),
				awsutil.Ec2Tags(),
				[]ec2types.RouteTableAssociation{
					{
						Main: ptr.To(true),
					},
				})

			awsMockRemote.AddRouteTable(
				ptr.To(remoteRouteTableTagged),
				ptr.To(remoteVpcId),
				awsutil.Ec2Tags(kymaName, kymaName), // tag remote route table
				[]ec2types.RouteTableAssociation{})

			awsMockRemote.AddRouteTable(
				ptr.To(remoteRouteTable),
				ptr.To(remoteVpcId),
				awsutil.Ec2Tags(),
				[]ec2types.RouteTableAssociation{})
		})

		localKcpNetworkName := common.KcpNetworkKymaCommonName(scope.Name)
		remoteKcpNetworkName := scope.Name + "--remote"

		var localKcpNet *cloudcontrolv1beta1.Network

		By("And Given local KCP Network exists", func() {
			localKcpNet = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(scope.Name).
				WithAwsRef(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region, scope.Spec.Scope.Aws.Network.VPC.Id, localKcpNetworkName).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet, WithName(localKcpNetworkName)).
				Should(Succeed())
		})

		By("And Given remote KCP Network is ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).Should(Succeed())
		})

		var remoteKcpNet *cloudcontrolv1beta1.Network

		By("And Given remote KCP Network exists", func() {
			remoteKcpNet = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(scope.Name).
				WithAwsRef(remoteAccountId, remoteRegion, remoteVpcId, remoteVpcName).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet, WithName(remoteKcpNetworkName)).
				Should(Succeed())
		})

		By("And Given remote KCP Network is ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).Should(Succeed())
		})

		var kcpPeering *cloudcontrolv1beta1.VpcPeering

		By("And Given KCP VpcPeering is created", func() {
			kcpPeering = (&cloudcontrolv1beta1.VpcPeeringBuilder{}).
				WithScope(kymaName).
				WithRemoteRef("skr-namespace", "skr-aws-ip-range").
				WithDetails(localKcpNetworkName, infra.KCP().Namespace(), remoteKcpNetworkName, infra.KCP().Namespace(), "", false, false).
				WithRemoteRouteTableUpdateStrategy(cloudcontrolv1beta1.AwsRouteTableUpdateStrategyMatched).
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					WithName(kcpPeeringName),
				).Should(Succeed())

		})

		By("And Given KCP VpcPeering have status id", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingKcpVpcPeeringStatusIdNotEmpty(),
				).Should(Succeed())
		})

		By("And Given AWS VpcPeeringConnections are active", func() {

			// initiate remote vpc peering connection
			awsMockRemote.InitiateVpcPeeringConnection(kcpPeering.Status.Id, localVpcId, remoteVpcId)

			// change local vpc peering status to pending-acceptance (not necessary but leaving it for the clarity)

			Expect(
				awsMockLocal.SetVpcPeeringConnectionStatusCode(localVpcId, remoteVpcId, ec2types.VpcPeeringConnectionStateReasonCodePendingAcceptance),
			).NotTo(HaveOccurred())

			// sets vpc peering connections active
			Expect(
				awsMockLocal.SetVpcPeeringConnectionStatusCode(localVpcId, remoteVpcId, ec2types.VpcPeeringConnectionStateReasonCodeActive),
			).NotTo(HaveOccurred())

			Expect(
				awsMockRemote.SetVpcPeeringConnectionStatusCode(localVpcId, remoteVpcId, ec2types.VpcPeeringConnectionStateReasonCodeActive),
			).NotTo(HaveOccurred())
		})

		By("When VpcPeering is ready", func() {
			Eventually(LoadAndCheck, "2s").
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed())
		})

		By("Then remote VpcPeeringConnection exists", func() {
			remotePeering, _ := awsMockRemote.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(remotePeering).NotTo(BeNil())
		})

		By("And Then remote tagged route table has peering route to local VPC CIDR", func() {
			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteRouteTableTagged, kcpPeering.Status.RemoteId, localVpcCidr)).
				NotTo(BeNil())
		})

		By("And Then remote untagged route tables do not have peering routes to local VPC CIDR", func() {
			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteRouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				To(BeNil())

			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteMainRouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				To(BeNil())

		})

		// DELETE

		By("When KCP VpcPeering is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "failed deleting VpcPeering")
		})

		By("Then VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "expected VpcPeering not to exist (be deleted), but it still exists")
		})

		By("And Then local VpcPeeringConnection is deleted", func() {
			localPeering, _ := awsMockLocal.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(localPeering).To(BeNil())
		})

		By("And Then remote VpcPeeringConnection is not deleted", func() {
			remotePeering, _ := awsMockRemote.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(remotePeering).NotTo(BeNil())
		})

		By("And Then remote route table has route to local VPC CIDR", func() {
			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteRouteTableTagged, kcpPeering.Status.Id, localVpcCidr)).
				NotTo(BeNil())
		})

	})

	It("Scenario: KCP AWS VpcPeering update route tables", func() {
		const (
			kymaName               = "deeae7fe-7744-45cb-bb10-62f0cb160979"
			kcpPeeringName         = "b88be55c-80e5-43b5-acde-aaefea92b991"
			localVpcId             = "vpc-5b6b20ad2b85f9451"
			localVpcCidr           = "10.180.0.0/16"
			remoteVpcId            = "vpc-9c7b1757ffb17f3dc"
			remoteVpcCidr          = "10.200.0.0/16"
			remoteVpcCidr2         = "10.201.0.0/16"
			remoteVpcCidr3         = "10.202.0.0/16"
			remoteAccountId        = "777755558888"
			remoteRegion           = "eu-west1"
			localMainRouteTable    = "rtb-b88be55c80e543b5a"
			localRouteTable        = "rtb-1c0daaaefea92b991"
			remoteMainRouteTable   = "rtb-d160cabc72657d3b9"
			remoteRouteTable       = "rtb-e03bf0c1dd5385ee0"
			remoteRouteTableTagged = "rtb-042142cb73d698886"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		vpcName := scope.Spec.Scope.Aws.VpcNetwork
		remoteVpcName := "Remote Network Name"

		awsMockLocal := infra.AwsMock().MockConfigs(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region)
		awsMockRemote := infra.AwsMock().MockConfigs(remoteAccountId, remoteRegion)

		By("And Given local AWS VPC exists", func() {
			awsMockLocal.AddVpc(
				localVpcId,
				localVpcCidr,
				awsutil.Ec2Tags("Name", vpcName),
				awsmock.VpcSubnetsFromScope(scope),
			)
		})

		By("And Given local AWS route table exists", func() {
			awsMockLocal.AddRouteTable(
				ptr.To(localMainRouteTable),
				ptr.To(localVpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", vpcName), "1"),
				[]ec2types.RouteTableAssociation{
					{
						Main: ptr.To(true),
					},
				})

			awsMockLocal.AddRouteTable(
				ptr.To(localRouteTable),
				ptr.To(localVpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", vpcName), "1"),
				[]ec2types.RouteTableAssociation{})
		})

		By("And Given remote AWS VPC exists", func() {
			awsMockRemote.AddVpc(
				remoteVpcId,
				remoteVpcCidr,
				awsutil.Ec2Tags("Name", remoteVpcName, kymaName, kymaName),
				nil,
			)
		})

		var remoteVpcCidrBlockAssociation2 *ec2types.VpcCidrBlockAssociation
		By("And Given remote AWS VPC additional cidr exists", func() {
			var err error
			remoteVpcCidrBlockAssociation2, err = awsMockRemote.AssociateVpcCidrBlock(infra.Ctx(), remoteVpcId, remoteVpcCidr2)
			Expect(err).ToNot(HaveOccurred())
		})

		By("And Given remote AWS route table exists", func() {

			awsMockRemote.AddRouteTable(
				ptr.To(remoteMainRouteTable),
				ptr.To(remoteVpcId),
				awsutil.Ec2Tags(),
				[]ec2types.RouteTableAssociation{
					{
						Main: ptr.To(true),
					},
				})

			awsMockRemote.AddRouteTable(
				ptr.To(remoteRouteTableTagged),
				ptr.To(remoteVpcId),
				awsutil.Ec2Tags(kymaName, kymaName), // tag remote route table
				[]ec2types.RouteTableAssociation{})

			awsMockRemote.AddRouteTable(
				ptr.To(remoteRouteTable),
				ptr.To(remoteVpcId),
				awsutil.Ec2Tags(),
				[]ec2types.RouteTableAssociation{})
		})

		localKcpNetworkName := common.KcpNetworkKymaCommonName(scope.Name)
		remoteKcpNetworkName := scope.Name + "--remote"

		var localKcpNet *cloudcontrolv1beta1.Network

		By("And Given local KCP Network exists", func() {
			localKcpNet = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(scope.Name).
				WithAwsRef(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region, scope.Spec.Scope.Aws.Network.VPC.Id, localKcpNetworkName).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet, WithName(localKcpNetworkName)).
				Should(Succeed())
		})

		By("And Given remote KCP Network is ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).Should(Succeed())
		})

		var remoteKcpNet *cloudcontrolv1beta1.Network

		By("And Given remote KCP Network exists", func() {
			remoteKcpNet = cloudcontrolv1beta1.NewNetworkBuilder().
				WithScope(scope.Name).
				WithAwsRef(remoteAccountId, remoteRegion, remoteVpcId, remoteVpcName).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet, WithName(remoteKcpNetworkName)).
				Should(Succeed())
		})

		By("And Given remote KCP Network is ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).Should(Succeed())
		})

		var kcpPeering *cloudcontrolv1beta1.VpcPeering

		By("And Given KCP VpcPeering is created", func() {
			kcpPeering = (&cloudcontrolv1beta1.VpcPeeringBuilder{}).
				WithScope(kymaName).
				WithRemoteRef("skr-namespace", "skr-aws-ip-range").
				WithDetails(localKcpNetworkName, infra.KCP().Namespace(), remoteKcpNetworkName, infra.KCP().Namespace(), "", false, true).
				WithRemoteRouteTableUpdateStrategy(cloudcontrolv1beta1.AwsRouteTableUpdateStrategyMatched).
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					WithName(kcpPeeringName),
				).Should(Succeed())

		})

		By("And Given KCP VpcPeering have status id", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingKcpVpcPeeringStatusIdNotEmpty(),
				).Should(Succeed())
		})

		By("And Given AWS VPC peering connections are active", func() {

			// initiate remote vpc peering connection
			awsMockRemote.InitiateVpcPeeringConnection(kcpPeering.Status.Id, localVpcId, remoteVpcId)

			// change local vpc peering status to pending-acceptance (not necessary but leaving it for the clarity)
			Expect(
				awsMockLocal.SetVpcPeeringConnectionStatusCode(localVpcId, remoteVpcId, ec2types.VpcPeeringConnectionStateReasonCodePendingAcceptance),
			).NotTo(HaveOccurred())

			// sets vpc peering connections active
			Expect(
				awsMockLocal.SetVpcPeeringConnectionStatusCode(localVpcId, remoteVpcId, ec2types.VpcPeeringConnectionStateReasonCodeActive),
			).NotTo(HaveOccurred())

			Expect(
				awsMockRemote.SetVpcPeeringConnectionStatusCode(localVpcId, remoteVpcId, ec2types.VpcPeeringConnectionStateReasonCodeActive),
			).NotTo(HaveOccurred())
		})

		By("When VpcPeering is ready", func() {
			Eventually(LoadAndCheck, "2s").
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed())
		})

		By("Then remote AWS VPC peering connection exists", func() {
			remotePeering, _ := awsMockRemote.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(remotePeering).NotTo(BeNil())
		})

		By("And Then local route tables has routes matching remote VPC CIDRs", func() {
			Expect(awsMockLocal.GetRoute(localVpcId, localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				NotTo(BeNil(), fmt.Sprintf("Local main route table should have peering route to %s", remoteVpcCidr))

			Expect(awsMockLocal.GetRoute(localVpcId, localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr2)).
				NotTo(BeNil(), fmt.Sprintf("Local main route table should have peering route to %s", remoteVpcCidr2))

			Expect(awsMockLocal.GetRoute(localVpcId, localRouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				NotTo(BeNil(), fmt.Sprintf("Local route table should have peering route to %s", remoteVpcCidr))

			Expect(awsMockLocal.GetRoute(localVpcId, localRouteTable, kcpPeering.Status.Id, remoteVpcCidr2)).
				NotTo(BeNil(), fmt.Sprintf("Local route table should have peering route to %s", remoteVpcCidr2))

		})

		By("And Then remote tagged route table has routes matching local VPC CIDRs", func() {
			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteRouteTableTagged, kcpPeering.Status.RemoteId, localVpcCidr)).
				NotTo(BeNil(), fmt.Sprintf("Remote route table should have peering route %s", localVpcCidr))
		})

		By("And Then remote untagged route tables have no routes matching local VPC CIDRs", func() {
			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteMainRouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				To(BeNil(), fmt.Sprintf("Remote main route table should not have peering route %s", localVpcCidr))

			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteRouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				To(BeNil(), fmt.Sprintf("Remote route table should not have peering route %s", localVpcCidr))
		})

		By("And When remote VPC additional CIDR is added", func() {
			_, err := awsMockRemote.AssociateVpcCidrBlock(infra.Ctx(), remoteVpcId, remoteVpcCidr3)
			Expect(err).ToNot(HaveOccurred())
		})

		By("And When remote VPC additional CIDR is removed", func() {
			err := awsMockRemote.DisassociateVpcCidrBlockInput(infra.Ctx(), *remoteVpcCidrBlockAssociation2.AssociationId)
			Expect(err).ToNot(HaveOccurred())
		})

		// Cloud Manager reconciles =======================================

		By("Then Cloud Manager detects VPC CIDR changes and update routes", func() {
			// trigger the reconciliation
			_, err := composed.PatchObjMergeAnnotation(infra.Ctx(), "test", "1", kcpPeering, infra.KCP().Client())
			Expect(err).To(Succeed())

			// Removing ready condition to make sure that reconciliation happened
			Expect(meta.RemoveStatusCondition(&kcpPeering.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)).To(BeTrue())
		})

		By("And Then VpcPeering is ready", func() {
			Eventually(LoadAndCheck, "2s").
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed())
		})

		if feature.VpcPeeringSync.Value(infra.Ctx()) {
			By("And Then local route tables have no peering routes to removed remote VPC CIDR", func() {
				Expect(awsMockLocal.GetRoute(localVpcId, localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr2)).
					To(BeNil(), fmt.Sprintf("Local main route table should not have peering route to %s", remoteVpcCidr2))

				Expect(awsMockLocal.GetRoute(localVpcId, localRouteTable, kcpPeering.Status.Id, remoteVpcCidr2)).
					To(BeNil(), fmt.Sprintf("Local route table should not have peering route to %s", remoteVpcCidr2))

			})
		} else {
			By("And Then local route tables have peering routes to removed remote VPC CIDR", func() {
				Expect(awsMockLocal.GetRoute(localVpcId, localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr2)).
					NotTo(BeNil(), fmt.Sprintf("Local main route table should have peering route to %s", remoteVpcCidr2))

				Expect(awsMockLocal.GetRoute(localVpcId, localRouteTable, kcpPeering.Status.Id, remoteVpcCidr2)).
					NotTo(BeNil(), fmt.Sprintf("Local route table should have peering route to %s", remoteVpcCidr2))

			})
		}

		By("And Then local route tables have peering routes to added remote VPC CIDR", func() {
			Expect(awsMockLocal.GetRoute(localVpcId, localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr3)).
				NotTo(BeNil(), fmt.Sprintf("Local main route table should have peering route to %s", remoteVpcCidr3))

			Expect(awsMockLocal.GetRoute(localVpcId, localRouteTable, kcpPeering.Status.Id, remoteVpcCidr3)).
				NotTo(BeNil(), fmt.Sprintf("Local route table should have peering route to %s", remoteVpcCidr3))
		})

		// DELETE

		//
		By("When remote peering connection is deleted", func() {
			// The VPC peering connection remains visible to the party that deleted it for 2 hours
			Expect(awsMockRemote.DeleteVpcPeeringConnection(infra.Ctx(), &kcpPeering.Status.RemoteId)).To(Succeed())
			//  and visible to the other party for 2 days
			Expect(awsMockLocal.DeleteVpcPeeringConnection(infra.Ctx(), &kcpPeering.Status.Id)).To(Succeed())
		})

		By("When KCP VpcPeering is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "failed deleting VpcPeering")
		})

		By("Then VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "expected VpcPeering not to exist (be deleted), but it still exists")
		})

		By("And Then local VpcPeeringConnection is deleted", func() {
			localPeering, _ := awsMockLocal.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(localPeering).To(BeNil())
		})

		By("And Then local route tables have no peering routes to remote VPC CIDR", func() {
			Expect(awsMockLocal.GetRoute(localVpcId, localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				To(BeNil(), fmt.Sprintf("Local main route table should not have peering route to %s", remoteVpcCidr))

			Expect(awsMockLocal.GetRoute(localVpcId, localRouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				To(BeNil(), fmt.Sprintf("Local route table should not have peering route to %s", remoteVpcCidr))

			Expect(awsMockLocal.GetRoute(localVpcId, localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr3)).
				To(BeNil(), fmt.Sprintf("Local main route table should not have peering route to %s", remoteVpcCidr3))

			Expect(awsMockLocal.GetRoute(localVpcId, localRouteTable, kcpPeering.Status.Id, remoteVpcCidr3)).
				To(BeNil(), fmt.Sprintf("Local route table should not have peering route to %s", remoteVpcCidr3))
		})

		By("And Then remote VpcPeeringConnection is deleted", func() {
			remotePeering, _ := awsMockRemote.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(remotePeering).To(BeNil())
		})

		By("And Then remote route tables have no peering route to local VPC CIDR", func() {
			Expect(awsMockRemote.GetRoute(remoteVpcId, remoteRouteTableTagged, kcpPeering.Status.Id, localVpcCidr)).
				To(BeNil())
		})

	})
})
