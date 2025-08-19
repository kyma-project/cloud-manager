package cloudcontrol

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	kcpnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/network"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Feature: KCP VpcPeering", func() {

	It("Scenario: KCP Azure VpcPeering is created and deleted", func() {
		const (
			kymaName            = "6a62936d-aa6e-4d5b-aaaa-5eae646d1bd5"
			kcpPeeringName      = "281bc581-8635-4d56-ba52-fa48ec6f7c69"
			remoteSubscription  = "afdbc79f-de19-4df4-94cd-6be2739dc0e0"
			remoteResourceGroup = "MyResourceGroup"
			remoteVnetName      = "MyVnet"
			remotePeeringName   = "my-peering"
			localPeeringName    = "kyma-peering"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAzure).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		localResourceGroupName := scope.Spec.Scope.Azure.VpcNetwork
		localVirtualNetworkName := scope.Spec.Scope.Azure.VpcNetwork

		azureMockLocal := infra.AzureMock().MockConfigs(scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.TenantId)
		azureMockRemote := infra.AzureMock().MockConfigs(remoteSubscription, scope.Spec.Scope.Azure.TenantId)

		By("And Given local Azure VNET exists", func() {
			err := azureMockLocal.CreateNetwork(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, scope.Spec.Region, "10.200.0.0/25", nil)
			Expect(err).ToNot(HaveOccurred())
		})

		By("And Given remote Azure VNet exists with Kyma tag", func() {
			err := azureMockRemote.CreateNetwork(infra.Ctx(), remoteResourceGroup, remoteVnetName, scope.Spec.Region, "10.100.0.0/25", map[string]string{kymaName: kymaName})
			Expect(err).ToNot(HaveOccurred())
		})

		localKcpNetworkName := common.KcpNetworkKymaCommonName(scope.Name)
		remoteKcpNetworkName := scope.Name + "--remote"

		var kcpPeering *cloudcontrolv1beta1.VpcPeering

		By("When KCP VpcPeering is created", func() {
			kcpPeering = (&cloudcontrolv1beta1.VpcPeeringBuilder{}).
				WithScope(kymaName).
				WithRemoteRef("skr-namespace", "skr-azure-vpcpeering").
				WithDetails(localKcpNetworkName, infra.KCP().Namespace(), remoteKcpNetworkName, infra.KCP().Namespace(), remotePeeringName, true, true).
				WithLocalPeeringName(localPeeringName).
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					WithName(kcpPeeringName),
				).
				Should(Succeed())
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
				WithAzureRef(scope.Spec.Scope.Azure.TenantId, scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.VpcNetwork, scope.Spec.Scope.Azure.VpcNetwork).
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
			_, err := composed.PatchObjMergeAnnotation(infra.Ctx(), "test", "1", localKcpNet, infra.KCP().Client())
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
				WithAzureRef(scope.Spec.Scope.Azure.TenantId, remoteSubscription, remoteResourceGroup, remoteVnetName).
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
			_, err := composed.PatchObjMergeAnnotation(infra.Ctx(), "test", "1", remoteKcpNet, infra.KCP().Client())
			Expect(err).
				To(Succeed())
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet, NewObjActions(), HavingState("Ready")).
				Should(Succeed(), "expected remote kcp network to become ready but it didn't")
		})

		// Peering Created ===============================================================

		var localPeering *armnetwork.VirtualNetworkPeering

		By("Then local Azure VPC Peering is created", func() {
			remoteVnetId := util.NewVirtualNetworkResourceId(remoteSubscription, remoteResourceGroup, remoteVnetName).String()

			Eventually(func() error {
				p, err := azureMockLocal.GetPeering(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, localPeeringName)
				localPeering = p
				return err
			}).Should(Succeed())

			Expect(localPeering).NotTo(BeNil())
			Expect(ptr.Deref(localPeering.Properties.RemoteVirtualNetwork.ID, "")).To(Equal(remoteVnetId))
			Expect(ptr.Deref(localPeering.Properties.UseRemoteGateways, false)).To(BeFalse())

		})

		var remotePeering *armnetwork.VirtualNetworkPeering

		By("And Then remote Azure VPC Peering is created", func() {
			localVnetId := util.NewVirtualNetworkResourceId(scope.Spec.Scope.Azure.SubscriptionId, localResourceGroupName, localVirtualNetworkName).String()

			Eventually(func() error {
				p, err := azureMockRemote.GetPeering(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName)
				remotePeering = p
				return err

			}).Should(Succeed())

			Expect(remotePeering).NotTo(BeNil())
			Expect(ptr.Deref(remotePeering.Properties.RemoteVirtualNetwork.ID, "")).To(Equal(localVnetId))
			Expect(ptr.Deref(remotePeering.Properties.UseRemoteGateways, false)).To(BeFalse())
		})

		// Ready ==========================================================

		By("And Then local Azure VPC peering state is connected", func() {
			Eventually(azureMockLocal.SetPeeringConnectedFullInSync).
				WithArguments(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, localPeeringName).
				Should(Succeed())
		})

		By("And remote Azure VPC peering state is connected", func() {
			Eventually(azureMockRemote.SetPeeringConnectedFullInSync).
				WithArguments(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName).
				Should(Succeed())
		})

		By("Then KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then KCP VpcPeering has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(kcpPeering, api.CommonFinalizerDeletionHook)).
				To(BeTrue())
		})

		By("Then KCP VpcPeering state is Connected", func() {
			Expect(kcpPeering.Status.State).To(Equal(cloudcontrolv1beta1.VirtualNetworkPeeringStateConnected))
		})

		By("And Then KCP VpcPeering Id matches local Azure peering ID", func() {
			Expect(ptr.Deref(localPeering.ID, "")).To(Equal(kcpPeering.Status.Id))
		})

		By("And Then KCP VpcPeering RemoteId matches remote Azure peering ID", func() {
			Expect(ptr.Deref(remotePeering.ID, "")).To(Equal(kcpPeering.Status.RemoteId))
		})

		// DELETE

		By("When KCP VpcPeering is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "failed deleting VpcPeering")
		})

		By("Then KCP VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "expected VpcPeering not to exist (be deleted), but it still exists")
		})

		By("And Then local Azure peering does not exist", func() {
			peering, err := azureMockLocal.GetPeering(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, localPeeringName)
			Expect(err).To(HaveOccurred())
			Expect(azuremeta.IsNotFound(err)).To(BeTrue())
			Expect(peering).To(BeNil())
		})

		By("And Then remote Azure peering does not exists", func() {
			peering, err := azureMockRemote.GetPeering(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName)
			Expect(err).To(HaveOccurred())
			Expect(azuremeta.IsNotFound(err)).To(BeTrue())
			Expect(peering).To(BeNil())
		})

		By("// cleanup: Local KCP Network", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet).
				Should(Succeed())
		})

		By("// cleanup: Remote KCP Network", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet).
				Should(Succeed())
		})

		By("// cleanup: Scope", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})

	})

	// When prevent deletion of KCP Network while used by VpcPeering is implemented, this case
	// is obsolete, but keeping it just in case, with Network reconciler ignoring the created
	// networks, so they can be deleted while used by VpcPeering
	It("Scenario: KCP Azure VpcPeering is deleted when local and remote networks are missing", func() {
		const (
			kymaName            = "b64eab45-35b0-4015-b0f0-c819b351a6cd"
			kcpPeeringName      = "b9a927d6-08ff-467a-b2ec-047eeceb4a26"
			remoteSubscription  = "5bbd0ab5-51cd-440e-96a7-5db29dfef5e4"
			remoteResourceGroup = "MyResourceGroup"
			remoteVnetName      = "MyVnet"
			remotePeeringName   = "my-peering"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAzure).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		localResourceGroupName := scope.Spec.Scope.Azure.VpcNetwork
		localVirtualNetworkName := scope.Spec.Scope.Azure.VpcNetwork

		azureMockLocal := infra.AzureMock().MockConfigs(scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.TenantId)
		azureMockRemote := infra.AzureMock().MockConfigs(remoteSubscription, scope.Spec.Scope.Azure.TenantId)

		By("And Given local Azure VNET exists", func() {
			err := azureMockLocal.CreateNetwork(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, scope.Spec.Region, "10.200.0.0/25", nil)
			Expect(err).ToNot(HaveOccurred())
		})

		By("And Given remote Azure VNet exists with Kyma tag", func() {
			err := azureMockRemote.CreateNetwork(infra.Ctx(), remoteResourceGroup, remoteVnetName, scope.Spec.Region, "10.100.0.0/25", map[string]string{kymaName: kymaName})
			Expect(err).ToNot(HaveOccurred())
		})

		localKcpNetworkName := common.KcpNetworkKymaCommonName(scope.Name)
		remoteKcpNetworkName := scope.Name + "--remote"

		var localKcpNet *cloudcontrolv1beta1.Network

		By("And Given local KCP Network exists", func() {
			// must tell reconciler to ignore it, since it would prevent deletion when used by peering
			kcpnetwork.Ignore.AddName(localKcpNetworkName)
			localKcpNet = (&cloudcontrolv1beta1.NetworkBuilder{}).
				WithScope(scope.Name).
				WithAzureRef(scope.Spec.Scope.Azure.TenantId, scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.VpcNetwork, scope.Spec.Scope.Azure.VpcNetwork).
				Build()
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), localKcpNet, WithName(localKcpNetworkName))).
				Should(Succeed())

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
			remoteKcpNet = (&cloudcontrolv1beta1.NetworkBuilder{}).
				WithScope(scope.Name).
				WithAzureRef(scope.Spec.Scope.Azure.TenantId, remoteSubscription, remoteResourceGroup, remoteVnetName).
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
				WithRemoteRef("skr-namespace", "skr-azure-vpcpeering").
				WithDetails(localKcpNetworkName, infra.KCP().Namespace(), remoteKcpNetworkName, infra.KCP().Namespace(), remotePeeringName, true, true).
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					WithName(kcpPeeringName),
				).
				Should(Succeed())
		})

		By("And local Azure VPC peering state is connected", func() {
			Eventually(azureMockLocal.SetPeeringConnectedFullInSync).
				// When Spec.Details.LocalPeeringName is not specified localPeeringName defaults to metadata.Name
				WithArguments(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, kcpPeeringName).
				Should(Succeed())
		})

		By("And remote Azure VPC peering state is connected", func() {
			Eventually(azureMockRemote.SetPeeringConnectedFullInSync).
				WithArguments(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName).
				Should(Succeed())
		})

		// Ready ==========================================================

		By("Then KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("Then KCP VpcPeering state is Connected", func() {
			Expect(kcpPeering.Status.State).To(Equal(cloudcontrolv1beta1.VirtualNetworkPeeringStateConnected))
		})

		// Delete KCP local Network to check whether reconciler will be able to delete VpcPeering
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

		// Delete KCP remote Network to check whether reconciler will be able to delete VpcPeering
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

		By("Then KCP VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "expected VpcPeering not to exist (be deleted), but it still exists")
		})

		By("// cleanup: Scope", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})
	})

	It("Scenario: KCP Azure VpcPeering can be deleted when unauthorized", func() {
		const (
			kymaName            = "c86d383c-be2d-47c7-893e-ac93659760be"
			kcpPeeringName      = "187c9650-07b7-4996-9cb7-c96d195d9545"
			remoteSubscription  = "0a2b6080-f5c8-48ce-bcba-67bfacc3290a"
			remoteResourceGroup = "MyResourceGroup"
			remoteVnetName      = "MyVnet"
			remotePeeringName   = "my-peering"
			localPeeringName    = "kyma-peering"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAzure).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		localResourceGroupName := scope.Spec.Scope.Azure.VpcNetwork
		localVirtualNetworkName := scope.Spec.Scope.Azure.VpcNetwork

		azureMockLocal := infra.AzureMock().MockConfigs(scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.TenantId)
		azureMockRemote := infra.AzureMock().MockConfigs(remoteSubscription, scope.Spec.Scope.Azure.TenantId)

		By("And Given local Azure VNET exists", func() {
			err := azureMockLocal.CreateNetwork(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, scope.Spec.Region, "10.200.0.0/25", nil)
			Expect(err).ToNot(HaveOccurred())
		})

		By("And Given remote Azure VNet exists with Kyma tag", func() {
			err := azureMockRemote.CreateNetwork(infra.Ctx(), remoteResourceGroup, remoteVnetName, scope.Spec.Region, "10.100.0.0/25", map[string]string{kymaName: kymaName})
			Expect(err).ToNot(HaveOccurred())
		})

		localKcpNetworkName := common.KcpNetworkKymaCommonName(scope.Name)
		remoteKcpNetworkName := scope.Name + "--remote"

		var localKcpNet *cloudcontrolv1beta1.Network

		By("And Given local KCP Network is created", func() {
			localKcpNet = (&cloudcontrolv1beta1.NetworkBuilder{}).
				WithScope(scope.Name).
				WithAzureRef(scope.Spec.Scope.Azure.TenantId, scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.VpcNetwork, scope.Spec.Scope.Azure.VpcNetwork).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet, WithName(localKcpNetworkName)).
				Should(Succeed())
		})

		var remoteKcpNet *cloudcontrolv1beta1.Network

		By("And Given remote KCP Network is created", func() {
			remoteKcpNet = (&cloudcontrolv1beta1.NetworkBuilder{}).
				WithScope(scope.Name).
				WithAzureRef(scope.Spec.Scope.Azure.TenantId, remoteSubscription, remoteResourceGroup, remoteVnetName).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet, WithName(remoteKcpNetworkName)).
				Should(Succeed())
		})

		azureMockRemote.SetPeeringError(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName, azuremeta.NewAzureAuthorizationFailedError())

		var kcpPeering *cloudcontrolv1beta1.VpcPeering

		By("And Given KCP VpcPeering is created", func() {
			kcpPeering = (&cloudcontrolv1beta1.VpcPeeringBuilder{}).
				WithScope(kymaName).
				WithRemoteRef("skr-namespace", "skr-azure-vpcpeering").
				WithDetails(localKcpNetworkName, infra.KCP().Namespace(), remoteKcpNetworkName, infra.KCP().Namespace(), remotePeeringName, true, true).
				WithLocalPeeringName(localPeeringName).
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					WithName(kcpPeeringName),
				).
				Should(Succeed())
		})

		// Ready ==========================================================

		By("And Given KCP VpcPeering has Error condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HavingCondition(cloudcontrolv1beta1.ConditionTypeError,
						metav1.ConditionTrue,
						cloudcontrolv1beta1.ReasonUnauthorized,
						"Not authorized to perform action."),
				).
				Should(Succeed())
		})

		// DELETE

		By("When KCP VpcPeering is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "failed deleting VpcPeering")
		})

		By("Then KCP VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "expected VpcPeering not to exist (be deleted), but it still exists")
		})

		By("And Then local Azure peering does not exist", func() {
			peering, err := azureMockLocal.GetPeering(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, localPeeringName)
			Expect(err).To(HaveOccurred())
			Expect(azuremeta.IsNotFound(err)).To(BeTrue())
			Expect(peering).To(BeNil())
		})

		By("// cleanup: Local KCP Network", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet).
				Should(Succeed())
		})

		By("// cleanup: Remote KCP Network", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet).
				Should(Succeed())
		})

		By("// cleanup: Scope", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})
	})

	It("Scenario: KCP Azure cross-tenant VpcPeering can be created and deleted", func() {
		const (
			kymaName            = "315df1ab-2de8-44ec-81ca-b3967576bb6c"
			kcpPeeringName      = "b61238ec-5d45-4194-b947-fb1d851ac375"
			remoteSubscription  = "2d78e409-815e-44df-ab3c-326f553358b3"
			remoteTenant        = "66afe40d-f730-4556-9713-cdb3cf21a6df"
			remoteResourceGroup = "MyResourceGroup"
			remoteVnetName      = "MyVnet"
			remotePeeringName   = "my-peering"
			localPeeringName    = "kyma-peering"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAzure).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		localResourceGroupName := scope.Spec.Scope.Azure.VpcNetwork
		localVirtualNetworkName := scope.Spec.Scope.Azure.VpcNetwork

		azureMockLocal := infra.AzureMock().MockConfigs(scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.TenantId)
		azureMockRemote := infra.AzureMock().MockConfigs(remoteSubscription, remoteTenant)

		By("And Given local Azure VNET exists", func() {
			err := azureMockLocal.CreateNetwork(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, scope.Spec.Region, "10.200.0.0/25", nil)
			Expect(err).ToNot(HaveOccurred())
		})

		By("And Given remote Azure VNet exists with Kyma tag", func() {
			err := azureMockRemote.CreateNetwork(infra.Ctx(), remoteResourceGroup, remoteVnetName, scope.Spec.Region, "10.100.0.0/25", map[string]string{kymaName: kymaName})
			Expect(err).ToNot(HaveOccurred())
		})

		localKcpNetworkName := common.KcpNetworkKymaCommonName(scope.Name)
		remoteKcpNetworkName := scope.Name + "--remote"

		var localKcpNet *cloudcontrolv1beta1.Network

		By("And Given local KCP Network is created", func() {
			localKcpNet = (&cloudcontrolv1beta1.NetworkBuilder{}).
				WithScope(scope.Name).
				WithAzureRef(scope.Spec.Scope.Azure.TenantId, scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.VpcNetwork, scope.Spec.Scope.Azure.VpcNetwork).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet, WithName(localKcpNetworkName)).
				Should(Succeed())
		})

		var remoteKcpNet *cloudcontrolv1beta1.Network

		By("And Given remote KCP Network is created", func() {
			remoteKcpNet = (&cloudcontrolv1beta1.NetworkBuilder{}).
				WithScope(scope.Name).
				WithAzureRef(remoteTenant, remoteSubscription, remoteResourceGroup, remoteVnetName).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet, WithName(remoteKcpNetworkName)).
				Should(Succeed())
		})

		var kcpPeering *cloudcontrolv1beta1.VpcPeering

		By("And Given KCP VpcPeering is created", func() {
			kcpPeering = (&cloudcontrolv1beta1.VpcPeeringBuilder{}).
				WithScope(kymaName).
				WithRemoteRef("skr-namespace", "skr-azure-vpcpeering").
				WithDetails(localKcpNetworkName, infra.KCP().Namespace(), remoteKcpNetworkName, infra.KCP().Namespace(), remotePeeringName, true, true).
				WithLocalPeeringName(localPeeringName).
				WithUseRemoteGateway(true).
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					WithName(kcpPeeringName),
				).
				Should(Succeed())
		})

		By("And Given local Azure VPC peering state is connected", func() {
			Eventually(azureMockLocal.SetPeeringConnectedFullInSync).
				WithArguments(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, localPeeringName).
				Should(Succeed())
		})

		By("And Given remote Azure VPC peering state is connected", func() {
			Eventually(azureMockRemote.SetPeeringConnectedFullInSync).
				WithArguments(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName).
				Should(Succeed())
		})

		// Ready ==========================================================

		By("And Given KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then local Azure Peering has UseRemoteGateways equals to true", func() {
			localAzurePeering, err := azureMockLocal.GetPeering(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, localPeeringName)
			Expect(err).ToNot(HaveOccurred())
			Expect(ptr.Deref(localAzurePeering.Properties.UseRemoteGateways, false)).To(BeTrue())
			Expect(ptr.Deref(localAzurePeering.Properties.AllowGatewayTransit, false)).To(BeFalse())
		})

		By("And Then remote Azure Peering has UseRemoteGateways equals to false", func() {
			remoteAzurePeering, err := azureMockRemote.GetPeering(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName)
			Expect(err).ToNot(HaveOccurred())
			Expect(ptr.Deref(remoteAzurePeering.Properties.UseRemoteGateways, false)).To(BeFalse())
			Expect(ptr.Deref(remoteAzurePeering.Properties.AllowGatewayTransit, false)).To(BeTrue())

		})

		// DELETE

		By("When KCP VpcPeering is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "failed deleting VpcPeering")
		})

		By("Then KCP VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "expected VpcPeering not to exist (be deleted), but it still exists")
		})

		By("And Then local Azure peering does not exist", func() {
			peering, err := azureMockLocal.GetPeering(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, localPeeringName)
			Expect(err).To(HaveOccurred())
			Expect(azuremeta.IsNotFound(err)).To(BeTrue())
			Expect(peering).To(BeNil())
		})

		By("And Then remote Azure peering does not exists", func() {
			peering, err := azureMockRemote.GetPeering(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName)
			Expect(err).To(HaveOccurred())
			Expect(azuremeta.IsNotFound(err)).To(BeTrue())
			Expect(peering).To(BeNil())
		})

		By("// cleanup: Local KCP Network", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet).
				Should(Succeed())
		})

		By("// cleanup: Remote KCP Network", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet).
				Should(Succeed())
		})

		By("// cleanup: Scope", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})
	})

	It("Scenario: KCP Azure VpcPeering can sync", Focus, func() {
		const (
			kymaName            = "be27ef0c-013a-4011-96b8-86381ff64ab9"
			kcpPeeringName      = "f35f52ef-269d-49d6-8d6c-332e96d2293e"
			remoteSubscription  = "1cf7eb86-0a97-48da-bc94-d0e0d32ebd57"
			remoteResourceGroup = "MyResourceGroup"
			remoteVnetName      = "MyVnet"
			remotePeeringName   = "my-peering"
			localPeeringName    = "kyma-peering"
		)

		if !feature.VpcPeeringSync.Value(infra.Ctx()) {
			Skip("VPC peering sync for Azure is disabled")
		}

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAzure).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		localResourceGroupName := scope.Spec.Scope.Azure.VpcNetwork
		localVirtualNetworkName := scope.Spec.Scope.Azure.VpcNetwork

		azureMockLocal := infra.AzureMock().MockConfigs(scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.TenantId)
		azureMockRemote := infra.AzureMock().MockConfigs(remoteSubscription, scope.Spec.Scope.Azure.TenantId)

		azureMockRemote.AddRemoteSubscription(infra.Ctx(), &azureMockLocal)
		azureMockLocal.AddRemoteSubscription(infra.Ctx(), &azureMockRemote)

		By("And Given local Azure VNET exists", func() {
			err := azureMockLocal.CreateNetwork(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, scope.Spec.Region, "10.200.0.0/25", nil)
			Expect(err).ToNot(HaveOccurred())
		})

		By("And Given remote Azure VNet exists with Kyma tag", func() {
			err := azureMockRemote.CreateNetwork(infra.Ctx(), remoteResourceGroup, remoteVnetName, scope.Spec.Region, "10.100.0.0/25", map[string]string{kymaName: kymaName})
			Expect(err).ToNot(HaveOccurred())
		})

		localKcpNetworkName := common.KcpNetworkKymaCommonName(scope.Name)
		remoteKcpNetworkName := scope.Name + "--remote"

		var localKcpNet *cloudcontrolv1beta1.Network

		By("And Given local KCP Network is created", func() {
			localKcpNet = (&cloudcontrolv1beta1.NetworkBuilder{}).
				WithScope(scope.Name).
				WithAzureRef(scope.Spec.Scope.Azure.TenantId, scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.VpcNetwork, scope.Spec.Scope.Azure.VpcNetwork).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet, WithName(localKcpNetworkName)).
				Should(Succeed())
		})

		var remoteKcpNet *cloudcontrolv1beta1.Network

		By("And Given remote KCP Network is created", func() {
			remoteKcpNet = (&cloudcontrolv1beta1.NetworkBuilder{}).
				WithScope(scope.Name).
				WithAzureRef(scope.Spec.Scope.Azure.TenantId, remoteSubscription, remoteResourceGroup, remoteVnetName).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet, WithName(remoteKcpNetworkName)).
				Should(Succeed())
		})

		var kcpPeering *cloudcontrolv1beta1.VpcPeering

		By("And Given KCP VpcPeering is created", func() {
			kcpPeering = (&cloudcontrolv1beta1.VpcPeeringBuilder{}).
				WithScope(kymaName).
				WithRemoteRef("skr-namespace", "skr-azure-vpcpeering").
				WithDetails(localKcpNetworkName, infra.KCP().Namespace(), remoteKcpNetworkName, infra.KCP().Namespace(), remotePeeringName, true, true).
				WithLocalPeeringName(localPeeringName).
				WithUseRemoteGateway(true).
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					WithName(kcpPeeringName),
				).
				Should(Succeed())
		})

		By("And Given local Azure VPC peering is created", func() {
			Eventually(azureMockLocal.SetPeeringConnectedFullInSync).
				WithArguments(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, localPeeringName).
				Should(Succeed())
		})

		By("And Given remote Azure VPC peering is created", func() {
			Eventually(azureMockRemote.SetPeeringConnectedFullInSync).
				WithArguments(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName).
				Should(Succeed())
		})

		// Ready ==========================================================

		By("And Given KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("When local network address space is changed", func() {
			err := azureMockLocal.SetNetworkAddressSpace(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, "10.200.0.0/24")
			Expect(err).ToNot(HaveOccurred())
		})

		By("And When remote network address space is changed", func() {
			err := azureMockRemote.SetNetworkAddressSpace(infra.Ctx(), remoteResourceGroup, remoteVnetName, "10.100.0.0/24")
			Expect(err).ToNot(HaveOccurred())
		})

		By("Then local Azure network address space equals 10.200.0.0/24", func() {
			localNetwork, err := azureMockLocal.GetNetwork(infra.Ctx(), localResourceGroupName, localVirtualNetworkName)
			Expect(err).ToNot(HaveOccurred())
			Expect(*localNetwork.Properties.AddressSpace.AddressPrefixes[0]).To(Equal("10.200.0.0/24"))
		})

		By("And Then remote Azure peering is out of sync", func() {
			err := azureMockRemote.SetPeeringSyncLevel(infra.Ctx(),
				remoteResourceGroup,
				remoteVnetName,
				remotePeeringName,
				armnetwork.VirtualNetworkPeeringLevelLocalNotInSync)
			Expect(err).To(Succeed())
		})

		By("And Then remote Azure peering has PeeringSyncLevel equals VirtualNetworkPeeringLevelLocalNotInSync", func() {
			remoteAzurePeering, err := azureMockRemote.GetPeering(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName)
			Expect(err).ToNot(HaveOccurred())
			Expect(ptr.Deref(remoteAzurePeering.Properties.PeeringSyncLevel, "")).To(Equal(armnetwork.VirtualNetworkPeeringLevelLocalNotInSync))
		})

		By("And Then remote Azure network address space equals 10.100.0.0/24", func() {
			remoteNetwork, err := azureMockRemote.GetNetwork(infra.Ctx(), remoteResourceGroup, remoteVnetName)
			Expect(err).ToNot(HaveOccurred())
			Expect(*remoteNetwork.Properties.AddressSpace.AddressPrefixes[0]).To(Equal("10.100.0.0/24"))
		})

		By("And Then local Azure peering is out of sync", func() {
			err := azureMockLocal.SetPeeringSyncLevel(infra.Ctx(),
				localResourceGroupName,
				localVirtualNetworkName,
				localPeeringName,
				armnetwork.VirtualNetworkPeeringLevelLocalNotInSync)
			Expect(err).To(Succeed())
		})

		By("And Then local Azure peering has PeeringSyncLevel equals VirtualNetworkPeeringLevelLocalNotInSync", func() {
			localAzurePeering, err := azureMockLocal.GetPeering(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, localPeeringName)
			Expect(err).ToNot(HaveOccurred())
			Expect(ptr.Deref(localAzurePeering.Properties.PeeringSyncLevel, "")).To(Equal(armnetwork.VirtualNetworkPeeringLevelLocalNotInSync))
		})

		By("And Then Cloud Manager syncs local and remote peering", func() {
			// trigger the reconciliation
			_, err := composed.PatchObjMergeAnnotation(infra.Ctx(), "test", "1", kcpPeering, infra.KCP().Client())
			Expect(err).To(Succeed())
		})

		By("And Then local Azure peering address spaces are synced", func() {
			var localPeering *armnetwork.VirtualNetworkPeering
			Eventually(func() error {
				peering, err := azureMockLocal.GetPeering(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, localPeeringName)
				if err != nil {
					return err
				}

				if ptr.Deref(peering.Properties.PeeringSyncLevel, "") != armnetwork.VirtualNetworkPeeringLevelFullyInSync {
					return fmt.Errorf("local peering is not fully synced")
				}
				localPeering = peering
				return nil
			}).Should(Succeed())

			Expect(localPeering).ToNot(BeNil())
			Expect(ptr.Deref(localPeering.Properties.LocalAddressSpace.AddressPrefixes[0], "")).To(Equal("10.200.0.0/24"))
			Expect(ptr.Deref(localPeering.Properties.RemoteAddressSpace.AddressPrefixes[0], "")).To(Equal("10.100.0.0/24"))
		})

		By("And Then remote Azure peering address spaces are synced", func() {
			var remotePeering *armnetwork.VirtualNetworkPeering
			Eventually(func() error {
				peering, err := azureMockRemote.GetPeering(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName)
				if err != nil {
					return err
				}

				if ptr.Deref(peering.Properties.PeeringSyncLevel, "") != armnetwork.VirtualNetworkPeeringLevelFullyInSync {
					return fmt.Errorf("remote peering is not fully synced")
				}
				remotePeering = peering
				return nil
			}).Should(Succeed())

			Expect(remotePeering).ToNot(BeNil())
			Expect(remotePeering.Properties.LocalAddressSpace.AddressPrefixes[0]).To(Equal("10.100.0.0/24"))
			Expect(remotePeering.Properties.RemoteAddressSpace.AddressPrefixes[0]).To(Equal("10.200.0.0/24"))
		})

		// DELETE

		By("When KCP VpcPeering is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "failed deleting VpcPeering")
		})

		By("Then KCP VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "expected VpcPeering not to exist (be deleted), but it still exists")
		})

		By("And Then local Azure peering does not exist", func() {
			peering, err := azureMockLocal.GetPeering(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, localPeeringName)
			Expect(err).To(HaveOccurred())
			Expect(azuremeta.IsNotFound(err)).To(BeTrue())
			Expect(peering).To(BeNil())
		})

		By("And Then remote Azure peering does not exists", func() {
			peering, err := azureMockRemote.GetPeering(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName)
			Expect(err).To(HaveOccurred())
			Expect(azuremeta.IsNotFound(err)).To(BeTrue())
			Expect(peering).To(BeNil())
		})

		By("// cleanup: Local KCP Network", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet).
				Should(Succeed())
		})

		By("// cleanup: Remote KCP Network", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet).
				Should(Succeed())
		})

		By("// cleanup: Scope", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})
	})

	It("Scenario: KCP Azure VpcPeering does not sync", func() {
		const (
			kymaName            = "b982e044-3030-4f08-b4ce-1a14594284c6"
			kcpPeeringName      = "13c0671c-bcf5-43a9-94e8-d2300e2b0325"
			remoteSubscription  = "a8480909-af1c-447a-a93b-4a6e863c7c16"
			remoteResourceGroup = "MyResourceGroup"
			remoteVnetName      = "MyVnet"
			remotePeeringName   = "my-peering"
			localPeeringName    = "kyma-peering"
		)

		if feature.VpcPeeringSync.Value(infra.Ctx()) {
			Skip("VPC peering sync for Azure is enabled")
		}

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAzure).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		localResourceGroupName := scope.Spec.Scope.Azure.VpcNetwork
		localVirtualNetworkName := scope.Spec.Scope.Azure.VpcNetwork

		azureMockLocal := infra.AzureMock().MockConfigs(scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.TenantId)
		azureMockRemote := infra.AzureMock().MockConfigs(remoteSubscription, scope.Spec.Scope.Azure.TenantId)

		azureMockRemote.AddRemoteSubscription(infra.Ctx(), &azureMockLocal)
		azureMockLocal.AddRemoteSubscription(infra.Ctx(), &azureMockRemote)

		By("And Given local Azure VNET exists", func() {
			err := azureMockLocal.CreateNetwork(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, scope.Spec.Region, "10.200.0.0/25", nil)
			Expect(err).ToNot(HaveOccurred())
		})

		By("And Given remote Azure VNet exists with Kyma tag", func() {
			err := azureMockRemote.CreateNetwork(infra.Ctx(), remoteResourceGroup, remoteVnetName, scope.Spec.Region, "10.100.0.0/25", map[string]string{kymaName: kymaName})
			Expect(err).ToNot(HaveOccurred())
		})

		localKcpNetworkName := common.KcpNetworkKymaCommonName(scope.Name)
		remoteKcpNetworkName := scope.Name + "--remote"

		var localKcpNet *cloudcontrolv1beta1.Network

		By("And Given local KCP Network is created", func() {
			localKcpNet = (&cloudcontrolv1beta1.NetworkBuilder{}).
				WithScope(scope.Name).
				WithAzureRef(scope.Spec.Scope.Azure.TenantId, scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.VpcNetwork, scope.Spec.Scope.Azure.VpcNetwork).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet, WithName(localKcpNetworkName)).
				Should(Succeed())
		})

		var remoteKcpNet *cloudcontrolv1beta1.Network

		By("And Given remote KCP Network is created", func() {
			remoteKcpNet = (&cloudcontrolv1beta1.NetworkBuilder{}).
				WithScope(scope.Name).
				WithAzureRef(scope.Spec.Scope.Azure.TenantId, remoteSubscription, remoteResourceGroup, remoteVnetName).
				Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet, WithName(remoteKcpNetworkName)).
				Should(Succeed())
		})

		var kcpPeering *cloudcontrolv1beta1.VpcPeering

		By("And Given KCP VpcPeering is created", func() {
			kcpPeering = (&cloudcontrolv1beta1.VpcPeeringBuilder{}).
				WithScope(kymaName).
				WithRemoteRef("skr-namespace", "skr-azure-vpcpeering").
				WithDetails(localKcpNetworkName, infra.KCP().Namespace(), remoteKcpNetworkName, infra.KCP().Namespace(), remotePeeringName, true, true).
				WithLocalPeeringName(localPeeringName).
				WithUseRemoteGateway(true).
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					WithName(kcpPeeringName),
				).
				Should(Succeed())
		})

		By("And Given local Azure VPC peering is created", func() {
			Eventually(azureMockLocal.SetPeeringConnectedFullInSync).
				WithArguments(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, localPeeringName).
				Should(Succeed())
		})

		By("And Given remote Azure VPC peering is created", func() {
			Eventually(azureMockRemote.SetPeeringConnectedFullInSync).
				WithArguments(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName).
				Should(Succeed())
		})

		// Ready ==========================================================

		By("And Given KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		// Local and remote address space changes =========================

		By("When local network address space is changed", func() {
			err := azureMockLocal.SetNetworkAddressSpace(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, "10.200.0.0/24")
			Expect(err).ToNot(HaveOccurred())
		})

		By("And When remote network address space is changed", func() {
			err := azureMockRemote.SetNetworkAddressSpace(infra.Ctx(), remoteResourceGroup, remoteVnetName, "10.100.0.0/24")
			Expect(err).ToNot(HaveOccurred())
		})

		By("Then local Azure network address space equals 10.200.0.0/24", func() {
			localNetwork, err := azureMockLocal.GetNetwork(infra.Ctx(), localResourceGroupName, localVirtualNetworkName)
			Expect(err).ToNot(HaveOccurred())
			Expect(*localNetwork.Properties.AddressSpace.AddressPrefixes[0]).To(Equal("10.200.0.0/24"))
		})

		By("And Then remote Azure network address space equals 10.100.0.0/24", func() {
			remoteNetwork, err := azureMockRemote.GetNetwork(infra.Ctx(), remoteResourceGroup, remoteVnetName)
			Expect(err).ToNot(HaveOccurred())
			Expect(*remoteNetwork.Properties.AddressSpace.AddressPrefixes[0]).To(Equal("10.100.0.0/24"))
		})

		By("And Then remote Azure peering is out of sync", func() {
			err := azureMockRemote.SetPeeringSyncLevel(infra.Ctx(),
				remoteResourceGroup,
				remoteVnetName,
				remotePeeringName,
				armnetwork.VirtualNetworkPeeringLevelLocalNotInSync)
			Expect(err).To(Succeed())
		})

		By("And Then remote Azure peering has PeeringSyncLevel equals VirtualNetworkPeeringLevelLocalNotInSync", func() {
			remoteAzurePeering, err := azureMockRemote.GetPeering(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName)
			Expect(err).ToNot(HaveOccurred())
			Expect(ptr.Deref(remoteAzurePeering.Properties.PeeringSyncLevel, "")).To(Equal(armnetwork.VirtualNetworkPeeringLevelLocalNotInSync))
		})

		By("And Then local Azure peering is out of sync", func() {
			err := azureMockLocal.SetPeeringSyncLevel(infra.Ctx(),
				localResourceGroupName,
				localVirtualNetworkName,
				localPeeringName,
				armnetwork.VirtualNetworkPeeringLevelLocalNotInSync)
			Expect(err).To(Succeed())
		})

		By("And Then local Azure peering has PeeringSyncLevel equals VirtualNetworkPeeringLevelLocalNotInSync", func() {
			localAzurePeering, err := azureMockLocal.GetPeering(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, localPeeringName)
			Expect(err).ToNot(HaveOccurred())
			Expect(ptr.Deref(localAzurePeering.Properties.PeeringSyncLevel, "")).To(Equal(armnetwork.VirtualNetworkPeeringLevelLocalNotInSync))
		})

		// Cloud Manager reconciles =======================================

		By("And Then Cloud Manager detects out of sync but does not sync local and remote peering", func() {
			// trigger the reconciliation
			_, err := composed.PatchObjMergeAnnotation(infra.Ctx(), "test", "1", kcpPeering, infra.KCP().Client())
			Expect(err).To(Succeed())

			// Removing ready condition to make sure that reconciliation happened
			Expect(meta.RemoveStatusCondition(&kcpPeering.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)).To(BeTrue())

			Expect(composed.PatchObjStatus(infra.Ctx(), localKcpNet, infra.KCP().Client())).
				To(Succeed())
		})

		By("When KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("Then local Azure peering address spaces are not synced", func() {
			peering, err := azureMockLocal.GetPeering(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, localPeeringName)
			Expect(err).ToNot(HaveOccurred())
			Expect(peering).ToNot(BeNil())

			Expect(ptr.Deref(peering.Properties.LocalAddressSpace.AddressPrefixes[0], "")).To(Equal("10.200.0.0/25"))
			Expect(ptr.Deref(peering.Properties.RemoteAddressSpace.AddressPrefixes[0], "")).To(Equal("10.100.0.0/25"))
			Expect(ptr.Deref(peering.Properties.PeeringSyncLevel, "")).To(Equal(armnetwork.VirtualNetworkPeeringLevelLocalNotInSync))
		})

		By("And Then remote Azure peering address spaces are synced", func() {
			peering, err := azureMockRemote.GetPeering(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName)
			Expect(err).ToNot(HaveOccurred())
			Expect(peering).ToNot(BeNil())

			Expect(ptr.Deref(peering.Properties.LocalAddressSpace.AddressPrefixes[0], "")).To(Equal("10.100.0.0/25"))
			Expect(ptr.Deref(peering.Properties.RemoteAddressSpace.AddressPrefixes[0], "")).To(Equal("10.200.0.0/25"))
			Expect(ptr.Deref(peering.Properties.PeeringSyncLevel, "")).To(Equal(armnetwork.VirtualNetworkPeeringLevelLocalNotInSync))
		})

		// DELETE

		By("When KCP VpcPeering is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "failed deleting VpcPeering")
		})

		By("Then KCP VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering).
				Should(Succeed(), "expected VpcPeering not to exist (be deleted), but it still exists")
		})

		By("And Then local Azure peering does not exist", func() {
			peering, err := azureMockLocal.GetPeering(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, localPeeringName)
			Expect(err).To(HaveOccurred())
			Expect(azuremeta.IsNotFound(err)).To(BeTrue())
			Expect(peering).To(BeNil())
		})

		By("And Then remote Azure peering does not exists", func() {
			peering, err := azureMockRemote.GetPeering(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName)
			Expect(err).To(HaveOccurred())
			Expect(azuremeta.IsNotFound(err)).To(BeTrue())
			Expect(peering).To(BeNil())
		})

		By("// cleanup: Local KCP Network", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), localKcpNet).
				Should(Succeed())
		})

		By("// cleanup: Remote KCP Network", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet).
				Should(Succeed())
		})

		By("// cleanup: Scope", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})
	})
})
