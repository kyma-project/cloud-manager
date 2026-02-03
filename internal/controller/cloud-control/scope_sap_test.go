package cloudcontrol

import (
	"fmt"

	"github.com/elliotchance/pie/v2"
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/routers"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Feature: KCP Scope", func() {

	It("Scenario: KCP SAP Scope is created and deleted when GardenerCluster CR is created and deleted", func() {
		const (
			kymaName = "bfcfcb40-90e3-4e72-aa42-615863013b59"
		)

		kymaNetworkName := common.KcpNetworkKymaCommonName(kymaName)

		sapMock := infra.SapMock().NewProject()

		By("Given OpenStack external network exists", func() {
			Expect(SapInfraInitializeExternalNetworks(infra.Ctx(), sapMock)).
				To(Succeed())
		})

		shoot := &gardenertypes.Shoot{}

		By("And Given Shoot exists", func() {
			Expect(CreateShootSap(infra.Ctx(), infra, shoot, sapMock.ProviderParams(), WithName(kymaName))).
				To(Succeed(), "failed creating garden shoot for SAP")
		})

		var sapCreatedInfra *SapGardenerInfra

		By("An Given SAP Gardener infra exists", func() {
			createdInfra, err := SapInfraGardenerCreateResourcesAllWithNetwork(infra.Ctx(), sapMock, shoot.Namespace, shoot.Name, "10.250.0.0/16")
			Expect(err).NotTo(HaveOccurred())
			sapCreatedInfra = createdInfra
		})

		gardenerClusterCR := util.NewGardenerClusterUnstructured()

		By("And Given GardenerCluster CR exists", func() {
			Expect(CreateGardenerClusterCR(infra.Ctx(), infra, gardenerClusterCR, kymaName, shoot.Name, cloudcontrolv1beta1.ProviderOpenStack)).
				To(Succeed(), "failed creating GardenerCluster CR")
		})

		// kymaNetwork is not ignored, and should reconcile into ready state with network ref in the status!!!
		// a ready kymaNetwork is a prerequisite for Scope to become ready

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

		By("And Then Scope provider is OpenStack", func() {
			Expect(scope.Spec.Provider).To(Equal(cloudcontrolv1beta1.ProviderOpenStack))
		})

		By("And Then Scope has spec.kymaName to equal shoot.name", func() {
			Expect(scope.Spec.KymaName).To(Equal(shoot.Name), "expected Scope.spec.kymaName to equal shoot.name")
		})

		By("And Then Scope has spec.region equal to shoot.spec.region", func() {
			Expect(scope.Spec.Region).To(Equal(shoot.Spec.Region), "expected Shoot.spec.region equal to shoot.spec.region")
		})

		By("And Then Scope has spec.scope.azure equal to nil", func() {
			Expect(scope.Spec.Scope.Azure).To(BeNil(), "expected Shoot.spec.scope.azure to be nil")
		})

		By("And Then Scope has spec.scope.gcp equal to nil", func() {
			Expect(scope.Spec.Scope.Gcp).To(BeNil(), "expected Shoot.spec.scope.gcp to be nil")
		})

		By("And Then Scope has spec.scope.aws equal to nil", func() {
			Expect(scope.Spec.Scope.Aws).To(BeNil(), "expected Shoot.spec.scope.aws to be nil")
		})

		By("And Then Scope has spec.scope.openstack.tenantName", func() {
			Expect(scope.Spec.Scope.OpenStack).NotTo(BeNil())
			Expect(scope.Spec.Scope.OpenStack.TenantName).NotTo(BeEmpty())
			Expect(scope.Spec.Scope.OpenStack.TenantName).To(Equal(sapMock.ProjectName()))
		})

		By("And Then Scope has spec.scope.openstack.domainName", func() {
			Expect(scope.Spec.Scope.OpenStack).NotTo(BeNil())
			Expect(scope.Spec.Scope.OpenStack.DomainName).NotTo(BeEmpty())
			Expect(scope.Spec.Scope.OpenStack.DomainName).To(Equal(sapMock.DomainName()))
		})

		By("And Then Scope has vpc network name", func() {
			Expect(scope.Spec.Scope.OpenStack.VpcNetwork).To(Equal(common.GardenerVpcName(DefaultGardenNamespace, shoot.Name)))
		})

		By("And Then Scope has spec.scope.openstack.network.zones as shoot", func() {
			Expect(scope.Spec.Scope.OpenStack.Network.Zones).To(HaveLen(3))
			Expect(scope.Spec.Scope.OpenStack.Network.Zones[0]).To(Equal("eu-de-1a")) // as set in CreateShootSap
			Expect(scope.Spec.Scope.OpenStack.Network.Zones[1]).To(Equal("eu-de-1b")) // as set in CreateShootSap
			Expect(scope.Spec.Scope.OpenStack.Network.Zones[2]).To(Equal("eu-de-1c")) // as set in CreateShootSap
		})

		By("And Then Scope has status.exposedData.natGatewayIps", func() {
			expected := pie.Map(sapCreatedInfra.Router.GatewayInfo.ExternalFixedIPs, func(x routers.ExternalFixedIP) string {
				return x.IPAddress
			})
			expected = pie.Sort(pie.Unique(pie.Filter(expected, func(s string) bool {
				return s != ""
			})))
			Expect(scope.Status.ExposedData.NatGatewayIps).To(HaveLen(len(expected)))
			Expect(scope.Status.ExposedData.NatGatewayIps).To(ConsistOf(expected))
		})

		infoConfigMap := &corev1.ConfigMap{}

		By("And Then SKR kyma-info configmap exists", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), infoConfigMap, NewObjActions(WithNamespace("kyma-system"), WithName("kyma-info"))).
				Should(Succeed())
		})

		By("And Then SKR kyma-info configmap contains natGatewayIps", func() {
			Expect("cloud.natGatewayIps").To(BeKeyOf(infoConfigMap.Data))
			Expect(infoConfigMap.Data["cloud.natGatewayIps"]).To(Equal(pie.Join(scope.Status.ExposedData.NatGatewayIps, ", ")))
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

		By("And Then Kyma Network has OpenStack reference details", func() {
			Expect(kymaNetwork.Spec.Network.Reference).NotTo(BeNil())
			Expect(kymaNetwork.Spec.Network.Reference.OpenStack).NotTo(BeNil())
			Expect(kymaNetwork.Spec.Network.Reference.OpenStack.NetworkName).To(Equal(scope.Spec.Scope.OpenStack.VpcNetwork))
			Expect(kymaNetwork.Spec.Network.Reference.OpenStack.Domain).To(Equal(scope.Spec.Scope.OpenStack.DomainName))
			Expect(kymaNetwork.Spec.Network.Reference.OpenStack.Project).To(Equal(scope.Spec.Scope.OpenStack.TenantName))
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
