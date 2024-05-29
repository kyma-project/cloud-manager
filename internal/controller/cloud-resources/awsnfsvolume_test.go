package cloudresources

import (
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Feature: SKR AwsNfsVolume", func() {

	It("Scenario: SKR AwsNfsVolume is created", func() {

		skrIpRangeName := "aws-nfs-iprange-1"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		skrIpRangeId := "db018167-dd48-4d8c-aa3c-ea9e2ed05307"

		By("Given SKR namespace exists", func() {
			Eventually(CreateNamespace).
				WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
				Should(Succeed())
		})
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

		awsNfsVolumeName := "aws-nfs-volume-1"
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

		By("Given SKR namespace exists", func() {
			Eventually(CreateNamespace).
				WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
				Should(Succeed())
		})
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

		By("Then SKR PersistentVolume is deleted", func() {
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
})
