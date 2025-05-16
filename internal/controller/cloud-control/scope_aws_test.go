package cloudcontrol

import (
	"fmt"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Feature: KCP Scope", func() {

	It("Scenario: KCP AWS Scope is created and deleted when GardenerCluster CR is created and deleted", func() {
		const (
			kymaName = "5d60be8c-e422-48ff-bd0a-166b0e09dc58"
		)

		kymaNetworkName := common.KcpNetworkKymaCommonName(kymaName)

		shoot := &gardenertypes.Shoot{}

		By("Given Shoot exists", func() {
			Eventually(CreateShootAws).
				WithArguments(infra.Ctx(), infra, shoot, WithName(kymaName)).
				Should(Succeed(), "failed creating garden shoot for AWS")
		})

		awsMock := infra.AwsMock().MockConfigs(infra.AwsMock().GetAccount(), shoot.Spec.Region)

		var awsCreatedInfra *AwsGardenerInfra

		By("An Given AWS infra exists", func() {
			createdInfra, err := CreateAwsGardenerResources(infra.Ctx(), awsMock, shoot.Namespace, shoot.Name, "10.180.0.0/16", "10.180.0.0/16")
			Expect(err).NotTo(HaveOccurred())
			awsCreatedInfra = createdInfra
		})

		gardenerClusterCR := util.NewGardenerClusterUnstructured()

		By("And Given GardenerCluster CR exists", func() {
			Expect(CreateGardenerClusterCR(infra.Ctx(), infra, gardenerClusterCR, kymaName, shoot.Name, cloudcontrolv1beta1.ProviderAws)).
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

		By("And Then Scope provider is AWS", func() {
			Expect(scope.Spec.Provider).To(Equal(cloudcontrolv1beta1.ProviderAws))
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

		By("And Then Scope has spec.scope.aws.accountId", func() {
			Expect(scope.Spec.Scope.Aws).NotTo(BeNil())
			Expect(scope.Spec.Scope.Aws.AccountId).NotTo(BeEmpty())
			Expect(scope.Spec.Scope.Aws.AccountId).To(Equal(infra.AwsMock().GetAccount()))
		})

		By("And Then Scope has vpc network name", func() {
			Expect(scope.Spec.Scope.Aws.VpcNetwork).To(Equal(common.GardenerVpcName(DefaultGardenNamespace, shoot.Name)))
		})

		By("And Then Scope has spec.scope.aws.network.zones as shoot", func() {
			Expect(scope.Spec.Scope.Aws.Network.Zones).To(HaveLen(3))
			Expect(scope.Spec.Scope.Aws.Network.Zones[0].Name).To(Equal("eu-west-1a")) // as set in GivenGardenShootAwsExists
			Expect(scope.Spec.Scope.Aws.Network.Zones[1].Name).To(Equal("eu-west-1b")) // as set in GivenGardenShootAwsExists
			Expect(scope.Spec.Scope.Aws.Network.Zones[2].Name).To(Equal("eu-west-1c")) // as set in GivenGardenShootAwsExists
		})

		By("And Then Scope has status.exposedData.natGatewayIps", func() {
			expected := pie.Map(awsCreatedInfra.NatGateways, func(x *ec2types.NatGateway) string {
				return ptr.Deref(x.NatGatewayAddresses[0].PublicIp, "")
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

		By("And Then Kyma Network has AWS reference details", func() {
			Expect(kymaNetwork.Spec.Network.Reference).NotTo(BeNil())
			Expect(kymaNetwork.Spec.Network.Reference.Aws).NotTo(BeNil())
			Expect(kymaNetwork.Spec.Network.Reference.Aws.NetworkName).To(Equal(scope.Spec.Scope.Aws.VpcNetwork))
			Expect(kymaNetwork.Spec.Network.Reference.Aws.AwsAccountId).To(Equal(scope.Spec.Scope.Aws.AccountId))
			Expect(kymaNetwork.Spec.Network.Reference.Aws.Region).To(Equal(scope.Spec.Region))
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
