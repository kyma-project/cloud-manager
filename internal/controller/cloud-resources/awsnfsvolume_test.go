package cloudresources

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
)

var _ = Describe("Feature: SKR AwsNfsVolume", func() {

	It("Scenario: SKR AwsNfsVolume is created with specified IpRange", func() {

		skrIpRangeName := "3ef3cbbc-b347-4762-b63a-c1ec9555be65"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		skrIpRangeId := "db018167-dd48-4d8c-aa3c-ea9e2ed05307"

		By("And Given SKR IpRange exists", func() {
			// tell skriprange reconciler to ignore this SKR IpRange
			skriprange.Ignore.AddName(skrIpRangeName)

			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				).
				Should(Succeed())
		})
		By("And Given SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusCidr(skrIpRange.Spec.Cidr),
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		awsNfsVolumeName := "b0fe166e-917c-4dd0-8bb3-978190b6661d"
		awsNfsVolume := &cloudresourcesv1beta1.AwsNfsVolume{}
		awsNfsVolumeCapacity := "100G"

		const (
			pvName = "4e0a550e-a247-44b1-8232-cb973ba053b3"
		)
		pvLabels := map[string]string{
			"foo": "1",
		}
		pvAnnotations := map[string]string{
			"bar": "2",
		}

		By("When AwsNfsVolume is created", func() {
			Eventually(CreateAwsNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsNfsVolume,
					WithName(awsNfsVolumeName),
					WithNfsVolumeIpRange(skrIpRange.Name),
					WithAwsNfsVolumeCapacity(awsNfsVolumeCapacity),
					WithAwsNfsVolumePvName(pvName),
					WithAwsNfsVolumePvLabels(pvLabels),
					WithAwsNfsVolumePvAnnotations(pvAnnotations),
				).
				Should(Succeed())
		})

		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}

		By("Then KCP NfsInstance is created", func() {
			// load SKR AwsNfsVolume to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsNfsVolume,
					NewObjActions(),
					HavingAwsNfsVolumeStatusId(),
					HavingAwsNfsVolumeStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR AwsNfsVolume to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpNfsInstance,
					NewObjActions(
						WithName(awsNfsVolume.Status.Id),
					),
				).
				Should(Succeed())

			By("And has label cloud-manager.kyma-project.io/kymaName")
			Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))

			By("And has label cloud-manager.kyma-project.io/remoteName")
			Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(awsNfsVolume.Name))

			By("And has label cloud-manager.kyma-project.io/remoteNamespace")
			Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(awsNfsVolume.Namespace))

			By("And has spec.scope.name equal to SKR Cluster kyma name")
			Expect(kcpNfsInstance.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))

			By("And has spec.remoteRef matching to to SKR IpRange")
			Expect(kcpNfsInstance.Spec.RemoteRef.Namespace).To(Equal(awsNfsVolume.Namespace))
			Expect(kcpNfsInstance.Spec.RemoteRef.Name).To(Equal(awsNfsVolume.Name))

			By("And has spec.instance.aws equal to SKR AwsNfsVolume.spec values")
			Expect(string(kcpNfsInstance.Spec.Instance.Aws.Throughput)).To(Equal(string(awsNfsVolume.Spec.Throughput)))
			Expect(string(kcpNfsInstance.Spec.Instance.Aws.PerformanceMode)).To(Equal(string(awsNfsVolume.Spec.PerformanceMode)))

			By("And has spec.ipRange.name equal to SKR IpRange.status.id")
			Expect(kcpNfsInstance.Spec.IpRange.Name).To(Equal(skrIpRange.Status.Id))
		})

		By("When KCP NfsInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpNfsInstance,
					WithNfsInstanceStatusHost(""),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR AwsNfsVolume has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsNfsVolume,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAwsNfsVolumeStatusState(cloudresourcesv1beta1.StateReady),
				).
				Should(Succeed())
		})

		pv := &corev1.PersistentVolume{}
		By("And Then SKR PersistentVolume is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					pv,
					NewObjActions(
						WithName(pvName),
					),
				).
				Should(Succeed())

			for k, v := range pvLabels {
				Expect(pv.Labels).To(HaveKeyWithValue(k, v), fmt.Sprintf("expected PV to have label %s=%s", k, v))
			}
			for k, v := range pvAnnotations {
				Expect(pv.Annotations).To(HaveKeyWithValue(k, v), fmt.Sprintf("expected PV to have annotation %s=%s", k, v))
			}
		})

		pvc := &corev1.PersistentVolumeClaim{}
		By("And Then SKR PersistentVolumeClaim is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					pvc,
					NewObjActions(
						WithName(pvName),
						WithNamespace(awsNfsVolume.Namespace),
					),
				).
				Should(Succeed())

			By("And its .spec.volumeName is PV name", func() {
				Expect(pvc.Spec.VolumeName).To(Equal(pv.GetName()))
			})

			By("And it has same Name as the PV", func() {
				Expect(pvc.GetName()).To(Equal(pv.GetName()))
			})

			By("And it has defined label for capacity", func() {
				Expect(pvc.Labels[cloudresourcesv1beta1.LabelStorageCapacity]).ToNot(BeNil())
			})

			// for k, v := range pvLabels {
			// 	Expect(pvc.Labels).To(HaveKeyWithValue(k, v), fmt.Sprintf("expected PVC to have label %s=%s", k, v))
			// }
			// Expect(pvc.Labels).To(HaveKeyWithValue(cloudresourcesv1beta1.LabelStorageCapacity, awsNfsVolume.Spec.Capacity.String()),
			// 	fmt.Sprintf("expected PVC to have label %s=%s", cloudresourcesv1beta1.LabelStorageCapacity, awsNfsVolume.Spec.Capacity.String()))
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolume).
			Should(Succeed())
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
			Should(Succeed())
	})

	It("Scenario: SKR AwsNfsVolume is deleted", func() {

		skrIpRangeName := "aws-nfs-iprange-2"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		skrIpRangeId := "034fd495-3222-465c-8dc9-4617f7ff0013"

		By("And Given SKR IpRange exists", func() {
			// tell skriprange reconciler to ignore this SKR IpRange
			skriprange.Ignore.AddName(skrIpRangeName)

			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				).
				Should(Succeed())
		})
		By("And Given SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusCidr(skrIpRange.Spec.Cidr),
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		awsNfsVolumeName := "aws-nfs-volume-2"
		awsNfsVolume := &cloudresourcesv1beta1.AwsNfsVolume{}
		awsNfsVolumeCapacity := "100G"

		By("And Given AwsNfsVolume is created", func() {
			Eventually(CreateAwsNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsNfsVolume,
					WithName(awsNfsVolumeName),
					WithNfsVolumeIpRange(skrIpRange.Name),
					WithAwsNfsVolumeCapacity(awsNfsVolumeCapacity),
				).
				Should(Succeed())
		})

		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}

		By("And Given KCP NfsInstance is created", func() {
			// load SKR AwsNfsVolume to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsNfsVolume,
					NewObjActions(),
					HavingAwsNfsVolumeStatusId(),
					HavingAwsNfsVolumeStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR AwsNfsVolume to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpNfsInstance,
					NewObjActions(
						WithName(awsNfsVolume.Status.Id),
					),
				).
				Should(Succeed(), "expected KCP AwsNfsInstance to be created, but it was not")

			Eventually(Update).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNfsInstance, AddFinalizer(cloudcontrolv1beta1.FinalizerName)).
				Should(Succeed(), "failed adding finalizer on KCP NfsInstance")
		})

		By("And Given KCP NfsInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpNfsInstance,
					WithNfsInstanceStatusHost(""),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed(), "failed setting KCP NfsInstance Ready condition")
		})

		By("And Given SKR AwsNfsVolume has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsNfsVolume,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAwsNfsVolumeStatusState(cloudresourcesv1beta1.StateReady),
				).
				Should(Succeed(), "expected AwsNfsVolume to exist and have Ready condition")
		})

		pv := &corev1.PersistentVolume{}
		By("And Given SKR PersistentVolume is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					pv,
					NewObjActions(
						WithName(awsNfsVolume.Name),
					),
				).
				Should(Succeed(), "failed creating PV")
		})

		pvc := &corev1.PersistentVolumeClaim{}
		By("And Given SKR PersistentVolumeClaim is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					pvc,
					NewObjActions(
						WithName(pv.Name),
						WithNamespace(awsNfsVolume.Namespace),
					),
				).
				Should(Succeed())
		})

		// DELETE START HERE

		By("When AwsNfsVolume is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolume).
				Should(Succeed(), "failed deleting PV")
		})

		By("Then SKR AwsNfsVolume has Deleting state", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsNfsVolume,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAwsNfsVolumeStatusState(cloudresourcesv1beta1.StateDeleting),
				).
				Should(Succeed(), "expected AwsNfsVolume to have Deleting state")
		})

		By("Then SKR PersistentVolumeClaim is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), pvc).
				Should(Succeed(), "expected PVC not to exist")
		})

		By("And Then SKR PersistentVolume is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), pv).
				Should(Succeed(), "expected PV not to exist")
		})

		By("And Then KCP NfsInstance is marked for deletion", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNfsInstance, NewObjActions(), HavingDeletionTimestamp()).
				Should(Succeed(), "expected KCP NfsInstance to be marked for deletion")
		})

		By("When KCP NfsInstance finalizer is removed and it is deleted", func() {
			Eventually(Update).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNfsInstance, RemoveFinalizer(cloudcontrolv1beta1.FinalizerName)).
				Should(Succeed(), "failed removing finalizer on KCP NfsInstance")
		})

		By("Then SKR AwsNfsVolume is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolume).
				Should(Succeed(), "expected AwsNfsVolume not to exist")
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
			Should(Succeed())
	})

	It("Scenario: SKR AwsNfsVolume is created with empty IpRange when default IpRange does not exist", func() {

		By("Given ff IpRangeAutomaticCidrAllocation is enabled", func() {
			if !feature.IpRangeAutomaticCidrAllocation.Value(context.Background()) {
				Skip("IpRangeAutomaticCidrAllocation is disabled")
			}
		})

		awsNfsVolumeName := "10359994-aed2-4454-bb5f-7c246fa4d9e2"
		skrIpRangeId := "ca4a3f9a-5539-4383-8e0a-3e7c86577a09"
		awsNfsVolume := &cloudresourcesv1beta1.AwsNfsVolume{}
		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		By("Given default SKR IpRange does not exist", func() {
			Consistently(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				ShouldNot(Succeed())
		})

		By("When AwsNfsVolume is created with empty IpRange", func() {
			Eventually(CreateAwsNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsNfsVolume,
					WithName(awsNfsVolumeName),
					WithAwsNfsVolumeCapacity("100G"),
				).
				Should(Succeed())
		})

		By("Then default SKR IpRange is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("And Then default SKR IpRange has label app.kubernetes.io/managed-by: cloud-manager", func() {
			Expect(skrIpRange.Labels["app.kubernetes.io/managed-by"]).To(Equal("cloud-manager"))
		})

		By("And Then default SKR IpRange has label app.kubernetes.io/part-of: kyma", func() {
			Expect(skrIpRange.Labels["app.kubernetes.io/part-of"]).To(Equal("kyma"))
		})

		By("And Then AwsNfsVolume is not ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolume, NewObjActions()).
				Should(Succeed())
			Expect(meta.IsStatusConditionTrue(awsNfsVolume.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)).
				To(BeFalse(), "expected AwsNfsVolume not to have Ready condition, but it has")
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

		By("Then KCP NfsInstance is created", func() {
			// load SKR AwsNfsVolume to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsNfsVolume,
					NewObjActions(),
					HavingAwsNfsVolumeStatusId(),
					HavingAwsNfsVolumeStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR AwsNfsVolume to get status.id and status creating")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpNfsInstance,
					NewObjActions(
						WithName(awsNfsVolume.Status.Id),
					),
				).
				Should(Succeed())
		})

		By("When KCP NfsInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpNfsInstance,
					WithNfsInstanceStatusHost(""),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR AwsNfsVolume has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsNfsVolume,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAwsNfsVolumeStatusState(cloudresourcesv1beta1.StateReady),
				).
				Should(Succeed())
		})

		pv := &corev1.PersistentVolume{}
		By("And Then SKR PersistentVolume is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					pv,
					NewObjActions(WithName(awsNfsVolume.Name)),
				).
				Should(Succeed())
		})

		By("When AwsNfsVolume is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolume).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolume).
				Should(Succeed(), "expected AwsNfsVolume not to exist, but it still does")
		})

		By("Then PersistentVolume does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), pv).
				Should(Succeed(), "expected PersistentVolume not to exist, but it still does")
		})

		By("And Then KCP NfsInstance does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNfsInstance).
				Should(Succeed(), "expected KCP NfsInstance not to exist, but it still does")
		})

		By("And Then SKR default IpRange exists", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange, NewObjActions()).
				Should(Succeed())
		})

		By("// cleanup: delete default SKR IpRange", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
		})
	})

	It("Scenario: SKR AwsNfsVolume is created with empty IpRangeRef when default IpRange already exist", func() {

		By("Given ff IpRangeAutomaticCidrAllocation is enabled", func() {
			if !feature.IpRangeAutomaticCidrAllocation.Value(context.Background()) {
				Skip("IpRangeAutomaticCidrAllocation is disabled")
			}
		})

		awsNfsVolumeName := "50c9827a-76b7-4790-9343-d4ed457a3d25"
		skrIpRangeId := "b42613aa-ece1-4c3d-84fc-a2467fc38cc6"
		awsNfsVolume := &cloudresourcesv1beta1.AwsNfsVolume{}
		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		By("Given default SKR IpRange exists", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange, WithName("default"), WithNamespace("kyma-system")).
				Should(Succeed())
		})

		By("And Given default SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		By("When AwsNfsVolume is created with empty IpRangeRef", func() {
			Eventually(CreateAwsNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsNfsVolume,
					WithName(awsNfsVolumeName),
					WithAwsNfsVolumeCapacity("100G"),
				).
				Should(Succeed())
		})

		By("Then KCP NfsInstance is created", func() {
			// load SKR AwsNfsVolume to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsNfsVolume,
					NewObjActions(),
					HavingAwsNfsVolumeStatusId(),
					HavingAwsNfsVolumeStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR AwsNfsVolume to get status.id and status creating")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpNfsInstance,
					NewObjActions(
						WithName(awsNfsVolume.Status.Id),
					),
				).
				Should(Succeed())
		})

		By("When KCP NfsInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpNfsInstance,
					WithNfsInstanceStatusHost(""),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR AwsNfsVolume has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsNfsVolume,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAwsNfsVolumeStatusState(cloudresourcesv1beta1.StateReady),
				).
				Should(Succeed())
		})

		pv := &corev1.PersistentVolume{}
		By("And Then SKR PersistentVolume is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					pv,
					NewObjActions(WithName(awsNfsVolume.Name)),
				).
				Should(Succeed())
		})

		By("When AwsNfsVolume is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolume).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolume).
				Should(Succeed(), "expected AwsNfsVolume not to exist, but it still does")
		})

		By("Then PersistentVolume does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), pv).
				Should(Succeed(), "expected PersistentVolume not to exist, but it still does")
		})

		By("And Then KCP NfsInstance does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNfsInstance).
				Should(Succeed(), "expected KCP NfsInstance not to exist, but it still does")
		})

		By("And Then SKR default IpRange exists", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange, NewObjActions()).
				Should(Succeed())
		})

		By("// cleanup: delete default SKR IpRange", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
		})
	})

	It("Scenario: AwsNfsVolume IpRangeRef is required when ff IpRangeAutomaticCidrAllocation is disabled", func() {

		By("Given ff IpRangeAutomaticCidrAllocation is disabled", func() {
			if feature.IpRangeAutomaticCidrAllocation.Value(context.Background()) {
				Skip("IpRangeAutomaticCidrAllocation is enabled")
			}
		})

		awsNfsVolumeName := "d67cb9e9-3ac0-4205-be15-866aeedfeddd"
		awsNfsVolume := &cloudresourcesv1beta1.AwsNfsVolume{}

		By("When AwsNfsVolume is created with empty IpRangeRef", func() {
			Eventually(CreateAwsNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsNfsVolume,
					WithName(awsNfsVolumeName),
					WithAwsNfsVolumeCapacity("100G"),
				).
				Should(Succeed())
		})

		By("Then AwsNfsVolume has Error condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsNfsVolume,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeError),
				).
				Should(Succeed())
		})

		By("And Then AwsNfsVolume has Error state", func() {
			Expect(awsNfsVolume.Status.State).To(Equal(cloudresourcesv1beta1.StateError))
		})

		By("And Then AwsNfsVolume Error condition message is: IpRangeRef is required", func() {
			Expect(meta.FindStatusCondition(awsNfsVolume.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError).Message).
				To(Equal("IpRangeRef is required"))
		})
	})

})
