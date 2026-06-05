package cloudresources

import (
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	skrazuremanagedredis "github.com/kyma-project/cloud-manager/pkg/skr/azuremanagedredis"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// One happy-path scenario covers the create→ready→delete loop end-to-end on
// a representative production tier (P2: HA + EnterpriseCluster). The full
// tier-letter → KCP-spec expansion table is exercised by unit tests in
// pkg/skr/azuremanagedredis/util_test.go, so this controller test does not
// run per-tier — that would only re-execute the same reconciler plumbing.
var _ = Describe("Feature: SKR AzureManagedRedis", func() {

	It("Scenario: SKR AzureManagedRedis happy path on P2 tier", func() {
		const (
			tier    = cloudresourcesv1beta1.AzureManagedRedisTierP2
			amrName = "amr-p2-instance"
		)

		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: amrName}))
		const skrIpRangeId = "5c70629f-a13f-4b04-af47-1ab274c1c7p2"
		amr := &cloudresourcesv1beta1.AzureManagedRedis{}
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		By("Given default SKR IpRange does not exist", func() {
			Consistently(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				ShouldNot(Succeed())
		})

		By("When SKR AzureManagedRedis is created", func() {
			Eventually(CreateSkrAzureManagedRedis).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), amr,
					WithName(amrName),
					WithSkrAzureManagedRedisTier(tier),
				).
				Should(Succeed())
		})

		By("Then default SKR IpRange is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("When default SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		kcpAMR := &cloudcontrolv1beta1.AzureManagedRedis{}

		By("Then KCP AzureManagedRedis is created with the resolved tier spec", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), amr,
					NewObjActions(),
					HavingFieldSet("status", "id"),
					HavingFieldValue(cloudresourcesv1beta1.StateCreating, "status", "state"),
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpAMR,
					NewObjActions(WithName(amr.Status.Id)),
				).
				Should(Succeed())

			By("And has annotations referencing the SKR object")
			Expect(kcpAMR.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(skrKymaRef.Name))
			Expect(kcpAMR.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(amr.Name))
			Expect(kcpAMR.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(amr.Namespace))

			By("And has spec.remoteRef matching the SKR object")
			Expect(kcpAMR.Spec.RemoteRef.Namespace).To(Equal(amr.Namespace))
			Expect(kcpAMR.Spec.RemoteRef.Name).To(Equal(amr.Name))

			By("And has spec.vpcNetwork.name = KymaCommonName(kymaRef.Name)")
			Expect(kcpAMR.Spec.VpcNetwork.Name).To(Equal(common.KcpNetworkKymaCommonName(skrKymaRef.Name)))

			By("And has spec.ipRange.name = SKR IpRange status.id")
			Expect(kcpAMR.Spec.IpRange.Name).To(Equal(skrIpRangeId))

			By("And has SKU/HA/ClusteringPolicy expanded from the SKR tier letter")
			expected, err := skrazuremanagedredis.TierToSpec(tier)
			Expect(err).NotTo(HaveOccurred())
			Expect(kcpAMR.Spec.SKU).To(Equal(string(expected.SKU)))
			Expect(kcpAMR.Spec.HighAvailability).To(Equal(expected.HighAvailability))
			Expect(kcpAMR.Spec.ClusteringPolicy).To(Equal(string(expected.ClusteringPolicy)))
		})

		const (
			kcpAMRPrimaryEndpoint = "amr.privatelink.redis.azure.net"
			kcpAMRPort            = int32(10000)
			kcpAMRAuthString      = "fake-amr-auth-secret-key"
		)

		By("When KCP AzureManagedRedis has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpAMR,
					WithKcpAzureManagedRedisStatusPrimaryEndpoint(kcpAMRPrimaryEndpoint),
					WithKcpAzureManagedRedisStatusPort(kcpAMRPort),
					WithKcpAzureManagedRedisStatusAuthString(kcpAMRAuthString),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR AzureManagedRedis has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), amr,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingFieldValue(cloudresourcesv1beta1.StateReady, "status", "state"),
				).
				Should(Succeed())
		})

		authSecret := &corev1.Secret{}
		By("And Then SKR auth Secret is created with primaryEndpoint/port/authString", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), authSecret,
					NewObjActions(
						WithName(amrName),
						WithNamespace(amr.Namespace),
					),
					HavingLabel(cloudresourcesv1beta1.LabelRedisInstanceStatusId, amr.Status.Id),
				).
				Should(Succeed())
			Expect(authSecret.Data).To(HaveKeyWithValue("primaryEndpoint", []byte(kcpAMRPrimaryEndpoint)))
			Expect(authSecret.Data).To(HaveKeyWithValue("host", []byte(kcpAMRPrimaryEndpoint)))
			Expect(authSecret.Data).To(HaveKeyWithValue("port", []byte("10000")))
			Expect(authSecret.Data).To(HaveKeyWithValue("authString", []byte(kcpAMRAuthString)))

			By("And it has defined cloud-manager finalizer")
			Expect(authSecret.Finalizers).To(ContainElement(api.CommonFinalizerDeletionHook))
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), amr).
			Should(Succeed())

		By("// cleanup: delete default SKR IpRange", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
		})
	})
})
