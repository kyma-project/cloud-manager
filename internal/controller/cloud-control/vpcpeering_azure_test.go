package cloudcontrol

import (
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Feature: KCP VpcPeering", func() {

	It("Scenario: KCP Azure VpcPeering is created", func() {
		const (
			kymaName            = "6a62936d-aa6e-4d5b-aaaa-5eae646d1bd5"
			vpcpeeringName      = "281bc581-8635-4d56-ba52-fa48ec6f7c69"
			remoteSubscription  = "afdbc79f-de19-4df4-94cd-6be2739dc0e0"
			remoteResourceGroup = "MyResourceGroup"
			remoteVnetName      = "MyVnet"
			remoteRefNamespace  = "skr-namespace"
			remoteRefName       = "skr-azure-vpcpeering"
		)

		remoteVnetId := util.VirtualNetworkResourceId(remoteSubscription, remoteResourceGroup, remoteVnetName)
		remotePeeringId := util.VirtualNetworkPeeringResourceId(remoteSubscription, remoteResourceGroup, remoteVnetName, vpcpeeringName)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			scopePkg.Ignore.AddName(kymaName)

			Eventually(CreateScopeAzure).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		localVirtualNetworkName := scope.Spec.Scope.Azure.VpcNetwork
		localResourceGroupName := localVirtualNetworkName //TODO resource group name is the same as VPC name

		azureMockLocal := infra.AzureMock().MockConfigs(scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.TenantId)
		azureMockRemote := infra.AzureMock().MockConfigs(remoteSubscription, scope.Spec.Scope.Azure.TenantId)

		By("And Given local Azure VNET exists", func() {
			_, err := azureMockLocal.CreateNetwork(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, scope.Spec.Region, "10.200.0.0/25", nil)
			Expect(err).ToNot(HaveOccurred())
		})

		By("And Given remote Azure VNet exists with Kyma tag", func() {
			_, err := azureMockRemote.CreateNetwork(infra.Ctx(), remoteResourceGroup, remoteVnetName, scope.Spec.Region, "10.100.0.0/25", map[string]string{kymaName: kymaName})
			Expect(err).ToNot(HaveOccurred())
		})

		obj := &cloudcontrolv1beta1.VpcPeering{}

		By("When KCP VpcPeering is created", func() {
			Eventually(CreateKcpVpcPeering).
				WithArguments(infra.Ctx(), infra.KCP().Client(), obj,
					WithName(vpcpeeringName),
					WithKcpVpcPeeringRemoteRef(remoteRefNamespace, remoteRefName),
					WithScope(kymaName),
					WithKcpVpcPeeringSpecAzure(true, vpcpeeringName, remoteVnetId, remoteResourceGroup),
				).
				Should(Succeed())
		})

		var localAzurePeering *armnetwork.VirtualNetworkPeering

		By("Then local Azure VPC Peering is created", func() {
			Eventually(func() error {
				p, err := azureMockLocal.GetPeering(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, vpcpeeringName)
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

		// this was a bit confusing at first, but it's actually checking through the peering ID that
		// it was created in the appropriate resource group, network and with name as specified in the KCP VpcPeering resource
		By("And Then local Azure Peering has RemoteVirtualNetwork.ID equal to KCP VpcPeering remote vpc id", func() {
			Expect(ptr.Deref(localAzurePeering.Properties.RemoteVirtualNetwork.ID, "xxx")).To(Equal(remoteVnetId))
		})

		var remoteAzurePeering *armnetwork.VirtualNetworkPeering

		By("And Then remote Azure VPC Peering is created", func() {
			Eventually(func() error {
				p, err := azureMockRemote.GetPeering(infra.Ctx(), remoteResourceGroup, remoteVnetName, vpcpeeringName)
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

		// this was a bit confusing at first, but it's actually checking through the peering ID that
		// it was created in the appropriate resource group, network and with name as specified in the KCP VpcPeering resource
		By("And Then remote Azure Peering has ID equal to KCP VpcPeering remote vpc peering id", func() {
			Expect(ptr.Deref(remoteAzurePeering.ID, "xxx")).To(Equal(remotePeeringId))
		})

		By("Then KCP VpcPeering state is Initiated", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), obj,
					NewObjActions(),
					HavingState(cloudcontrolv1beta1.VirtualNetworkPeeringStateInitiated),
				).Should(Succeed())
		})

		By("When Azure VPC Peering state is Connected", func() {
			err := azureMockLocal.SetPeeringStateConnected(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, vpcpeeringName)
			Expect(err).ToNot(HaveOccurred())
			err = azureMockRemote.SetPeeringStateConnected(infra.Ctx(), remoteResourceGroup, remoteVnetName, vpcpeeringName)
			Expect(err).ToNot(HaveOccurred())

			p, err := azureMockLocal.GetPeering(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, vpcpeeringName)
			Expect(err).ToNot(HaveOccurred())
			Expect(p).ToNot(BeNil())
			localAzurePeering = p

			p, err = azureMockRemote.GetPeering(infra.Ctx(), remoteResourceGroup, remoteVnetName, vpcpeeringName)
			Expect(err).ToNot(HaveOccurred())
			Expect(p).ToNot(BeNil())
			remoteAzurePeering = p
		})

		By("Then KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), obj,
					NewObjActions(),
					HaveFinalizer(cloudcontrolv1beta1.FinalizerName),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then KCP VpcPeering has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(obj, cloudcontrolv1beta1.FinalizerName)).
				To(BeTrue())
		})

		By("And Then KCP VpcPeering has status.id equal to local Azure Peering ID", func() {
			Expect(obj.Status.Id).To(Equal(ptr.Deref(localAzurePeering.ID, "xxx")))
		})

		By("And Then KCP VpcPeering has status.remoteId equal to remote Azure Peering ID", func() {
			Expect(obj.Status.RemoteId).To(Equal(ptr.Deref(remoteAzurePeering.ID, "xxx")))
		})

		// DELETE

		By("When KCP VpcPeering is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), obj).
				Should(Succeed(), "failed deleting VpcPeering")
		})

		By("Then KCP VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), obj).
				Should(Succeed(), "expected VpcPeering not to exist (be deleted), but it still exists")
		})

		By("And Then local Azure peering does not exist", func() {
			peering, err := azureMockLocal.GetPeering(infra.Ctx(), localResourceGroupName, localVirtualNetworkName, vpcpeeringName)
			Expect(err).To(HaveOccurred())
			Expect(azuremeta.IsNotFound(err)).To(BeTrue())
			Expect(peering).To(BeNil())
		})

		By("// cleanup: Scope", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})

	})

})
