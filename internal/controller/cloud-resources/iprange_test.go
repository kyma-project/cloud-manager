package cloudresources

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

var _ = Describe("Feature: SKR IpRange", func() {

	const (
		consistentlyDuration = 500 * time.Millisecond
	)

	BeforeEach(func() {
		By("Given SKR namespace exists", func() {
			Eventually(CreateNamespace).
				WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
				Should(Succeed())
		})
	})

	It("Scenario: SKR IpRange is created", func() {
		const (
			skrIpRangeName = "b75c4076-3230-4890-8bd0-c1c84c109675"
		)
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		By("When SKR IpRange is created", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				).
				Should(Succeed(), "failed creating SKR IpRange")
		})

		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		By("Then KCP IpRange is created", func() {
			// load SKR IpRange to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					skrIpRange,
					NewObjActions(),
					AssertSkrIpRangeHasId(),
				).
				Should(Succeed(), "expected SKR IpRange to get status.id, but it didn't")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpIpRange,
					NewObjActions(WithName(skrIpRange.Status.Id)),
				).
				Should(Succeed(), "expected KCP IpRange to exists, but none found")
		})

		By("And Then KCP IpRange has label cloud-manager.kyma-project.io/kymaName", func() {
			Expect(kcpIpRange.Labels[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))
		})

		By("And Then KCP IpRange has label cloud-manager.kyma-project.io/remoteName", func() {
			Expect(kcpIpRange.Labels[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(skrIpRange.Name))
		})

		By("And Then KCP IpRange has label cloud-manager.kyma-project.io/remoteNamespace", func() {
			Expect(kcpIpRange.Labels[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(skrIpRange.Namespace))
		})

		By("And Then KCP IpRange has spec.scope.name equal to SKR Cluster kyma name", func() {
			Expect(kcpIpRange.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))
		})

		By("And Then KCP IpRange has spec.cidr equal to SKR IpRange cidr", func() {
			Expect(kcpIpRange.Spec.Cidr).To(Equal(skrIpRange.Spec.Cidr))
		})

		By("And Then KCP IpRange has spec.remoteRef matching to to SKR IpRange", func() {
			Expect(kcpIpRange.Spec.RemoteRef.Namespace).To(Equal(skrIpRange.Namespace))
			Expect(kcpIpRange.Spec.RemoteRef.Name).To(Equal(skrIpRange.Name))
		})

		By("When KCP IpRange gets Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpIpRange,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed(), "failed updating KCP IpRange status with Ready condition")
		})

		By("Then SKR IpRange has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					skrIpRange,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed(), "expected SKR IpRange to has Ready condition, but it does not")
		})

		By("And Then SKR IpRange has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(skrIpRange, cloudresourcesv1beta1.Finalizer))
		})

		By("And Then SKR IpRange has spec.cidr copy in status", func() {
			Expect(skrIpRange.Status.Cidr).To(Equal(skrIpRange.Spec.Cidr))
		})

		By("And Then SKR IpRange does not have Error condition", func() {
			Expect(meta.FindStatusCondition(skrIpRange.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)).To(BeNil())

			// CleanUp -------------------
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed(), "failed deleting SKR IpRange to clean up")
		})

	})

	It("Scenario: SKR IpRange is deleted", func() {

		By("Given SKR namespace exists", func() {
			Eventually(CreateNamespace).
				WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
				Should(Succeed(), "failed creating SKR namespace")
		})

		const (
			skrIpRangeName = "1261ba97-f4f0-4465-a339-f8691aee8c48"
		)
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		By("And Given SKR IpRange exists", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				).
				Should(Succeed(), "failed creating SKR IpRange")
		})

		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		By("And Given KCP IpRange exists", func() {
			// load SKR IpRange to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					skrIpRange,
					NewObjActions(),
					AssertSkrIpRangeHasId(),
				).
				Should(Succeed(), "expected SKR IpRange to get status.id, but it didn't")

			// KCP IpRange is created by manager
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpIpRange,
					NewObjActions(WithName(skrIpRange.Status.Id)),
				).
				Should(Succeed(), "expected SKR IpRange to be created, but none found")

			Eventually(Update).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange, AddFinalizer(cloudcontrolv1beta1.FinalizerName)).
				Should(Succeed(), "failed adding finalizer to KCP IpRange")

			// When Kcp IpRange gets Ready condition
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpIpRange,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed(), "failed updating KCP IpRange status")

			// And when SKR IpRange gets Ready condition by manager
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					skrIpRange,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed(), "exepcted SKR IpRange to get Ready condition, but it did not")
		})

		By("When SKR IpRange is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed(), "failed deleted SKR IpRange")
		})

		By("Then KCP IpRange is marked for deletion", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange, NewObjActions(), HavingDeletionTimestamp()).
				Should(Succeed(), "expected KCP IpRange to marked for deletion")
		})

		By("When KCP IpRange finalizer is removed and it is deleted", func() {
			Eventually(Update).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange, RemoveFinalizer(cloudcontrolv1beta1.FinalizerName)).
				Should(Succeed(), "failed removing finalizer on KCP IpRange")
		})

		By("Then SKR IpRange is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed(), "expected SKR IpRange not to exist")
		})
	})

	It("Scenario: SKR IpRange can be created with max size /30", func() {

		skrIpRangeName := "4fa01092-2527-4b3d-a22a-b0eaf63d3b3e"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		By("When SKR IpRange is created", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
					WithSkrIpRangeSpecCidr("10.251.0.0/30"),
				).
				Should(Succeed(), "failed creating SKR IpRange")
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					skrIpRange,
					NewObjActions(),
					AssertSkrIpRangeHasId(),
				).
				Should(Succeed(), "expected SKR IpRange to get status.id, but it didn't")
		})

		By("Then KCP IpRange is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpIpRange,
					NewObjActions(WithName(skrIpRange.Status.Id)),
				).
				Should(Succeed(), "expected KCP IpRange to exists, but none found")

			// CleanUp -------------------
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed(), "failed deleting SKR IpRange to clean up")
		})
	})

	It("Scenario: SKR IpRange can be created with min size /16", func() {

		skrIpRangeName := "cdede000-3a4b-4054-bad7-941761282968"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		By("When SKR IpRange is created", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
					WithSkrIpRangeSpecCidr("10.251.0.0/16"),
				).
				Should(Succeed(), "failed creating SKR IpRange")
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					skrIpRange,
					NewObjActions(),
					AssertSkrIpRangeHasId(),
				).
				Should(Succeed(), "expected SKR IpRange to get status.id, but it didn't")
		})

		By("Then KCP IpRange is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpIpRange,
					NewObjActions(WithName(skrIpRange.Status.Id)),
				).
				Should(Succeed(), "expected KCP IpRange to exists, but none found")

			// CleanUp -------------------
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed(), "failed deleting SKR IpRange to clean up")
		})
	})

	It("Scenario: SKR IpRange can not be created with size greater then /30", func() {

		skrIpRangeName := "5851ab1a-d0ef-474e-bb1a-749909d61a4e"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		By("When SKR IpRange is created", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
					WithSkrIpRangeSpecCidr("10.251.0.0/31"),
				).
				Should(Succeed(), "failed creating SKR IpRange")
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					skrIpRange,
					NewObjActions(),
					AssertSkrIpRangeHasId(),
				).
				Should(Succeed(), "expected SKR IpRange to get status.id, but it didn't")
		})

		By("Then KCP IpRange is not created", func() {
			Consistently(LoadAndCheck, consistentlyDuration).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpIpRange,
					NewObjActions(WithName(skrIpRange.Status.Id)),
				).
				ShouldNot(Succeed(), "expected KCP IpRange not to exists, but it is created")
		})

		By("And Then SKR IpRange has Error condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					skrIpRange,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeError),
				).
				Should(Succeed(), "expected SKR IpRange to have error condition, but it has none")
		})

		var errCond *metav1.Condition

		By("And Then SKR IpRange Error condition has InvalidCidr reason", func() {
			errCond = meta.FindStatusCondition(skrIpRange.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
			Expect(errCond).NotTo(BeNil())
			Expect(errCond.Reason).To(Equal(cloudresourcesv1beta1.ConditionReasonInvalidCidr))
		})

		By("And Then SKR IpRange Error condition has message", func() {
			Expect(errCond.Message).To(Equal("CIDR 10.251.0.0/31 block size must not be greater than 30"))
		})
	})

	It("Scenario: SKR IpRange can not be created with size smaller then /16", func() {

		skrIpRangeName := "6256f016-d63f-4e4c-af16-9f3174145389"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		By("When SKR IpRange is created", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
					WithSkrIpRangeSpecCidr("10.251.0.0/15"),
				).
				Should(Succeed(), "failed creating SKR IpRange")
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					skrIpRange,
					NewObjActions(),
					AssertSkrIpRangeHasId(),
				).
				Should(Succeed(), "expected SKR IpRange to get status.id, but it didn't")
		})

		By("Then KCP IpRange is not created", func() {
			Consistently(LoadAndCheck, consistentlyDuration).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpIpRange,
					NewObjActions(WithName(skrIpRange.Status.Id)),
				).
				ShouldNot(Succeed(), "expected KCP IpRange not to exists, but it is created")
		})

		By("And Then SKR IpRange has Error condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					skrIpRange,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeError),
				).
				Should(Succeed(), "expected SKR IpRange to have error condition, but it has none")
		})

		var errCond *metav1.Condition

		By("And Then SKR IpRange Error condition has InvalidCidr reason", func() {
			errCond = meta.FindStatusCondition(skrIpRange.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
			Expect(errCond).NotTo(BeNil())
			Expect(errCond.Reason).To(Equal(cloudresourcesv1beta1.ConditionReasonInvalidCidr))
		})

		By("And Then SKR IpRange Error condition has message", func() {
			Expect(errCond.Message).To(Equal("CIDR 10.251.0.0/15 block size must not be less than 16"))
		})
	})

})
