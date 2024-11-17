package cloudcontrol

import (
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
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
			remoteVpcId          = "vpc-2c41e43fcd5340f8f"
			remoteVpcCidr        = "10.200.0.0/16"
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
			scopePkg.Ignore.AddName(kymaName)

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

		By("And Given AWS route table exists", func() {

			awsMockLocal.AddRouteTable(
				ptr.To(localMainRouteTable),
				ptr.To(localVpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", vpcName), "1"),
				[]ec2Types.RouteTableAssociation{
					{
						Main: ptr.To(true),
					},
				})

			awsMockLocal.AddRouteTable(
				ptr.To(wrong2RouteTable),
				ptr.To(wrong2VpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", wrong2VpcId), "1"),
				[]ec2Types.RouteTableAssociation{})

			awsMockLocal.AddRouteTable(
				ptr.To(localRouteTable),
				ptr.To(localVpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", vpcName), "1"),
				[]ec2Types.RouteTableAssociation{})
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

		By("And Given AWS remote route table exists", func() {

			awsMockRemote.AddRouteTable(
				ptr.To(remoteMainRouteTable),
				ptr.To(remoteVpcId),
				awsutil.Ec2Tags(),
				[]ec2Types.RouteTableAssociation{
					{
						Main: ptr.To(true),
					},
				})

			awsMockRemote.AddRouteTable(
				ptr.To(wrong3RouteTable),
				ptr.To(wrong3VpcId),
				awsutil.Ec2Tags(),
				[]ec2Types.RouteTableAssociation{})

			awsMockRemote.AddRouteTable(
				ptr.To(remoteRouteTable),
				ptr.To(remoteVpcId),
				awsutil.Ec2Tags(),
				[]ec2Types.RouteTableAssociation{})
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

		By("Then KCP VpcPeering is in missing local network error state", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering, NewObjActions(),
					HavingCondition(cloudcontrolv1beta1.ConditionTypeError, metav1.ConditionTrue, cloudcontrolv1beta1.ReasonMissingDependency, "Local network not found"),
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
			remoteKcpNet = cloudcontrolv1beta1.NewNetworkBuilder().
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

		By("Then KCP VpcPeering have status.id set", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HaveFinalizer(cloudcontrolv1beta1.FinalizerName),
					HavingState(string(cloudcontrolv1beta1.ErrorState)),
					HavingKcpVpcPeeringStatusIdNotEmpty(),
				).Should(Succeed())
		})

		By("When remote VpcPeeringConnection is initiated", func() {
			awsMockRemote.InitiateVpcPeeringConnection(infra.Ctx(), ptr.To(kcpPeering.Status.Id), ptr.To(remoteVpcId), ptr.To(localVpcId), ptr.To(scope.Spec.Region), ptr.To(scope.Spec.Scope.Aws.AccountId))
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

		By("When AWS VPC Peering state is active", func() {
			awsMockLocal.SetVpcPeeringConnectionActive(infra.Ctx(), ptr.To(localVpcId), ptr.To(remoteVpcId))
			awsMockRemote.SetVpcPeeringConnectionActive(infra.Ctx(), ptr.To(remoteVpcId), ptr.To(localVpcId))
		})

		By("Then KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then KCP VpcPeering has status.id equals to status.remoteId", func() {
			Expect(kcpPeering.Status.Id).To(Equal(kcpPeering.Status.RemoteId))
		})

		By("And Then KCP VpcPeering has status.vpcId equals to existing AWS VPC id", func() {
			Expect(kcpPeering.Status.VpcId).To(Equal(localVpcId))
		})

		By("And Then found local VpcPeeringConnection has AccepterVpcInfo.VpcId equals remote vpc id", func() {
			localPeering, _ := awsMockLocal.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(*localPeering.AccepterVpcInfo.VpcId).To(Equal(remoteVpcId))
		})

		By("And Then all local route tables has one new route with destination CIDR matching remote VPC CIDR", func() {
			Expect(awsMockLocal.GetRouteCount(infra.Ctx(), localVpcId, kcpPeering.Status.Id, remoteVpcCidr)).
				To(Equal(2))

			Expect(awsMockLocal.GetRoute(infra.Ctx(), localVpcId, localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				NotTo(BeNil(), fmt.Sprintf("Route table %s should have route with target %s and destination %s",
					localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr))

			Expect(awsMockLocal.GetRoute(infra.Ctx(), localVpcId, localRouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				ToNot(BeNil(), fmt.Sprintf("Route table %s should have route with target %s and destination %s",
					localRouteTable, kcpPeering.Status.Id, remoteVpcCidr))

			Expect(awsMockLocal.GetRoute(infra.Ctx(), wrong2VpcId, wrong2RouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				To(BeNil(), fmt.Sprintf("Route table %s should not have route with target %s and destination %s",
					wrong2RouteTable, kcpPeering.Status.Id, remoteVpcCidr))

		})

		By("And Then all remote route tables has one new route with destination CIDR matching VPC CIDR", func() {
			Expect(awsMockRemote.GetRouteCount(infra.Ctx(), remoteVpcId, kcpPeering.Status.RemoteId, localVpcCidr)).
				To(Equal(2))

			Expect(awsMockRemote.GetRoute(infra.Ctx(), remoteVpcId, remoteMainRouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				NotTo(BeNil(), fmt.Sprintf("Route table %s should have route with target %s and destination %s",
					localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr))

			Expect(awsMockRemote.GetRoute(infra.Ctx(), remoteVpcId, remoteRouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				ToNot(BeNil(), fmt.Sprintf("Route table %s should have route with target %s and destination %s",
					localRouteTable, kcpPeering.Status.Id, remoteVpcCidr))

			Expect(awsMockRemote.GetRoute(infra.Ctx(), wrong3VpcId, wrong3RouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
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

		By("And Then all local route tables has no routes with destination CIDR matching remote VPC CIDR", func() {
			Expect(awsMockLocal.GetRouteCount(infra.Ctx(), localVpcId, kcpPeering.Status.Id, remoteVpcCidr)).
				To(Equal(0))
		})

		By("And Then remote VpcPeeringConnection is deleted", func() {
			remotePeering, _ := awsMockRemote.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(remotePeering).To(BeNil())
		})

		By("And Then all remote route tables has no routes with destination CIDR matching local VPC CIDR", func() {
			Expect(awsMockRemote.GetRouteCount(infra.Ctx(), remoteVpcId, kcpPeering.Status.RemoteId, localVpcCidr)).
				To(Equal(0))
		})
	})

	// When prevent deletion of KCP Network while used by VpcPeering is implemented, this test case
	// is obsolete, but keeping it just in case, but with Network reconciler ignoring the created
	// networks so they can be deleted while used by VpcPeering
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
			scopePkg.Ignore.AddName(kymaName)

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
				[]ec2Types.RouteTableAssociation{
					{
						Main: ptr.To(true),
					},
				})

			awsMockLocal.AddRouteTable(
				ptr.To(localRouteTable),
				ptr.To(localVpcId),
				awsutil.Ec2Tags(fmt.Sprintf("kubernetes.io/cluster/%s", vpcName), "1"),
				[]ec2Types.RouteTableAssociation{})
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
				[]ec2Types.RouteTableAssociation{
					{
						Main: ptr.To(true),
					},
				})

			awsMockRemote.AddRouteTable(
				ptr.To(remoteRouteTable),
				ptr.To(remoteVpcId),
				awsutil.Ec2Tags(),
				[]ec2Types.RouteTableAssociation{})
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
			localKcpNet.Status.State = string(cloudcontrolv1beta1.ReadyState)
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
			remoteKcpNet.Status.State = string(cloudcontrolv1beta1.ReadyState)
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

		By("Then KCP VpcPeering have status.id set", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HaveFinalizer(cloudcontrolv1beta1.FinalizerName),
					HavingState(string(cloudcontrolv1beta1.ErrorState)),
					HavingKcpVpcPeeringStatusIdNotEmpty(),
				).Should(Succeed())
		})

		By("When remote VpcPeeringConnection is initiated", func() {
			awsMockRemote.InitiateVpcPeeringConnection(infra.Ctx(), ptr.To(kcpPeering.Status.Id), ptr.To(remoteVpcId), ptr.To(localVpcId), ptr.To(scope.Spec.Region), ptr.To(scope.Spec.Scope.Aws.AccountId))
			//TODO move step "When AWS VPC Peering state is active" here
			//TODO remove step "Then KCP VpcPeering is initiating-request"
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

		By("When AWS VPC Peering state is active", func() {
			awsMockLocal.SetVpcPeeringConnectionActive(infra.Ctx(), ptr.To(localVpcId), ptr.To(remoteVpcId))
			awsMockRemote.SetVpcPeeringConnectionActive(infra.Ctx(), ptr.To(remoteVpcId), ptr.To(localVpcId))
		})

		By("Then KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then KCP VpcPeering has status.id equals to status.remoteId", func() {
			Expect(kcpPeering.Status.Id).To(Equal(kcpPeering.Status.RemoteId))
		})

		By("And Then KCP VpcPeering has status.vpcId equals to existing AWS VPC id", func() {
			Expect(kcpPeering.Status.VpcId).To(Equal(localVpcId))
		})

		By("And Then found local VpcPeeringConnection has AccepterVpcInfo.VpcId equals remote vpc id", func() {
			localPeering, _ := awsMockLocal.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(*localPeering.AccepterVpcInfo.VpcId).To(Equal(remoteVpcId))
		})

		By("And Then all local route tables has one new route with destination CIDR matching remote VPC CIDR", func() {
			Expect(awsMockLocal.GetRouteCount(infra.Ctx(), localVpcId, kcpPeering.Status.Id, remoteVpcCidr)).
				To(Equal(2))

			Expect(awsMockLocal.GetRoute(infra.Ctx(), localVpcId, localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				NotTo(BeNil(), fmt.Sprintf("Route table %s should have route with target %s and destination %s",
					localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr))

			Expect(awsMockLocal.GetRoute(infra.Ctx(), localVpcId, localRouteTable, kcpPeering.Status.Id, remoteVpcCidr)).
				ToNot(BeNil(), fmt.Sprintf("Route table %s should have route with target %s and destination %s",
					localRouteTable, kcpPeering.Status.Id, remoteVpcCidr))
		})

		By("And Then all remote route tables has one new route with destination CIDR matching VPC CIDR", func() {
			Expect(awsMockRemote.GetRouteCount(infra.Ctx(), remoteVpcId, kcpPeering.Status.RemoteId, localVpcCidr)).
				To(Equal(2))

			Expect(awsMockRemote.GetRoute(infra.Ctx(), remoteVpcId, remoteMainRouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				NotTo(BeNil(), fmt.Sprintf("Route table %s should have route with target %s and destination %s",
					localMainRouteTable, kcpPeering.Status.Id, remoteVpcCidr))

			Expect(awsMockRemote.GetRoute(infra.Ctx(), remoteVpcId, remoteRouteTable, kcpPeering.Status.RemoteId, localVpcCidr)).
				ToNot(BeNil(), fmt.Sprintf("Route table %s should have route with target %s and destination %s",
					localRouteTable, kcpPeering.Status.Id, remoteVpcCidr))
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
			localPeering, _ := awsMockLocal.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(localPeering).To(BeNil())
		})

		By("And Then all local route tables has no routes with destination CIDR matching remote VPC CIDR", func() {
			Expect(awsMockLocal.GetRouteCount(infra.Ctx(), localVpcId, kcpPeering.Status.Id, remoteVpcCidr)).
				To(Equal(0))
		})

		// VpcPeeringConnection and Routes are not deleted since KCP remote Network is deleted previously
		By("And Then remote VpcPeeringConnection is not deleted", func() {
			remotePeering, _ := awsMockRemote.DescribeVpcPeeringConnection(infra.Ctx(), kcpPeering.Status.Id)
			Expect(remotePeering).NotTo(BeNil())
		})

		By("And Then all remote route tables has routes with destination CIDR matching local VPC CIDR", func() {
			Expect(awsMockRemote.GetRouteCount(infra.Ctx(), remoteVpcId, kcpPeering.Status.RemoteId, localVpcCidr)).
				To(Equal(2))
		})
	})
})
