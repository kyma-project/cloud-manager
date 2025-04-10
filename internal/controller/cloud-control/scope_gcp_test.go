package cloudcontrol

import (
	"fmt"
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	kcpnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/network"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Feature: KCP Scope", func() {

	It("Scenario: KCP GCP Scope is created and deleted when GardenerCluster CR is created and deleted", func() {
		const (
			kymaName = "51485d74-0e28-44f9-ae80-3088128d8747"
		)

		kymaNetworkName := common.KcpNetworkKymaCommonName(kymaName)
		kcpnetwork.Ignore.AddName(kymaNetworkName)

		// Set the path to an arbitrary file path to prevent errors
		Expect(os.Setenv("GCP_SA_JSON_KEY_PATH", "testdata/serviceaccount.json")).
			To(Succeed())

		shoot := &gardenertypes.Shoot{}

		By("Given Shoot exists", func() {
			Eventually(CreateShootGcp).
				WithArguments(infra.Ctx(), infra, shoot, WithName(kymaName)).
				Should(Succeed(), "failed creating garden shoot for GCP")
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
				cloudcontrolv1beta1.LabelScopeBrokerPlanName:  string(cloudcontrolv1beta1.ProviderGCP),
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
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope, NewObjActions(), HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed(), "expected Scope to have Ready condition")
		})

		By("And Then Scope has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(scope, api.CommonFinalizerDeletionHook)).To(BeTrue())
		})

		for _, label := range cloudcontrolv1beta1.ScopeLabels {
			By(fmt.Sprintf("And Then Scope has label %s", label), func() {
				Expect(scope.Labels).To(HaveKeyWithValue(label, gardenerClusterCR.GetLabels()[label]))
			})
		}

		By("And Then Scope provider is GCP", func() {
			Expect(scope.Spec.Provider).To(Equal(cloudcontrolv1beta1.ProviderGCP))
		})

		By("And Then Scope has spec.kymaName to equal shoot.name", func() {
			Expect(scope.Spec.KymaName).To(Equal(shoot.Name), "expected Scope.spec.kymaName to equal shoot.name")
		})

		By("And Then Scope has spec.region equal to shoot.spec.region", func() {
			Expect(scope.Spec.Region).To(Equal(shoot.Spec.Region), "expected Shoot.spec.region equal to shoot.spec.region")
		})

		By("And Then Scope has nil spec.scope.azure", func() {
			Expect(scope.Spec.Scope.Azure).To(BeNil(), "expected Shoot.spec.scope.azure to be nil")
		})

		By("And Then Scope has nil spec.scope.aws", func() {
			Expect(scope.Spec.Scope.Aws).To(BeNil(), "expected Shoot.spec.scope.aws to be nil")
		})

		By("And Then Scope has GCP subscriptionId and tenantId", func() {
			Expect(scope.Spec.Scope.Gcp).NotTo(BeNil())
			Expect(scope.Spec.Scope.Gcp.Project).To(Equal("project_id")) // fixed value from CreateShootGcp
		})

		By("And Then Scope has vpc network name", func() {
			Expect(scope.Spec.Scope.Gcp.VpcNetwork).To(Equal(common.GardenerVpcName(DefaultGardenNamespace, shoot.Name)))
		})

		By("And Then Scope has nodes/podes/services cidr as shoot", func() {
			Expect(scope.Spec.Scope.Gcp.Network.Nodes).To(Equal(ptr.Deref(shoot.Spec.Networking.Nodes, "")))
			Expect(scope.Spec.Scope.Gcp.Network.Pods).To(Equal(ptr.Deref(shoot.Spec.Networking.Pods, "")))
			Expect(scope.Spec.Scope.Gcp.Network.Services).To(Equal(ptr.Deref(shoot.Spec.Networking.Services, "")))
		})

		By("And Then Scope has worker zones as shoot", func() {
			Expect(scope.Spec.Scope.Gcp.Workers).To(HaveLen(1))
			Expect(scope.Spec.Scope.Gcp.Workers[0].Zones).To(HaveLen(3))
			Expect(scope.Spec.Scope.Gcp.Workers[0].Zones[0]).To(Equal("europe-west1-a")) // fixed value from CreateShootGcp
			Expect(scope.Spec.Scope.Gcp.Workers[0].Zones[1]).To(Equal("europe-west1-b")) // fixed value from CreateShootGcp
			Expect(scope.Spec.Scope.Gcp.Workers[0].Zones[2]).To(Equal("europe-west1-c")) // fixed value from CreateShootGcp
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

		By("And Then Kyma Network has GCP reference details", func() {
			Expect(kymaNetwork.Spec.Network.Reference).NotTo(BeNil())
			Expect(kymaNetwork.Spec.Network.Reference.Gcp).NotTo(BeNil())
			Expect(kymaNetwork.Spec.Network.Reference.Gcp.NetworkName).To(Equal(scope.Spec.Scope.Gcp.VpcNetwork))
			Expect(kymaNetwork.Spec.Network.Reference.Gcp.GcpProject).To(Equal(scope.Spec.Scope.Gcp.Project))
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
