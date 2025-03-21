package cloudresources

import (
	"fmt"
	"time"

	"github.com/kyma-project/cloud-manager/api"

	"github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolume"
	"github.com/kyma-project/cloud-manager/pkg/skr/awsrediscluster"
	"github.com/kyma-project/cloud-manager/pkg/skr/awsredisinstance"
	"github.com/kyma-project/cloud-manager/pkg/skr/azureredisinstance"
	"github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolume"
	"github.com/kyma-project/cloud-manager/pkg/skr/gcpredisinstance"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Feature: SKR IpRange", func() {

	const (
		consistentlyDuration = 500 * time.Millisecond
	)

	BeforeEach(func() {
		Eventually(CreateNamespace).
			WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
			Should(Succeed())
	})

	runSkrIpRangeCreateScenario := func(shouldSkip func() string, titleSuffix, skrIpRangeName string, cidrAction ObjAction) {

		It("Scenario: SKR IpRange is created "+titleSuffix, func() {

			if shouldSkip != nil {
				if msg := shouldSkip(); msg != "" {
					Skip(msg)
				}
			}

			skrIpRange := &cloudresourcesv1beta1.IpRange{}

			By("When SKR IpRange is created", func() {
				args := []interface{}{
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				}
				if cidrAction != nil {
					args = append(args, cidrAction)
				}
				Eventually(CreateSkrIpRange).
					WithArguments(args...).
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
				Expect(controllerutil.ContainsFinalizer(skrIpRange, api.CommonFinalizerDeletionHook))
			})

			By("And Then SKR IpRange has spec.cidr copy in status", func() {
				if len(skrIpRange.Status.Cidr) > 0 {
					Expect(skrIpRange.Status.Cidr).To(Equal(skrIpRange.Spec.Cidr))
				}
			})

			By("And Then SKR IpRange does not have Error condition", func() {
				Expect(meta.FindStatusCondition(skrIpRange.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)).To(BeNil())
			})

			By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", skrIpRange.Name), func() {
				Eventually(Delete).
					WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
					Should(SucceedIgnoreNotFound(), "failed deleting SKR IpRange to clean up")
			})
		})
	}

	runSkrIpRangeCreateScenario(
		nil,
		"with specified CIDR",
		"b75c4076-3230-4890-8bd0-c1c84c109675",
		WithSkrIpRangeSpecCidr(addressSpace.MustAllocate(24)),
	)
	runSkrIpRangeCreateScenario(
		nil,
		"with empty CIDR",
		"a7c8cbf9-56f4-4ef2-bc41-98f5c6133ab0",
		nil,
	)

	runSkrIpRangeDeleteScenario := func(shouldSkip func() string, titleSuffix, skrIpRangeName string, cidrAction ObjAction) {

		It("Scenario: SKR IpRange is deleted "+titleSuffix, func() {

			if shouldSkip != nil {
				if msg := shouldSkip(); msg != "" {
					Skip(msg)
				}
			}

			skrIpRange := &cloudresourcesv1beta1.IpRange{}

			By("Given SKR IpRange exists", func() {
				args := []interface{}{
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				}
				if cidrAction != nil {
					args = append(args, cidrAction)
				}
				Eventually(CreateSkrIpRange).
					WithArguments(args...).
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
					WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange, AddFinalizer(api.CommonFinalizerDeletionHook)).
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
					Should(Succeed(), "expected SKR IpRange to get Ready condition, but it did not")
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
					WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange, RemoveFinalizer(api.CommonFinalizerDeletionHook)).
					Should(Succeed(), "failed removing finalizer on KCP IpRange")
			})

			By("Then SKR IpRange is deleted", func() {
				Eventually(IsDeleted).
					WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
					Should(Succeed(), "expected SKR IpRange not to exist")
			})
		})
	}

	runSkrIpRangeDeleteScenario(
		nil,
		"with specified CIDR",
		"1261ba97-f4f0-4465-a339-f8691aee8c48",
		WithSkrIpRangeSpecCidr(addressSpace.MustAllocate(24)),
	)
	runSkrIpRangeDeleteScenario(
		nil,
		"with empty CIDR",
		"a05b3025-0874-455a-a852-80bf4f706192",
		nil,
	)

	It("Scenario: SKR IpRange can be created with max size /30", func() {

		skrIpRangeName := "4fa01092-2527-4b3d-a22a-b0eaf63d3b3e"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		By("When SKR IpRange is created", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
					WithSkrIpRangeSpecCidr(addressSpace.MustAllocate(30)),
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

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", skrIpRange.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR IpRange to clean up")
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
					WithSkrIpRangeSpecCidr(addressSpace.MustAllocate(16)),
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

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", skrIpRange.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR IpRange to clean up")
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
					WithSkrIpRangeSpecCidr(addressSpace.MustAllocate(31)),
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

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", skrIpRange.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR IpRange to clean up")
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
					WithSkrIpRangeSpecCidr("10.128.0.0/15"),
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

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", skrIpRange.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR IpRange to clean up")
		})
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
					WithSkrIpRangeSpecCidr(addressSpace.MustAllocate(24)),
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
				Should(Succeed(), "expected SKR IpRange to get Ready condition, but it did not")
		})

		By("And Given AwsNfsVolume exists", func() {
			// tell AwsNfsVolume reconciler to ignore this AwsNfsVolume
			awsnfsvolume.Ignore.AddName(awsNfsVolumeName)

			Eventually(CreateAwsNfsVolume).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolume,
					WithName(awsNfsVolumeName),
					WithIpRange(skrIpRange.Name),
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

		By(fmt.Sprintf("// cleanup: delete SKR AwsNfsVolume %s", skrAwsNfsVolume.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolume).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR AwsNfsVolume to clean up")
		})

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", skrIpRange.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR IpRange to clean up")
		})

	})

	It("Scenario: SKR IpRange can not be deleted if used by GcpNfsVolume", func() {
		const (
			skrIpRangeName   = "9f024c36-4ea8-420a-a297-33bf93a01da9"
			gcpNfsVolumeName = "8dfb5b8b-297b-4ebb-b813-08bebbf42976"
		)
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		skrGcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}

		By("Given SKR IpRange exists", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
					WithSkrIpRangeSpecCidr(addressSpace.MustAllocate(24)),
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
				Should(Succeed(), "expected SKR IpRange to get Ready condition, but it did not")
		})

		By("And Given GcpNfsVolume exists", func() {
			// tell GcpNfsVolume reconciler to ignore this GcpNfsVolume
			gcpnfsvolume.Ignore.AddName(gcpNfsVolumeName)

			Eventually(CreateGcpNfsVolume).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpNfsVolume,
					WithName(gcpNfsVolumeName),
					WithIpRange(skrIpRange.Name),
				).
				Should(Succeed(), "failed creating GcpNfsVolume")

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
			Expect(cond.Message).To(Equal(fmt.Sprintf("Can not be deleted while used by: [%s/%s]", skrGcpNfsVolume.Namespace, skrGcpNfsVolume.Name)))
		})

		By(fmt.Sprintf("// cleanup: delete SKR GcpNfsVolume %s", skrGcpNfsVolume.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpNfsVolume).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR GcpNfsVolume to clean up")
		})

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", skrIpRange.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR IpRange to clean up")
		})

	})

	It("Scenario: SKR IpRange can not have CIDR overlapping with other IpRange", func() {
		const (
			iprange1Name = "da30273b-4775-40ac-91e4-0784091976f1"
			cidr1        = "10.130.1.0/22"
			iprange2Name = "33c48765-941e-45a7-81cd-1960eab3cd9f"
			cidr2        = "10.130.2.0/25"
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
			time.Sleep(time.Second) // must sleep 1s so first IpRange is older
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
			Expect(condErr.Message).To(Equal(fmt.Sprintf("CIDR overlaps with %s", iprange1Name)))
		})

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", iprange1.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), iprange1).
				Should(SucceedIgnoreNotFound())
		})

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", iprange2.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), iprange2).
				Should(SucceedIgnoreNotFound())
		})
	})

	It("Scenario: SKR IpRange can not have same CIDR as other IpRange", func() {
		iprange1Name := "7ccb2b7a-fbd7-4613-94e9-f7a2c09d243b"
		cidr := addressSpace.MustAllocate(24)
		iprange2Name := "166968b9-8a54-4fee-b268-54b33333487c"

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
			time.Sleep(time.Second) // must sleep 1s so first IpRange is older
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
			Expect(condErr.Message).To(Equal(fmt.Sprintf("CIDR overlaps with %s", iprange1Name)))
		})

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", iprange1.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), iprange1).
				Should(SucceedIgnoreNotFound())
		})

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", iprange2.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), iprange2).
				Should(SucceedIgnoreNotFound())
		})
	})

	It("Scenario: SKR IpRange can be created with empty CIDR", func() {
		skrIpRangeName := "db19ab47-8361-448d-a841-227b686982e8"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		By("When SKR IpRange is created with empty spec.cidr", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				).
				Should(Succeed(), "failed creating SKR IpRange")
		})

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

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", skrIpRangeName), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(SucceedIgnoreNotFound())
		})

	})

	It("Scenario: SKR IpRange can not be deleted if used by AwsRedisInstance", func() {
		const (
			skrIpRangeName = "de648eb4-bab8-484a-a13f-fd8258787de3"
			awsRedisName   = "b1411011-ee5e-4ddd-ba7c-b20509e81bf6"
		)
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		skrAwsRedisInstance := &cloudresourcesv1beta1.AwsRedisInstance{}

		By("Given SKR IpRange exists", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
					WithSkrIpRangeSpecCidr(addressSpace.MustAllocate(24)),
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
				Should(Succeed(), "expected SKR IpRange to get Ready condition, but it did not")
		})

		By("And Given AwsRedisInstance exists", func() {
			// tell AwsRedisInstance reconciler to ignore this AwsRedisInstance
			awsredisinstance.Ignore.AddName(awsRedisName)

			Eventually(CreateAwsRedisInstance).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsRedisInstance,
					WithName(awsRedisName),
					WithIpRange(skrIpRange.Name),
					WithAwsRedisInstanceDefautSpecs(),
				).
				Should(Succeed(), "failed creating AwsRedisInstance")

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
			Expect(cond.Message).To(Equal(fmt.Sprintf("Can not be deleted while used by: [%s/%s]", skrAwsRedisInstance.Namespace, skrAwsRedisInstance.Name)))
		})

		By(fmt.Sprintf("// cleanup: delete SKR AwsRedisInstance %s", skrAwsRedisInstance.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsRedisInstance).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR AwsRedisInstance to clean up")
		})

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", skrIpRange.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR IpRange to clean up")
		})

	})

	It("Scenario: SKR IpRange can not be deleted if used by AwsRedisCluster", func() {
		const (
			skrIpRangeName = "a652e689-5214-441a-906f-7e01a8761197"
			awsRedisName   = "5219f064-d52e-4d8a-83c2-c3056f8a0b17"
		)
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		skrAwsRedisCluster := &cloudresourcesv1beta1.AwsRedisCluster{}

		By("Given SKR IpRange exists", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
					WithSkrIpRangeSpecCidr(addressSpace.MustAllocate(24)),
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
				Should(Succeed(), "expected SKR IpRange to get Ready condition, but it did not")
		})

		By("And Given AwsRedisCluster exists", func() {
			// tell AwsRedisCluster reconciler to ignore this AwsRedisCluster
			awsrediscluster.Ignore.AddName(awsRedisName)

			Eventually(CreateAwsRedisCluster).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsRedisCluster,
					WithName(awsRedisName),
					WithIpRange(skrIpRange.Name),
					WithAwsRedisClusterDefautSpecs(),
				).
				Should(Succeed(), "failed creating AwsRedisCluster")

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
			Expect(cond.Message).To(Equal(fmt.Sprintf("Can not be deleted while used by: [%s/%s]", skrAwsRedisCluster.Namespace, skrAwsRedisCluster.Name)))
		})

		By(fmt.Sprintf("// cleanup: delete SKR AwsRedisCluster %s", skrAwsRedisCluster.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsRedisCluster).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR AwsRedisCluster to clean up")
		})

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", skrIpRange.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR IpRange to clean up")
		})

	})

	It("Scenario: SKR IpRange can not be deleted if used by GcpRedisInstance", func() {
		const (
			skrIpRangeName = "d776bc88-801e-4125-a549-8ecf6a2b6f23"
			gcpRedisName   = "95d238db-a29a-4285-ae6c-041efd156f8a"
		)
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		skrGcpRedisInstance := &cloudresourcesv1beta1.GcpRedisInstance{}

		By("Given SKR IpRange exists", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
					WithSkrIpRangeSpecCidr(addressSpace.MustAllocate(24)),
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
				Should(Succeed(), "expected SKR IpRange to get Ready condition, but it did not")
		})

		By("And Given GcpRedisInstance exists", func() {
			// tell GcpRedisInstance reconciler to ignore this GcpRedisInstance
			gcpredisinstance.Ignore.AddName(gcpRedisName)

			Eventually(CreateGcpRedisInstance).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpRedisInstance,
					WithName(gcpRedisName),
					WithIpRange(skrIpRange.Name),
					WithGcpRedisInstanceDefaultSpec(),
				).
				Should(Succeed(), "failed creating GcpRedisInstance")

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
			Expect(cond.Message).To(Equal(fmt.Sprintf("Can not be deleted while used by: [%s/%s]", skrGcpRedisInstance.Namespace, skrGcpRedisInstance.Name)))
		})

		By(fmt.Sprintf("// cleanup: delete SKR GcpRedisInstance %s", skrGcpRedisInstance.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrGcpRedisInstance).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR GcpRedisInstance to clean up")
		})

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", skrIpRange.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR IpRange to clean up")
		})

	})

	It("Scenario: SKR IpRange can not be deleted if used by AzureRedisInstance", func() {
		const (
			skrIpRangeName = "de648eb4-bab8-484a-a13f-fd8258787daz"
			azureRedisName = "b1411011-ee5e-4ddd-ba7c-b20509e81baz"
		)
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		skrAzureRedisInstance := &cloudresourcesv1beta1.AzureRedisInstance{}

		By("Given SKR IpRange exists", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
					WithSkrIpRangeSpecCidr(addressSpace.MustAllocate(24)),
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
				Should(Succeed(), "expected SKR IpRange to get Ready condition, but it did not")
		})

		By("And Given AzureRedisInstance exists", func() {
			// tell AzureRedisInstance reconciler to ignore this AzureRedisInstance
			azureredisinstance.Ignore.AddName(azureRedisName)

			Eventually(CreateAzureRedisInstance).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAzureRedisInstance,
					WithName(azureRedisName),
					WithIpRange(skrIpRange.Name),
					WithAzureRedisInstanceDefaultSpecs(),
				).
				Should(Succeed(), "failed creating AzureRedisInstance")

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
			Expect(cond.Message).To(Equal(fmt.Sprintf("Can not be deleted while used by: [%s/%s]", skrAzureRedisInstance.Namespace, skrAzureRedisInstance.Name)))
		})

		By(fmt.Sprintf("// cleanup: delete SKR AzureRedisInstance %s", skrAzureRedisInstance.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAzureRedisInstance).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR AzureRedisInstance to clean up")
		})

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", skrIpRange.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(SucceedIgnoreNotFound(), "failed deleting SKR IpRange to clean up")
		})

	})
})
