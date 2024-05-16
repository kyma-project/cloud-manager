package cloudcontrol

import (
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP VpcPeering", func() {

	It("Scenario: KCP Azure VpcPeering is created", func() {
		const (
			kymaName            = "6a62936d-aa6e-4d5b-aaaa-5eae646d1bd5"
			vpcpeeringName      = "281bc581-8635-4d56-ba52-fa48ec6f7c69"
			subscriptionId      = "2bfba5a4-c5d1-4b03-a7db-4ead64232fd6"
			remoteSubscription  = "afdbc79f-de19-4df4-94cd-6be2739dc0e0"
			remoteResourceGroup = "MyResourceGroup"
			remoteVnetName      = "MyVnet"
			remoteRefNamespace  = "skr-namespace"
			remoteRefName       = "skr-azure-vpcpeering"
			remoteVnet          = "/subscriptions/afdbc79f-de19-4df4-94cd-6be2739dc0e0/resourceGroups/MyResourceGroup/providers/Microsoft.Network/virtualNetworks/MyVnet"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			scopePkg.Ignore.AddName(kymaName)

			Eventually(CreateScopeAzure).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		virtualNetworkName := scope.Spec.Scope.Azure.VpcNetwork
		resourceGroupName := virtualNetworkName //TODO resource group name is the same as VPC name

		obj := &cloudcontrolv1beta1.VpcPeering{}

		infra.AzureMock().SetSubscription(subscriptionId)

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

		peering, _ := infra.AzureMock().Get(infra.Ctx(), resourceGroupName, virtualNetworkName, vpcpeeringName)

		By("And Then found VirtualNetworkPeering has ID equal to Status.Id", func() {
			Expect(peering.ID, obj.Status.Id)
		})

		virtualNetworkPeeringName := fmt.Sprintf("%s-%s",
			remoteRefNamespace,
			remoteRefName)

		remoteConnectionId := util.VirtualNetworkPeeringResourceId(remoteSubscription, remoteResourceGroup, virtualNetworkName, virtualNetworkPeeringName)
		By("And Then found VirtualNetworkPeering has RemoteVirtualNetwork.ID equal remote vpc id", func() {
			Expect(peering.Properties.RemoteVirtualNetwork.ID, remoteConnectionId)
		})

	})

})
