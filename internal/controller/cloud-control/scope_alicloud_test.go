package cloudcontrol

import (
	"fmt"

	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Feature: KCP Scope", func() {

	It("Scenario: KCP Alicloud Scope is created and deleted when GardenerCluster CR is created and deleted", func() {
		const (
			kymaName = "7f3e5b2a-1c4d-4e8f-9a0b-2d6c8e1f3a5b"
		)

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()

		kymaNetworkName := common.KcpNetworkKymaCommonName(kymaName)

		shoot := &gardenertypes.Shoot{}

		By("Given Shoot exists", func() {
			Eventually(CreateShootAlicloud).
				WithArguments(
					infra.Ctx(), infra, shoot,
					alicloudAccount.Credentials().AccessKeyId,
					alicloudAccount.Credentials().AccessKeySecret,
					WithName(kymaName),
				).
				Should(Succeed(), "failed creating garden shoot for Alicloud")
		})

		gardenerClusterCR := util.NewGardenerClusterUnstructured()

		By("And Given GardenerCluster CR exists", func() {
			Expect(CreateGardenerClusterCR(infra.Ctx(), infra, gardenerClusterCR, kymaName, shoot.Name, cloudcontrolv1beta1.ProviderAlicloud)).
				To(Succeed(), "failed creating GardenerCluster CR for Alicloud")
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
				Should(Succeed(), "expected Scope to have Ready condition")
		})

		By("And Then Scope has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(scope, api.CommonFinalizerDeletionHook)).To(BeTrue())
		})

		for _, label := range cloudcontrolv1beta1.ScopeLabels {
			By(fmt.Sprintf("And Then Scope has label %s", label), func() {
				Expect(scope.Labels).To(HaveKey(label))
			})
		}

		By("And Then Scope provider is Alicloud", func() {
			Expect(scope.Spec.Provider).To(Equal(cloudcontrolv1beta1.ProviderAlicloud))
		})

		By("And Then Scope has spec.scope.alicloud not nil", func() {
			Expect(scope.Spec.Scope.Alicloud).NotTo(BeNil())
		})

		By("And Then Scope has spec.scope.aws equal to nil", func() {
			Expect(scope.Spec.Scope.Aws).To(BeNil())
		})

		By("And Then Scope has spec.scope.azure equal to nil", func() {
			Expect(scope.Spec.Scope.Azure).To(BeNil())
		})

		By("And Then Scope has spec.scope.gcp equal to nil", func() {
			Expect(scope.Spec.Scope.Gcp).To(BeNil())
		})

		By("And Then Scope has vpc network name", func() {
			Expect(scope.Spec.Scope.Alicloud.VpcNetwork).To(Equal(common.GardenerVpcName(DefaultGardenNamespace, shoot.Name)))
		})

		By("And Then Scope has spec.scope.alicloud.network.vpc.cidr", func() {
			Expect(scope.Spec.Scope.Alicloud.Network.VPC.CIDR).To(Equal("10.180.0.0/16"))
		})

		By("And Then Scope has spec.scope.alicloud.network.zones", func() {
			Expect(scope.Spec.Scope.Alicloud.Network.Zones).To(HaveLen(1))
			Expect(scope.Spec.Scope.Alicloud.Network.Zones[0].Name).To(Equal("ap-southeast-1a"))
			Expect(scope.Spec.Scope.Alicloud.Network.Zones[0].Workers).To(Equal("10.180.0.0/18"))
		})

		kymaNetwork := &cloudcontrolv1beta1.Network{}

		By("And Then Kyma Network is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kymaNetwork, NewObjActions(WithName(kymaNetworkName))).
				Should(Succeed(), "expected Kyma Network to be created")
		})

		By("And Then Kyma Network has Alicloud reference", func() {
			Expect(kymaNetwork.Spec.Network.Reference).NotTo(BeNil())
			Expect(kymaNetwork.Spec.Network.Reference.Alicloud).NotTo(BeNil())
			Expect(kymaNetwork.Spec.Network.Reference.Alicloud.NetworkName).To(Equal(scope.Spec.Scope.Alicloud.VpcNetwork))
		})

		// DELETE =======================================================

		By("When GardenerCluster CR is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gardenerClusterCR).
				Should(Succeed())
		})

		By("Then Scope does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed(), "expected Scope to be deleted")
		})

		By("And Then Kyma Network does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kymaNetwork).
				Should(Succeed(), "expected Kyma Network to be deleted")
		})

		// CLEANUP =======================================================

		By("// cleanup: delete GardenerCluster", func() {
			Expect(client.IgnoreNotFound(Delete(infra.Ctx(), infra.KCP().Client(), gardenerClusterCR))).To(Succeed())
		})

		By("// cleanup: delete Shoot", func() {
			Expect(Delete(infra.Ctx(), infra.Garden().Client(), shoot)).To(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.Garden().Client(), shoot).
				Should(Succeed())
		})
	})

	It("Scenario: KCP Alicloud Scope is created directly via CreateScopeAlicloud helper", func() {
		const (
			kymaName = "4a2b8c3d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
		)

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(kymaName)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, scope, alicloudAccount.Credentials().AccessKeyId, WithName(kymaName)).
				Should(Succeed())
		})

		By("Then Scope has Alicloud provider", func() {
			Expect(scope.Spec.Provider).To(Equal(cloudcontrolv1beta1.ProviderAlicloud))
		})

		By("And Then Scope has VPC CIDR", func() {
			Expect(scope.Spec.Scope.Alicloud.Network.VPC.CIDR).To(Equal("10.180.0.0/16"))
		})

		By("And Then Scope has zone workers", func() {
			Expect(scope.Spec.Scope.Alicloud.Network.Zones).To(HaveLen(1))
			Expect(scope.Spec.Scope.Alicloud.Network.Zones[0].Workers).To(Equal("10.180.0.0/18"))
		})

		By("// cleanup: delete Scope", func() {
			Expect(client.IgnoreNotFound(infra.KCP().Client().Delete(infra.Ctx(), scope))).To(Succeed())
		})
	})
})
