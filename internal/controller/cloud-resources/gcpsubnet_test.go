package cloudresources

import (
	"fmt"
	"time"

	"github.com/kyma-project/cloud-manager/api"
	"k8s.io/apimachinery/pkg/api/meta"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/skr/gcprediscluster"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: SKR GcpSubnet", func() {

	It("Scenario: SKR GcpSubnet is created with specified CIDR", func() {

		gcpSubnetName := "skr-a6db5917-390d-4ee2-bea4-93bafbdade96"
		gcpSubnet := &cloudresourcesv1beta1.GcpSubnet{}
		gcpSubnetCidr := "10.252.0.0/24"

		By("When SKR GcpSubnet is created", func() {
			Eventually(CreateSkrGcpSubnet).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpSubnet,
					WithName(gcpSubnetName),
					WithSkrGcpSubnetCidr(gcpSubnetCidr),
				).
				Should(Succeed())
		})

		kcpGcpSubnet := &cloudcontrolv1beta1.GcpSubnet{}

		By("Then KCP GcpSubnet is created", func() {
			// load SKR GcpSubnet to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpSubnet,
					NewObjActions(),
					HavingGcpSubnetStatusId(),
					HavingGcpSubnetStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR GcpSubnet to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpGcpSubnet,
					NewObjActions(
						WithName(gcpSubnet.Status.Id),
					),
				).
				Should(Succeed())

			By("And has annotaton cloud-manager.kyma-project.io/kymaName")
			Expect(kcpGcpSubnet.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))

			By("And has annotaton cloud-manager.kyma-project.io/remoteName")
			Expect(kcpGcpSubnet.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(gcpSubnet.Name))

			By("And has spec.scope.name equal to SKR Cluster kyma name")
			Expect(kcpGcpSubnet.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))

			By("And has spec.remoteRef matching to to SKR GcpSubnet")
			Expect(kcpGcpSubnet.Spec.RemoteRef.Name).To(Equal(gcpSubnet.Name))

			By("And has spec.cidr equal to SKR GcpSubnet.spec.cidr")
			Expect(kcpGcpSubnet.Spec.Cidr).To(Equal(gcpSubnet.Spec.Cidr))

			By("And has spec.purpose set to be PRIVATE")
			Expect(kcpGcpSubnet.Spec.Purpose).To(Equal(cloudcontrolv1beta1.GcpSubnetPurpose_PRIVATE))

			By("And has spec.network.name set to be kyma network name")
			Expect(kcpGcpSubnet.Spec.Network.Name).To(Equal(common.KcpNetworkKymaCommonName(infra.SkrKymaRef().Name)))

		})

		By("When KCP GcpSubnet has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpGcpSubnet,

					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR GcpSubnet has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpSubnet,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingGcpSubnetStatusState(cloudresourcesv1beta1.StateReady),
				).
				Should(Succeed())
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), gcpSubnet).
			Should(Succeed())
	})

	It("Scenario: SKR GcpSubnet is deleted", func() {

		gcpSubnetName := "skr-413fd399-96a4-4ceb-b76c-f61e99ea7028"
		gcpSubnet := &cloudresourcesv1beta1.GcpSubnet{}
		gcpSubnetCidr := "10.254.0.0/24"

		By("And Given GcpSubnet is created", func() {
			Eventually(CreateSkrGcpSubnet).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpSubnet,
					WithName(gcpSubnetName),
					WithSkrGcpSubnetCidr(gcpSubnetCidr),
				).
				Should(Succeed())
		})

		kcpGcpSubnet := &cloudcontrolv1beta1.GcpSubnet{}

		By("And Given KCP GcpSubnet is created", func() {
			// load SKR GcpSubnet to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpSubnet,
					NewObjActions(),
					HavingGcpSubnetStatusId(),
					HavingGcpSubnetStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR GcpSubnet to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpGcpSubnet,
					NewObjActions(
						WithName(gcpSubnet.Status.Id),
					),
				).
				Should(Succeed(), "expected KCP GcpSubnet to be created, but it was not")

			Eventually(Update).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpGcpSubnet, AddFinalizer(api.CommonFinalizerDeletionHook)).
				Should(Succeed(), "failed adding finalizer on KCP GcpSubnet")
		})

		By("And Given KCP GcpSubnet has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpGcpSubnet,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed(), "failed setting KCP GcpSubnet Ready condition")
		})

		By("And Given SKR GcpSubnet has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpSubnet,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingGcpSubnetStatusState(cloudresourcesv1beta1.StateReady),
				).
				Should(Succeed(), "expected GcpSubnet to exist and have Ready condition")
		})

		// DELETE START HERE

		By("When GcpSubnet is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpSubnet).
				Should(Succeed(), "failed deleting GcpSubnet")
		})

		By("Then SKR GcpSubnet has Deleting state", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpSubnet,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.StateDeleting),
					HavingGcpSubnetStatusState(cloudresourcesv1beta1.StateDeleting),
				).
				Should(Succeed(), "expected GcpSubnet to have Deleting state")
		})

		By("And Then KCP GcpSubnet is marked for deletion", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpGcpSubnet, NewObjActions(), HavingDeletionTimestamp()).
				Should(Succeed(), "expected KCP GcpSubnet to be marked for deletion")
		})

		By("When KCP GcpSubnet finalizer is removed and it is deleted", func() {
			Eventually(Update).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpGcpSubnet, RemoveFinalizer(api.CommonFinalizerDeletionHook)).
				Should(Succeed(), "failed removing finalizer on KCP GcpSubnet")
		})

		By("Then SKR GcpSubnet is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpSubnet).
				Should(Succeed(), "expected GcpSubnet not to exist")
		})
	})

	It("Scenario: SKR GcpSubnet can not be deleted if used by GcpRedisCluster", func() {
		const (
			skrGcpSubnetName       = "acc073fe-8a43-41cc-a566-08e99053a6fa"
			skrGcpRedisClusterName = "2a34a935-7877-4456-9e73-963001c67463"
		)
		skrGcpSubnet := &cloudresourcesv1beta1.GcpSubnet{}
		kcpGcpSubnet := &cloudcontrolv1beta1.GcpSubnet{}
		skrGcpRedisCluster := &cloudresourcesv1beta1.GcpRedisCluster{}

		By("Given SKR GcpSubnet exists", func() {
			Eventually(CreateSkrGcpSubnet).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpSubnet,
					WithName(skrGcpSubnetName),
					WithSkrGcpSubnetCidr(addressSpace.MustAllocate(22)),
				).
				Should(Succeed(), "failed creating SKR GcpSubnet")

		})

		By("And Given SKR GcpSubnet has Ready condition", func() {
			// load SKR GcpSubnet to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					skrGcpSubnet,
					NewObjActions(),
					HavingGcpSubnetStatusId(),
				).
				Should(Succeed(), "expected SKR GcpSubnet to get status.id, but it didn't")

			// KCP GcpSubnet is created by manager
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpGcpSubnet,
					NewObjActions(WithName(skrGcpSubnet.Status.Id)),
				).
				Should(Succeed(), "expected SKR GcpSubnet to be created, but none found")

			// When Kcp GcpSubnet gets Ready condition
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpGcpSubnet,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed(), "failed updating KCP GcpSubnet status")

			// And when SKR GcpSubnet gets Ready condition by manager
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					skrGcpSubnet,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed(), "expected SKR GcpSubnet to get Ready condition, but it did not")
		})

		By("And Given GcpRedisCluster exists", func() {
			// tell GcpRedisCluster reconciler to ignore this GcpRedisCluster
			gcprediscluster.Ignore.AddName(skrGcpRedisClusterName)

			Eventually(CreateSkrGcpRedisCluster).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpRedisCluster,
					WithName(skrGcpRedisClusterName),
					WithGcpSubnet(skrGcpSubnet.Name),
					WithSkrGcpRedisClusterDefaultSpec(),
				).
				Should(Succeed(), "failed creating GcpRedisCluster")

			time.Sleep(50 * time.Millisecond)
		})

		By("When SKR GcpSubnet is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet).
				Should(Succeed(), "failed deleting SKR GcpSubnet")
		})

		By("Then SKR GcpSubnet has Warning condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet, NewObjActions(), HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeWarning)).
				Should(Succeed(), "expected SKR GcpSubnet to have Warning condition")
		})

		By("And Then SKR GcpSubnet has DeleteWhileUsed reason", func() {
			cond := meta.FindStatusCondition(skrGcpSubnet.Status.Conditions, cloudresourcesv1beta1.ConditionTypeWarning)
			Expect(cond.Reason).To(Equal(cloudresourcesv1beta1.ConditionTypeDeleteWhileUsed))
			Expect(cond.Message).To(Equal(fmt.Sprintf("Can not be deleted while used by: [%s/%s]", skrGcpRedisCluster.Namespace, skrGcpRedisCluster.Name)))
		})

		By(fmt.Sprintf("// cleanup: delete SKR GcpRedisCluster %s", skrGcpRedisCluster.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpRedisCluster).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR GcpRedisCluster to clean up")
		})

		By(fmt.Sprintf("// cleanup: delete SKR GcpSubnet %s", skrGcpSubnet.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpSubnet).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR GcpSubnet to clean up")
		})

	})

})
