package cloudcontrol

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
	"time"
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

		remoteVnet := util.VirtualNetworkResourceId(remoteSubscription, remoteResourceGroup, remoteVnetName)
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			scopePkg.Ignore.AddName(kymaName)

			Eventually(CreateScopeAzure).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		virtualNetworkName := scope.Spec.Scope.Azure.VpcNetwork
		subscriptionId := scope.Spec.Scope.Azure.SubscriptionId
		resourceGroupName := virtualNetworkName //TODO resource group name is the same as VPC name

		infra.AzureMock().AddNetwork(remoteSubscription, remoteResourceGroup, remoteVnetName, map[string]*string{"shoot-name": ptr.To(kymaName)})

		obj := &cloudcontrolv1beta1.VpcPeering{}

		By("When KCP VpcPeering is created", func() {
			Eventually(CreateKcpVpcPeering).
				WithArguments(infra.Ctx(), infra.KCP().Client(), obj,
					WithName(vpcpeeringName),
					WithKcpVpcPeeringRemoteRef(remoteRefNamespace, remoteRefName),
					WithKcpVpcPeeringSpecScope(kymaName),
					WithKcpVpcPeeringSpecAzure(true, remoteVnet, remoteResourceGroup),
				).
				Should(Succeed())
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

		c, _ := infra.AzureMock().VpcPeeringSkrProvider()(infra.Ctx(), "", "", subscriptionId, "")

		peering, _ := c.GetPeering(infra.Ctx(), resourceGroupName, virtualNetworkName, vpcpeeringName)

		By("And Then found VirtualNetworkPeering has ID equal to Status.Id", func() {
			Expect(ptr.Deref(peering.ID, "xxx")).To(Equal(obj.Status.Id))
		})

		virtualNetworkPeeringName := fmt.Sprintf("%s-%s",
			remoteRefNamespace,
			remoteRefName)

		r, _ := infra.AzureMock().VpcPeeringSkrProvider()(infra.Ctx(), "", "", remoteSubscription, "")

		remotePeering, _ := r.GetPeering(infra.Ctx(), remoteResourceGroup, remoteVnetName, virtualNetworkPeeringName)

		By("And Then found remote VirtualNetworkPeering has ID equal to Status.RemoteId", func() {
			Expect(ptr.Deref(remotePeering.ID, "xxx")).To(Equal(obj.Status.RemoteId))
		})

		By("And Then found VirtualNetworkPeering has RemoteVirtualNetwork.ID equal remote vpc id", func() {
			Expect(ptr.Deref(peering.Properties.RemoteVirtualNetwork.ID, "xxx")).To(Equal(remoteVnet))
		})

		remotePeeringId := util.VirtualNetworkPeeringResourceId(remoteSubscription, remoteResourceGroup, remoteVnetName, virtualNetworkPeeringName)

		By("And Then found remote VirtualNetworkPeering has ID equal to remote vpc peering id", func() {
			Expect(ptr.Deref(remotePeering.ID, "xxx")).To(Equal(remotePeeringId))
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

		By("And Then VirtualNetworkPeering is deleted", func() {
			peering, _ = c.GetPeering(infra.Ctx(), resourceGroupName, virtualNetworkName, vpcpeeringName)
			Expect(peering).To(Equal((*armnetwork.VirtualNetworkPeering)(nil)))
		})

	})

})
