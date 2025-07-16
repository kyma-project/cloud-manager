package cloudcontrol

import (
	"time"

	"github.com/kyma-project/cloud-manager/api"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	kcpgcpsubnet "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	kcpvpcpeering "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Feature: KCP Network reference", func() {

	It("Scenario: Network reference is created and deleted", func() {

		kymaName := "0b10bb9c-a727-4f5b-8d64-803110402586"
		scope := &cloudcontrolv1beta1.Scope{}
		netObjName := "bacd3153-ecd4-4666-86de-067f94aaae53"
		var net *cloudcontrolv1beta1.Network

		awsAccount := "acc-123"
		awsRegion := "us-east-1"
		netId := "net-345"
		netName := "my-net"

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(kymaName)
			Expect(CreateScopeAws(infra.Ctx(), infra, scope, WithName(kymaName))).
				To(Succeed())
		})

		By("When Network reference is created", func() {
			net = cloudcontrolv1beta1.NewNetworkBuilder().WithAwsRef(awsAccount, awsRegion, netId, netName).Build()
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), net, WithName(netObjName), WithScope(kymaName))).
				To(Succeed())
		})

		By("Then Network state is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), net, NewObjActions(), HavingState(string(cloudcontrolv1beta1.StateReady))).
				Should(Succeed())
		})

		By("And Then Network has Ready condition", func() {
			cond := meta.FindStatusCondition(*net.Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
		})

		By("And Then Network reference in status equals to its spec value", func() {
			Expect(net.Status.Network).To(Not(BeNil()))
			Expect(net.Spec.Network.Reference.Equals(net.Status.Network)).To(BeTrue())
		})

		By("And Then Network has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(net, api.CommonFinalizerDeletionHook)).To(BeTrue())
		})

		By("When Network is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), net)).
				To(Succeed())
		})

		By("Then Network does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), net).
				Should(Succeed())
		})

		By("// cleanup: delete Scope", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), scope)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})
	})

	It("Scenario: Network reference can not be deleted when used by IpRange", func() {
		kymaName := "19cf354a-aa43-4e53-aad7-23b2428e2eb4"
		scope := &cloudcontrolv1beta1.Scope{}
		netObjName := "47619a11-7259-43fe-ba26-94c5e6260f02"
		var net *cloudcontrolv1beta1.Network
		ipRangeName := "f86295fc-b2af-4e14-adb5-597fbe03ae04"

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(kymaName)
			Expect(CreateScopeAws(infra.Ctx(), infra, scope, WithName(kymaName))).
				To(Succeed())
		})

		By("And Given Network reference is created", func() {
			net = cloudcontrolv1beta1.NewNetworkBuilder().WithAwsRef("acc-876", "us-east-1", "net-987", "my-net").Build()
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), net, WithName(netObjName), WithScope(kymaName))).
				To(Succeed())
		})

		By("And Given Network state is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), net, NewObjActions(), HavingState(string(cloudcontrolv1beta1.StateReady))).
				Should(Succeed())
		})

		ipRange := &cloudcontrolv1beta1.IpRange{}

		By("And Given IpRange using Network is created", func() {
			kcpiprange.Ignore.AddName(ipRangeName)
			Expect(CreateKcpIpRange(infra.Ctx(), infra.KCP().Client(), ipRange,
				WithName(ipRangeName),
				WithScope(kymaName),
				WithRemoteRef("foo"),
				WithKcpIpRangeNetwork(netObjName),
				WithKcpIpRangeSpecCidr("10.181.0.0/16"),
			)).To(Succeed())
			// give it time to be watched and indexed
			time.Sleep(500 * time.Millisecond)
		})

		By("When Network is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), net)).
				To(Succeed())
		})

		By("Then Network has Warning state", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), net,
					NewObjActions(),
					HavingState(string(cloudcontrolv1beta1.StateWarning)),
				).
				Should(Succeed())
		})

		By("And Then Network has DeleteWhileUsed Warning condition", func() {
			cond := meta.FindStatusCondition(net.Status.Conditions, cloudcontrolv1beta1.ConditionTypeWarning)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(cloudcontrolv1beta1.ReasonDeleteWhileUsed))
		})

		By("When IpRange is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), ipRange)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), ipRange).
				Should(Succeed())
		})

		By("Then Network does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), net).
				Should(Succeed())
		})

		By("// cleanup: delete Scope", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), scope)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})
	})

	It("Scenario: Network reference can not be deleted when used by GcpSubnet", func() {
		kymaName := "75840abb-44b3-4430-9ffb-e1806167ec38"
		scope := &cloudcontrolv1beta1.Scope{}
		netObjName := "b7e37580-5524-4fad-8380-5fa686c99f11"
		var net *cloudcontrolv1beta1.Network
		gcpSubnetName := "8eab2147-086e-42d4-9b36-fb9dfa8b9364"

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(kymaName)
			Expect(CreateScopeAws(infra.Ctx(), infra, scope, WithName(kymaName))).
				To(Succeed())
		})

		By("And Given Network reference is created", func() {
			net = cloudcontrolv1beta1.NewNetworkBuilder().WithGcpRef("proj-f-876", "my-net-94").Build()
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), net, WithName(netObjName), WithScope(kymaName))).
				To(Succeed())
		})

		By("And Given Network state is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), net, NewObjActions(), HavingState(string(cloudcontrolv1beta1.StateReady))).
				Should(Succeed())
		})

		gcpSubnet := &cloudcontrolv1beta1.GcpSubnet{}

		By("And Given GcpSubnet using Network is created", func() {
			kcpgcpsubnet.Ignore.AddName(gcpSubnetName)
			Expect(CreateKcpGcpSubnet(infra.Ctx(), infra.KCP().Client(), gcpSubnet,
				WithName(gcpSubnetName),
				WithScope(kymaName),
				WithRemoteRef("foo"),
				WithKcpGcpSubnetNetwork(netObjName),
				WithKcpGcpSubnetSpecCidr("10.204.0.0/16"),
				WithKcpGcpSubnetPurposePrivate(),
			)).To(Succeed())
			// give it time to be watched and indexed
			time.Sleep(500 * time.Millisecond)
		})

		By("When Network is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), net)).
				To(Succeed())
		})

		By("Then Network has Warning state", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), net,
					NewObjActions(),
					HavingState(string(cloudcontrolv1beta1.StateWarning)),
				).
				Should(Succeed())
		})

		By("And Then Network has DeleteWhileUsed Warning condition", func() {
			cond := meta.FindStatusCondition(net.Status.Conditions, cloudcontrolv1beta1.ConditionTypeWarning)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(cloudcontrolv1beta1.ReasonDeleteWhileUsed))
		})

		By("When GcpSubnet is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), gcpSubnet)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gcpSubnet).
				Should(Succeed())
		})

		By("Then Network does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), net).
				Should(Succeed())
		})

		By("// cleanup: delete Scope", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), scope)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})
	})

	It("Scenario: Network reference can not be deleted when used by VpcPeering local network", func() {
		kymaName := "a4797ac1-25b2-4853-97ae-72dbc8e10828"
		scope := &cloudcontrolv1beta1.Scope{}
		localNetworkName := "ef40f1bf-c8b4-47bb-ba83-ddba69abdefc"
		remoteNetworkName := "c9a49d3c-c64e-4b83-8e0d-72916538c6f6"
		var localNet *cloudcontrolv1beta1.Network
		var remoteNet *cloudcontrolv1beta1.Network
		vpcPeeringName := "6f0747be-058d-4957-89e7-6a67a219f089"

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(kymaName)
			Expect(CreateScopeAws(infra.Ctx(), infra, scope, WithName(kymaName))).
				To(Succeed())
		})

		By("And Given local Network reference is created", func() {
			localNet = cloudcontrolv1beta1.NewNetworkBuilder().WithAwsRef("acc-876", "us-east-1", "net-987", "my-net").Build()
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), localNet, WithName(localNetworkName), WithScope(kymaName))).
				To(Succeed())
		})

		By("And Given local Network state is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localNet, NewObjActions(), HavingState(string(cloudcontrolv1beta1.StateReady))).
				Should(Succeed())
		})

		By("And Given remote Network reference is created", func() {
			remoteNet = cloudcontrolv1beta1.NewNetworkBuilder().WithAwsRef("acc-543", "us-east-1", "net-876", "his-net").Build()
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), remoteNet, WithName(remoteNetworkName), WithScope(kymaName))).
				To(Succeed())
		})

		By("And Given remote Network state is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteNet, NewObjActions(), HavingState(string(cloudcontrolv1beta1.StateReady))).
				Should(Succeed())
		})

		var vpcPeering *cloudcontrolv1beta1.VpcPeering

		By("And Given VpcPeering using local and remote Network is created", func() {
			kcpvpcpeering.Ignore.AddName(vpcPeeringName)
			vpcPeering = cloudcontrolv1beta1.NewVpcPeeringBuilder().
				WithName(vpcPeeringName).
				WithScope(kymaName).
				WithRemoteRef(DefaultSkrNamespace, "foo").
				WithDetails(localNet.Name, localNet.Namespace, remoteNet.Name, remoteNet.Namespace, "peeringName", false, false).
				Build()
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), vpcPeering)).
				To(Succeed())
			// give it time to be watched and indexed
			time.Sleep(500 * time.Millisecond)
		})

		By("When local Network is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), localNet)).
				To(Succeed())
		})

		By("Then local Network has Warning state", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localNet,
					NewObjActions(),
					HavingState(string(cloudcontrolv1beta1.StateWarning)),
				).
				Should(Succeed())
		})

		By("And Then local Network has DeleteWhileUsed Warning condition", func() {
			cond := meta.FindStatusCondition(localNet.Status.Conditions, cloudcontrolv1beta1.ConditionTypeWarning)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(cloudcontrolv1beta1.ReasonDeleteWhileUsed))
		})

		By("When VpcPeering is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), vpcPeering)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcPeering).
				Should(Succeed())
		})

		By("Then local Network does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localNet).
				Should(Succeed())
		})

		By("// cleanup: delete remote Network", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), remoteNet)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteNet).
				Should(Succeed())
		})

		By("// cleanup: delete Scope", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), scope)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})
	})

	It("Scenario: Network reference can not be deleted when used by VpcPeering remote network", func() {
		kymaName := "0d79b58f-16e8-401c-b363-10f4d81d36e9"
		scope := &cloudcontrolv1beta1.Scope{}
		localNetworkName := "2c9ae7bc-73f6-494e-ae8b-f9a8e3e733ec"
		remoteNetworkName := "647dda6d-504f-4812-ae9e-d072fccb1ea4"
		var localNet *cloudcontrolv1beta1.Network
		var remoteNet *cloudcontrolv1beta1.Network
		vpcPeeringName := "8b4190ce-e401-4bb3-8a0e-608b6cf7da59"

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(kymaName)
			Expect(CreateScopeAws(infra.Ctx(), infra, scope, WithName(kymaName))).
				To(Succeed())
		})

		By("And Given local Network reference is created", func() {
			localNet = cloudcontrolv1beta1.NewNetworkBuilder().WithAwsRef("acc-876", "us-east-1", "net-987", "my-net").Build()
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), localNet, WithName(localNetworkName), WithScope(kymaName))).
				To(Succeed())
		})

		By("And Given local Network state is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localNet, NewObjActions(), HavingState(string(cloudcontrolv1beta1.StateReady))).
				Should(Succeed())
		})

		By("And Given remote Network reference is created", func() {
			remoteNet = cloudcontrolv1beta1.NewNetworkBuilder().WithAwsRef("acc-543", "us-east-1", "net-876", "his-net").Build()
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), remoteNet, WithName(remoteNetworkName), WithScope(kymaName))).
				To(Succeed())
		})

		By("And Given remote Network state is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteNet, NewObjActions(), HavingState(string(cloudcontrolv1beta1.StateReady))).
				Should(Succeed())
		})

		var vpcPeering *cloudcontrolv1beta1.VpcPeering

		By("And Given VpcPeering using local and remote Network is created", func() {
			kcpvpcpeering.Ignore.AddName(vpcPeeringName)
			vpcPeering = cloudcontrolv1beta1.NewVpcPeeringBuilder().
				WithName(vpcPeeringName).
				WithScope(kymaName).
				WithRemoteRef(DefaultSkrNamespace, "foo").
				WithDetails(localNet.Name, localNet.Namespace, remoteNet.Name, remoteNet.Namespace, "peeringName", false, false).
				Build()
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), vpcPeering)).
				To(Succeed())
			// give it time to be watched and indexed
			time.Sleep(500 * time.Millisecond)
		})

		By("When remote Network is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), remoteNet)).
				To(Succeed())
		})

		By("Then remote Network has Warning state", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteNet,
					NewObjActions(),
					HavingState(string(cloudcontrolv1beta1.StateWarning)),
				).
				Should(Succeed())
		})

		By("And Then remote Network has DeleteWhileUsed Warning condition", func() {
			cond := meta.FindStatusCondition(remoteNet.Status.Conditions, cloudcontrolv1beta1.ConditionTypeWarning)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(cloudcontrolv1beta1.ReasonDeleteWhileUsed))
		})

		By("When VpcPeering is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), vpcPeering)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcPeering).
				Should(Succeed())
		})

		By("Then remote Network does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteNet).
				Should(Succeed())
		})

		By("// cleanup: delete local Network", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), localNet)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localNet).
				Should(Succeed())
		})

		By("// cleanup: delete Scope", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), scope)).
				To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})
	})

	It("Scenario: kyma Network reference will delete itself if its Scope does not exist", func() {
		// doesn't exist
		kymaName := "eb3aac05-dac0-4e99-924c-dbb717935b6d"
		netObjName := common.KcpNetworkKymaCommonName(kymaName)
		var net *cloudcontrolv1beta1.Network

		awsAccount := "acc-123123"
		awsRegion := "us-east-1"
		netId := "net-345123"
		netName := "my-net"

		By("Given Scope does not exist", func() {})

		By("When kyma Network reference is created", func() {
			net = cloudcontrolv1beta1.NewNetworkBuilder().
				WithType(cloudcontrolv1beta1.NetworkTypeKyma).
				WithAwsRef(awsAccount, awsRegion, netId, netName).
				Build()
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), net, WithName(netObjName), WithScope(kymaName))).
				To(Succeed())
		})

		By("Then kyma Network reference does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), net).
				Should(Succeed())
		})
	})
})
