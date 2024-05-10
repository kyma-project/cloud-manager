package cloudcontrol

import (
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
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
			remoteVnet          = "/subscriptions/9c05f3c1-314b-4c4b-bfff-b5a0650177cb/resourceGroups/MyResourceGroup/providers/Microsoft.Network/virtualNetworks/MyVnet"
			remoteResourceGroup = "MyResourceGroup"
			subscriptionId      = "3f1d2fbd-117a-4742-8bde-6edbcdee6a04"
			remoteRefNamespace  = "skr-namespace"
			remoteRefName       = "skr-azure-vpcpeering"
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
		virtualNetworkPeeringName := fmt.Sprintf("%s-%s",
			remoteRefNamespace,
			remoteRefName)

		connectionId := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s/virtualNetworkPeerings/%s",
			subscriptionId,
			resourceGroupName,
			virtualNetworkName,
			virtualNetworkPeeringName)

		vpcpeering := &cloudcontrolv1beta1.VpcPeering{}

		infra.AzureMock().SetSubscription(subscriptionId)

		By("When KCP VpcPeering is created", func() {
			Eventually(CreateKcpVpcPeering).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					WithName(vpcpeeringName),
					WithKcpVpcPeeringRemoteRef(remoteRefNamespace, remoteRefName),
					WithKcpVpcPeeringSpecScope(kymaName),
					WithKcpVpcPeeringSpecAzure(true, remoteVnet, remoteResourceGroup),
				).
				Should(Succeed())
		})

		By("Then KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					NewObjActions(),
					HaveFinalizer(cloudcontrolv1beta1.FinalizerName),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then KCP VpcPeering has status.ConnectionId equal to existing AWS Connection id", func() {
			Expect(vpcpeering.Status.ConnectionId).To(Equal(connectionId))
		})
	})

})
