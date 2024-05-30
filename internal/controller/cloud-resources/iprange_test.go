package cloudresources

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature"
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
		Eventually(CreateNamespace).
			WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
			Should(Succeed())
	})

	// since Describe() calls are random this might execute before features are initialized from static test config
	// so this must be a function that will be called during spec execution, which is run
	// after features are initialized

	shouldSkipForIpRangeAutomaticCidrAllocation := func() string {
		if feature.IpRangeAutomaticCidrAllocation.Value(context.Background()) {
			return ""
		}
		return "IpRangeAutomaticCidrAllocation is disabled"
	}

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
				Expect(controllerutil.ContainsFinalizer(skrIpRange, cloudresourcesv1beta1.Finalizer))
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
					Should(Succeed(), "failed deleting SKR IpRange to clean up")
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
		shouldSkipForIpRangeAutomaticCidrAllocation,
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
	}

	runSkrIpRangeDeleteScenario(
		nil,
		"with specified CIDR",
		"1261ba97-f4f0-4465-a339-f8691aee8c48",
		WithSkrIpRangeSpecCidr(addressSpace.MustAllocate(24)),
	)
	runSkrIpRangeDeleteScenario(
		shouldSkipForIpRangeAutomaticCidrAllocation,
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
				Should(Succeed(), "failed deleting SKR IpRange to clean up")
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
				Should(Succeed(), "failed deleting SKR IpRange to clean up")
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

		By(fmt.Sprintf("// cleanup: delete SKR AwsNfsVolume %s", skrAwsNfsVolume.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolume).
				Should(Succeed(), "failed deleting SKR AwsNfsVolume to clean up")
		})

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", skrIpRange.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed(), "failed deleting SKR IpRange to clean up")
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
			Expect(condErr.Message).To(Equal(fmt.Sprintf("CIDR overlaps with %s/%s", DefaultSkrNamespace, iprange1Name)))
		})

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", iprange1.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), iprange1).
				Should(Succeed())
		})

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", iprange2.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), iprange2).
				Should(Succeed())
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
			Expect(condErr.Message).To(Equal(fmt.Sprintf("CIDR overlaps with %s/%s", DefaultSkrNamespace, iprange1Name)))
		})

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", iprange1.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), iprange1).
				Should(Succeed())
		})

		By(fmt.Sprintf("// cleanup: delete SKR IpRange %s", iprange2.Name), func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), iprange2).
				Should(Succeed())
		})
	})

	It("Scenario: SKR IpRange can be created with empty CIDR", func() {
		if !feature.IpRangeAutomaticCidrAllocation.Value(context.Background()) {
			Skip("IpRangeAutomaticCidrAllocation is disabled")
		}

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

	})
})
