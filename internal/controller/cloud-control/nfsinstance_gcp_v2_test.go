package cloudcontrol

import (
	"time"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP NfsInstance GCP v2", func() {

	// Skip v2 tests when GcpNfsInstanceV2 flag is not enabled
	BeforeEach(func() {
		if !feature.GcpNfsInstanceV2.Value(infra.Ctx()) {
			Skip("Skipping v2 tests when GcpNfsInstanceV2 flag is not enabled")
		}
	})

	It("Scenario: KCP GCP NfsInstance v2 is created and deleted", func() {

		const kymaName = "38bd0764-ba1b-4428-9afa-7736aee31ded"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeGcp).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		kcpIpRangeName := "ffff14c2-0937-43cb-872f-cc5573e7c5b9"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		// Tell IpRange reconciler to ignore this kymaName
		kcpiprange.Ignore.AddName(kcpIpRangeName)
		By("And Given KCP IpRange exists", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithScope(scope.Name),
				).
				Should(Succeed())
		})

		By("And Given KCP IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		nfsInstance := &cloudcontrolv1beta1.NfsInstance{}

		By("When NfsInstance is created", func() {
			Eventually(CreateNfsInstance).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), nfsInstance,
					WithName(kymaName),
					WithRemoteRef("skr-nfs-v2-example"),
					WithIpRange(kcpIpRangeName),
					WithScope(kymaName),
					WithNfsInstanceGcp(scope.Spec.Region),
				).
				Should(Succeed(), "failed creating NfsInstance")
		})

		var filestoreInstance *filestorepb.Instance
		By("Then GCP Filestore instance is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					NewObjActions(),
					HavingNfsInstanceStatusId()).
				Should(Succeed(), "expected NfsInstance to get status.id")

				// Get instance from mock using Status.Id as full key
			mockServer := infra.GcpMock()
			filestoreInstance = mockServer.GetInstanceByName(nfsInstance.Status.Id)
			Expect(filestoreInstance).NotTo(BeNil())
		})

		By("When GCP Filestore instance is Ready", func() {
			mockServer := infra.GcpMock()
			mockServer.SetInstanceReady(nfsInstance.Status.Id)
		})

		By("Then NfsInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).
				Should(Succeed(), "expected NfsInstance to have Ready state, but it didn't")
		})

		By("And Then NfsInstance has .status.host set", func() {
			Expect(len(nfsInstance.Status.Host) > 0).To(Equal(true))
		})

		By("And Then NfsInstance has .status.path set", func() {
			Expect(len(nfsInstance.Status.Path) > 0).To(Equal(true))
		})

		By("And Then NfsInstance has .status.capacity set", func() {
			Expect(nfsInstance.Status.Capacity.IsZero()).To(BeFalse())
		})

		// DELETE

		By("When NfsInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
				Should(Succeed(), "failed deleting NfsInstance")
		})

		By("Then NfsInstance does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
				Should(Succeed(), "expected NfsInstance not to exist (be deleted), but it still exists")
		})
	})

	It("Scenario: KCP GCP NfsInstance v2 is created, updated and deleted", func() {

		const kymaName = "fcd7ef63-d0db-4d31-bedf-18feac25edee"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeGcp).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		kcpIpRangeName := "45f60c1e-a851-4ab7-927b-8d59efc04db4"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		// Tell IpRange reconciler to ignore this kymaName
		kcpiprange.Ignore.AddName(kcpIpRangeName)
		By("And Given KCP IpRange exists", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithScope(scope.Name),
				).
				Should(Succeed())
		})

		By("And Given KCP IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		nfsInstance := &cloudcontrolv1beta1.NfsInstance{}

		By("When NfsInstance is created", func() {
			Eventually(CreateNfsInstance).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), nfsInstance,
					WithName(kymaName),
					WithRemoteRef("skr-nfs-v2-update-example"),
					WithIpRange(kcpIpRangeName),
					WithScope(kymaName),
					WithNfsInstanceGcp(scope.Spec.Region),
				).
				Should(Succeed(), "failed creating NfsInstance")
		})

		var filestoreInstance *filestorepb.Instance
		By("Then GCP Filestore instance is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					NewObjActions(),
					HavingNfsInstanceStatusId()).
				Should(Succeed(), "expected NfsInstance to get status.id")

				// Get instance from mock using Status.Id as full key
			mockServer := infra.GcpMock()
			filestoreInstance = mockServer.GetInstanceByName(nfsInstance.Status.Id)
			Expect(filestoreInstance).NotTo(BeNil())
		})

		By("When GCP Filestore instance is Ready", func() {
			mockServer := infra.GcpMock()
			mockServer.SetInstanceReady(nfsInstance.Status.Id)
		})

		By("Then NfsInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).
				Should(Succeed(), "expected NfsInstance to have Ready state, but it didn't")
		})

		By("And Then NfsInstance has .status.capacity set", func() {
			Expect(nfsInstance.Status.Capacity.IsZero()).To(BeFalse())
		})

		originalCapacity := nfsInstance.Status.CapacityGb

		// UPDATE

		By("When NfsInstance capacity is increased", func() {
			Eventually(UpdateNfsInstance).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), nfsInstance,
				).
				Should(Succeed())
		})

		By("Then NfsInstance still has Ready condition during update", func() {
			Expect(nfsInstance.Status.State).To(Equal(cloudcontrolv1beta1.StateReady))
		})

		By("And GCP Filestore instance reflects capacity change", func() {
			mockServer := infra.GcpMock()

			Eventually(func() int64 {
				instance := mockServer.GetInstanceByName(nfsInstance.Status.Id)
				if instance == nil || len(instance.FileShares) == 0 {
					return 0
				}
				return instance.FileShares[0].CapacityGb
			}).Should(BeNumerically(">", originalCapacity))
		})

		By("And Then NfsInstance status reflects the updated capacity", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					NewObjActions()).
				Should(Succeed())
			Expect(nfsInstance.Status.CapacityGb).To(Equal(2 * originalCapacity))
		})

		// DELETE

		By("When NfsInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
				Should(Succeed(), "failed deleting NfsInstance")
		})

		By("Then NfsInstance does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
				Should(Succeed(), "expected NfsInstance not to exist (be deleted), but it still exists")
		})
	})

})
