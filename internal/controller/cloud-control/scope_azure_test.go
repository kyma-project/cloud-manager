package cloudcontrol

import (
	"fmt"
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	kcpnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/network"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Feature: KCP Scope", func() {

	It("Scenario: KCP Azure Scope is created and deleted when GardenerCluster CR is created and deleted", func() {
		const (
			kymaName = "ca5b791b-87df-40ed-bea8-f10b84c483dd"
		)

		kymaNetworkName := common.KcpNetworkKymaCommonName(kymaName)
		kcpnetwork.Ignore.AddName(kymaNetworkName)

		shoot := &gardenertypes.Shoot{}

		By("Given Shoot exists", func() {
			Eventually(CreateShootAzure).
				WithArguments(infra.Ctx(), infra, shoot, WithName(kymaName)).
				Should(Succeed(), "failed creating garden shoot for Azure")
		})

		gardenerClusterCR := util.NewGardenerClusterUnstructured()

		By("And Given GardenerCluster CR exists", func() {
			Expect(util.SetGardenerClusterSummary(gardenerClusterCR, util.GardenerClusterSummary{
				Key:       "config",
				Name:      fmt.Sprintf("kubeconfig-%s", kymaName),
				Namespace: DefaultKcpNamespace,
				Shoot:     shoot.Name,
			}))
			gardenerClusterCR.SetLabels(map[string]string{
				cloudcontrolv1beta1.LabelScopeGlobalAccountId: "ga-account-id",
				cloudcontrolv1beta1.LabelScopeSubaccountId:    "subaccount-id",
				cloudcontrolv1beta1.LabelScopeShootName:       shoot.Name,
				cloudcontrolv1beta1.LabelScopeRegion:          "region-some",
				cloudcontrolv1beta1.LabelScopeBrokerPlanName:  string(cloudcontrolv1beta1.ProviderAzure),
			})
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gardenerClusterCR, WithName(kymaName)).
				Should(Succeed(), "failed creating GardenerCluster CR")
		})

		scope := &cloudcontrolv1beta1.Scope{}

		By("Then Scope is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope, NewObjActions(WithName(kymaName))).
				Should(Succeed(), "expected Scope to be created")
		})

		By("And Then Scope has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope, NewObjActions(), HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady)).
				Should(Succeed(), "expected created Scope to have Ready condition")
		})

		By("And Then Scope has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(scope, api.CommonFinalizerDeletionHook)).To(BeTrue())
		})

		for _, label := range cloudcontrolv1beta1.ScopeLabels {
			By(fmt.Sprintf("And Then Scope has label %s", label), func() {
				Expect(scope.Labels).To(HaveKeyWithValue(label, gardenerClusterCR.GetLabels()[label]))
			})
		}

		By("And Then Scope provider is Azure", func() {
			Expect(scope.Spec.Provider).To(Equal(cloudcontrolv1beta1.ProviderAzure))
		})

		By("And Then Scope has spec.kymaName to equal shoot.name", func() {
			Expect(scope.Spec.KymaName).To(Equal(shoot.Name), "expected Scope.spec.kymaName to equal shoot.name")
		})

		By("And Then Scope has spec.region equal to shoot.spec.region", func() {
			Expect(scope.Spec.Region).To(Equal(shoot.Spec.Region), "expected Shoot.spec.region equal to shoot.spec.region")
		})

		By("And Then Scope has spec.scope.aws equal to nil", func() {
			Expect(scope.Spec.Scope.Aws).To(BeNil(), "expected Shoot.spec.scope.azure to be nil")
		})

		By("And Then Scope has spec.scope.gcp equal to nil", func() {
			Expect(scope.Spec.Scope.Gcp).To(BeNil(), "expected Shoot.spec.scope.gcp to be nil")
		})

		By("And Then Scope has Azure subscriptionId and tenantId", func() {
			Expect(scope.Spec.Scope.Azure).NotTo(BeNil())
			Expect(scope.Spec.Scope.Azure.SubscriptionId).To(Equal("someAzureSubscriptionId")) // fixed value from CreateShootAzure
			Expect(scope.Spec.Scope.Azure.TenantId).To(Equal("someAzureTenantId"))             // fixed value from CreateShootAzure
		})

		By("And Then Scope has vpc network name", func() {
			Expect(scope.Spec.Scope.Azure.VpcNetwork).To(Equal(common.GardenerVpcName(DefaultGardenNamespace, shoot.Name)))
		})

		By("And Then Scope has spec.scope.azure.network.zones as shoot", func() {
			Expect(scope.Spec.Scope.Azure.Network.Zones).To(HaveLen(3))
			Expect(scope.Spec.Scope.Azure.Network.Zones[0].Name).To(Equal("2")) // as set in CreateShootAzure
			Expect(scope.Spec.Scope.Azure.Network.Zones[1].Name).To(Equal("3")) // as set in CreateShootAzure
			Expect(scope.Spec.Scope.Azure.Network.Zones[2].Name).To(Equal("1")) // as set in CreateShootAzure
		})

		kymaNetwork := &cloudcontrolv1beta1.Network{}

		By("And Then Kyma Network is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kymaNetwork, NewObjActions(WithName(kymaNetworkName))).
				Should(Succeed(), "expected Kyma Network to be created")
		})

		By("And Then Kyma Network has 'kyma' type", func() {
			Expect(kymaNetwork.Spec.Type).To(Equal(cloudcontrolv1beta1.NetworkTypeKyma))
		})

		By("And Then Kyma Network has scope reference", func() {
			Expect(kymaNetwork.Spec.Scope.Name).To(Equal(scope.Name))
		})

		By("And Then Kyma Network has Azure reference details", func() {
			Expect(kymaNetwork.Spec.Network.Reference).NotTo(BeNil())
			Expect(kymaNetwork.Spec.Network.Reference.Azure).NotTo(BeNil())
			Expect(kymaNetwork.Spec.Network.Reference.Azure.NetworkName).To(Equal(scope.Spec.Scope.Azure.VpcNetwork))
			Expect(kymaNetwork.Spec.Network.Reference.Azure.TenantId).To(Equal(scope.Spec.Scope.Azure.TenantId))
			Expect(kymaNetwork.Spec.Network.Reference.Azure.SubscriptionId).To(Equal(scope.Spec.Scope.Azure.SubscriptionId))
			Expect(kymaNetwork.Spec.Network.Reference.Azure.ResourceGroup).To(Equal(scope.Spec.Scope.Azure.VpcNetwork))
		})

		// DELETE =======================================================

		By("When GardenerCluster CR is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gardenerClusterCR).
				Should(Succeed(), "expected Gardener Cluster to be deleted")
		})

		By("Then Scope does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed(), "expectedScope to be deleted")
		})

		By("And Then Kyma Network does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kymaNetwork).
				Should(Succeed(), "expected Kyma Network to be deleted")
		})

		// CLEANUP =======================================================

		By("// cleanup: delete GardenerCluster", func() {
			Expect(client.IgnoreNotFound(Delete(infra.Ctx(), infra.KCP().Client(), gardenerClusterCR))).
				To(Succeed(), "error deleting GardenerCluster CR")
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gardenerClusterCR).
				Should(Succeed(), "expected Gardener Cluster to be deleted")
		})

		By("// cleanup: delete Shoot", func() {
			Expect(Delete(infra.Ctx(), infra.Garden().Client(), shoot)).
				To(Succeed(), "error deleting Shoot CR")
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.Garden().Client(), shoot).
				Should(Succeed(), "expected Shoot to be deleted")
		})

	})

})
