package cloudresources

import (
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolume"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
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
					WithSkrIpRangeSpecCidr("10.7.0.0/24"),
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

		By("And Then SKR IpRange has Ready state", func() {
			Expect(skrIpRange.Status.State).To(Equal(cloudresourcesv1beta1.StateReady))
		})

		By("And Then SKR IpRange has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(skrIpRange, cloudresourcesv1beta1.Finalizer))
		})

		By("And Then SKR IpRange has spec.cidr copy in status", func() {
			Expect(skrIpRange.Status.Cidr).To(Equal(skrIpRange.Spec.Cidr))
		})

		By("And Then SKR IpRange does not have Error condition", func() {
			Expect(meta.FindStatusCondition(skrIpRange.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)).To(BeNil())
		})

		// CleanUp -------------------
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
			Should(Succeed(), "failed deleting SKR IpRange to clean up")
	})

	It("Scenario: SKR IpRange is deleted", func() {

		const (
			skrIpRangeName = "1261ba97-f4f0-4465-a339-f8691aee8c48"
		)
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		By("Given SKR IpRange exists", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
					WithSkrIpRangeSpecCidr("10.7.1.0/24"),
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
					WithSkrIpRangeSpecCidr("10.7.2.0/30"),
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
		})

		// CleanUp -------------------
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
			Should(Succeed(), "failed deleting SKR IpRange to clean up")
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
					WithSkrIpRangeSpecCidr("10.6.0.0/16"),
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
		})

		// CleanUp -------------------
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
			Should(Succeed(), "failed deleting SKR IpRange to clean up")
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
					WithSkrIpRangeSpecCidr("10.7.3.0/31"),
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
			Expect(errCond.Message).To(Equal(fmt.Sprintf("CIDR %s block size must not be greater than 30", skrIpRange.Spec.Cidr)))
		})

		By("And Then SKR IpRange has Error state", func() {
			Expect(skrIpRange.Status.State).To(Equal(cloudresourcesv1beta1.StateError))
		})

		// CleanUp -------------------
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
			Should(Succeed(), "failed deleting SKR IpRange to clean up")
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
					WithSkrIpRangeSpecCidr("10.5.0.0/15"),
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
			Expect(errCond.Message).To(Equal(fmt.Sprintf("CIDR %s block size must not be less than 16", skrIpRange.Spec.Cidr)))
		})

		By("And Then SKR IpRange has Error state", func() {
			Expect(skrIpRange.Status.State).To(Equal(cloudresourcesv1beta1.StateError))
		})

		// CleanUp -------------------
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
			Should(Succeed(), "failed deleting SKR IpRange to clean up")
	})

	It("Scenario: SKR IpRange can not be deleted if used by AwsNfsVolume", func() {
		const (
			skrIpRangeName   = "bb4e456c-7a99-44d3-b387-d044c3f42812"
			awsNfsVolumeName = "3e0b8eee-64e1-402c-8399-619a23b6be0a"
		)
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		skrAwsNfsVolume := &cloudresourcesv1beta1.AwsNfsVolume{}

		By("Given SKR IpRange exists", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
					WithSkrIpRangeSpecCidr("10.7.4.0/24"),
				).
				Should(Succeed(), "failed creating SKR IpRange")

		})

		By("And Given SKR IpRange has Ready condition", func() {
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

		By("And Given AwsNfsVolume exists", func() {
			// tell AwsNfsVolume reconciler to ignore this AwsNfsVolume
			awsnfsvolume.Ignore.AddName(awsNfsVolumeName)

			Eventually(CreateAwsNfsVolume).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolume,
					WithName(awsNfsVolumeName),
					WithNfsVolumeIpRange(skrIpRange.Name),
					WithAwsNfsVolumeCapacity("10G"),
				).
				Should(Succeed(), "failed creating AwsNfsVolume")

			time.Sleep(50 * time.Millisecond)
		})

		By("When SKR IpRange is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed(), "failed deleting SKR IpRange")
		})

		By("Then SKR IpRange has Warning condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange, NewObjActions(), HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeWarning)).
				Should(Succeed(), "expected SKR IpRange to have Warning condition")
		})

		By("And Then SKR IpRange has DeleteWhileUsed reason", func() {
			cond := meta.FindStatusCondition(skrIpRange.Status.Conditions, cloudresourcesv1beta1.ConditionTypeWarning)
			Expect(cond.Reason).To(Equal(cloudresourcesv1beta1.ConditionTypeDeleteWhileUsed))
			Expect(cond.Message).To(Equal(fmt.Sprintf("Can not be deleted while used by: [%s/%s]", skrAwsNfsVolume.Namespace, skrAwsNfsVolume.Name)))
		})

		// CleanUp -------------------
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolume).
			Should(Succeed(), "failed deleting SKR AwsNfsVolume to clean up")
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
			Should(Succeed(), "failed deleting SKR IpRange to clean up")

	})

	It("Scenario: SKR IpRange can not have CIDR overlapping with other IpRange", func() {
		const (
			iprange1Name = "da30273b-4775-40ac-91e4-0784091976f1"
			cidr1        = "10.4.1.0/22"
			iprange2Name = "33c48765-941e-45a7-81cd-1960eab3cd9f"
			cidr2        = "10.4.2.0/25"
		)

		iprange1 := &cloudresourcesv1beta1.IpRange{}
		iprange2 := &cloudresourcesv1beta1.IpRange{}

		By("Given SKR IpRange exists", func() {
			skriprange.Ignore.AddName(iprange1Name)
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), iprange1,
					WithName(iprange1Name),
					WithSkrIpRangeSpecCidr(cidr1),
				).
				Should(Succeed(), "failed creating SKR IpRange1")
		})

		By("When SKR IpRange with overlapping CIDR is created", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), iprange2,
					WithName(iprange2Name),
					WithSkrIpRangeSpecCidr(cidr2),
				).
				Should(Succeed(), "failed creating SKR IpRange2")
		})

		var condErr *metav1.Condition
		By("Then SKR IpRange has Error condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), iprange2, NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeError),
				).
				Should(Succeed(), "expected SKR IpRange to have error condition")
			condErr = meta.FindStatusCondition(*iprange2.Conditions(), cloudresourcesv1beta1.ConditionTypeError)
		})

		By("And Then SKR IpRange Error condition has CidrOverlap reason", func() {
			Expect(condErr.Reason).To(Equal(cloudresourcesv1beta1.ConditionReasonCidrOverlap))
			Expect(condErr.Message).To(Equal(fmt.Sprintf("CIDR overlaps with %s/%s", DefaultSkrNamespace, iprange1Name)))
		})

		// cleanup
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), iprange1).
			Should(Succeed())
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), iprange2).
			Should(Succeed())
	})

	It("Scenario: SKR IpRange can not have same CIDR as other IpRange", func() {
		const (
			iprange1Name = "7ccb2b7a-fbd7-4613-94e9-f7a2c09d243b"
			cidr         = "10.7.5.0/24"
			iprange2Name = "166968b9-8a54-4fee-b268-54b33333487c"
		)

		iprange1 := &cloudresourcesv1beta1.IpRange{}
		iprange2 := &cloudresourcesv1beta1.IpRange{}

		By("Given SKR IpRange exists", func() {
			skriprange.Ignore.AddName(iprange1Name)
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), iprange1,
					WithName(iprange1Name),
					WithSkrIpRangeSpecCidr(cidr),
				).
				Should(Succeed(), "failed creating SKR IpRange1")
		})

		By("When SKR IpRange with overlapping CIDR is created", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), iprange2,
					WithName(iprange2Name),
					WithSkrIpRangeSpecCidr(cidr),
				).
				Should(Succeed(), "failed creating SKR IpRange2")
		})

		var condErr *metav1.Condition
		By("Then SKR IpRange has Error condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), iprange2, NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeError),
				).
				Should(Succeed(), "expected SKR IpRange to have error condition")
			condErr = meta.FindStatusCondition(*iprange2.Conditions(), cloudresourcesv1beta1.ConditionTypeError)
		})

		By("And Then SKR IpRange Error condition has CidrOverlap reason", func() {
			Expect(condErr.Reason).To(Equal(cloudresourcesv1beta1.ConditionReasonCidrOverlap))
			Expect(condErr.Message).To(Equal(fmt.Sprintf("CIDR overlaps with %s/%s", DefaultSkrNamespace, iprange1Name)))
		})

		// cleanup
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), iprange1).
			Should(Succeed())
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), iprange2).
			Should(Succeed())
	})
})
