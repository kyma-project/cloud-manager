package cloudcontrol

import (
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	kcpnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/network"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Feature: KCP VpcPeering", func() {

	It("Scenario: KCP Azure VpcPeering is created", func() {
		const (
			kymaName            = "6a62936d-aa6e-4d5b-aaaa-5eae646d1bd5"
			kcpPeeringName      = "281bc581-8635-4d56-ba52-fa48ec6f7c69"
			remoteSubscription  = "afdbc79f-de19-4df4-94cd-6be2739dc0e0"
			remoteResourceGroup = "MyResourceGroup"
			remoteVnetName      = "MyVnet"
			remotePeeringName   = "MyPeering"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			scopePkg.Ignore.AddName(kymaName)

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

		localKcpNetworkName := common.KymaNetworkCommonName(scope.Name)
		remoteKcpNetworkName := scope.Name + "--remote"

		var kcpPeering *cloudcontrolv1beta1.VpcPeering

		By("When KCP VpcPeering is created", func() {
			kcpPeering = (&cloudcontrolv1beta1.VpcPeeringBuilder{}).
				WithScope(kymaName).
				WithRemoteRef("skr-namespace", "skr-azure-vpcpeering").
				WithDetails(localKcpNetworkName, infra.KCP().Namespace(), remoteKcpNetworkName, infra.KCP().Namespace(), remotePeeringName, true).
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
			_, err := composed.PatchObjAddAnnotation(infra.Ctx(), "test", "1", remoteKcpNet, infra.KCP().Client())
			Expect(err).
				To(Succeed())
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteKcpNet, NewObjActions(), HavingState("Ready")).
				Should(Succeed(), "expected remote kcp network to become ready but it didn't")
		})

		// Peering Created ===============================================================

		var localAzurePeering *armnetwork.VirtualNetworkPeering

		By("Then local Azure VPC Peering is created", func() {
			Eventually(func() error {
				p, err := azureMockLocal.GetPeering(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, kcpPeeringName)
				if err != nil {
					return err
				}
				if p == nil {
					return errors.New("nil peering received")
				}
				localAzurePeering = p
				return nil
			}).Should(Succeed())
		})

		By("And Then local Azure Peering has RemoteVirtualNetwork.ID equal to KCP VpcPeering remote vpc id", func() {
			remoteVnetId := util.NewVirtualNetworkResourceId(remoteSubscription, remoteResourceGroup, remoteVnetName).String()
			Expect(ptr.Deref(localAzurePeering.Properties.RemoteVirtualNetwork.ID, "xxx")).To(Equal(remoteVnetId))
		})

		var remoteAzurePeering *armnetwork.VirtualNetworkPeering

		By("And Then remote Azure VPC Peering is created", func() {
			Eventually(func() error {
				p, err := azureMockRemote.GetPeering(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName)
				if err != nil {
					return err
				}
				if p == nil {
					return errors.New("nil peering received")
				}
				remoteAzurePeering = p
				return nil
			}).Should(Succeed())
		})

		By("And Then remote Azure Peering has RemoteVirtualNetwork.ID equal to KCP VpcPeering local vpc id", func() {
			localeVnetId := util.NewVirtualNetworkResourceId(scope.Spec.Scope.Azure.SubscriptionId, localResourceGroupName, localVirtualNetworkName).String()
			Expect(ptr.Deref(remoteAzurePeering.Properties.RemoteVirtualNetwork.ID, "xxx")).To(Equal(localeVnetId))
		})

		By("Then KCP VpcPeering state is Initiated", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HavingState(cloudcontrolv1beta1.VirtualNetworkPeeringStateInitiated),
				).Should(Succeed())
		})

		// Ready ==========================================================

		By("When Azure VPC Peering is Connected", func() {
			err := azureMockLocal.SetPeeringStateConnected(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, kcpPeeringName)
			Expect(err).ToNot(HaveOccurred())
			err = azureMockRemote.SetPeeringStateConnected(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName)
			Expect(err).ToNot(HaveOccurred())

			p, err := azureMockLocal.GetPeering(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, kcpPeeringName)
			Expect(err).ToNot(HaveOccurred())
			Expect(p).ToNot(BeNil())
			localAzurePeering = p

			p, err = azureMockRemote.GetPeering(infra.Ctx(), remoteResourceGroup, remoteVnetName, remotePeeringName)
			Expect(err).ToNot(HaveOccurred())
			Expect(p).ToNot(BeNil())
			remoteAzurePeering = p
		})

		By("Then KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpPeering,
					NewObjActions(),
					HaveFinalizer(cloudcontrolv1beta1.FinalizerName),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then KCP VpcPeering has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(kcpPeering, cloudcontrolv1beta1.FinalizerName)).
				To(BeTrue())
		})

		By("Then KCP VpcPeering state is Connected", func() {
			Expect(kcpPeering.Status.State).To(Equal(cloudcontrolv1beta1.VirtualNetworkPeeringStateConnected))
		})

		By("And Then KCP VpcPeering has status.id equal to local Azure Peering ID", func() {
			Expect(kcpPeering.Status.Id).To(Equal(ptr.Deref(localAzurePeering.ID, "xxx")))
		})

		By("And Then KCP VpcPeering has status.remoteId equal to remote Azure Peering ID", func() {
			Expect(kcpPeering.Status.RemoteId).To(Equal(ptr.Deref(remoteAzurePeering.ID, "xxx")))
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
			peering, err := azureMockLocal.GetPeering(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, kcpPeeringName)
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
