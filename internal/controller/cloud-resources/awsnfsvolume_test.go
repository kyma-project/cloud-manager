package cloudresources

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Created SKR AwsNfsVolume is projected into KCP and it gets Ready condition and PV created", Ordered, func() {

	Context("Given SKR Cluster", Ordered, func() {

		It("And Given SKR namespace exists", func() {
			Eventually(CreateNamespace).
				WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
				Should(Succeed())
		})

		skrIpRangeName := "aws-nfs-iprange-1"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		skrIpRangeId := "db018167-dd48-4d8c-aa3c-ea9e2ed05307"

		It("And Given SKR IpRange exists", func() {

			// tell skriprange reconciler to ignore this SKR IpRange
			skriprange.Ignore.AddName(skrIpRangeName)

			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				).
				Should(Succeed())
		})

		It("And SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusCidr(skrIpRange.Spec.Cidr),
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		awsNfsVolume := &cloudresourcesv1beta1.AwsNfsVolume{}
		awsNfsVolumeCapacity := "100G"

		It("When AwsNfsVolume is created", func() {
			Eventually(CreateAwsNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsNfsVolume,
					WithName("aws-nfs-volume-1"),
					WithNfsVolumeIpRange(skrIpRange.Name),
					WithAwsNfsVolumeCapacity(awsNfsVolumeCapacity),
				).
				Should(Succeed())
		})

		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}

		It("Then KCP NfsInstance is created", func() {
			// load SKR AwsNfsVolume to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsNfsVolume,
					NewObjActions(),
					AssertAwsNfsVolumeHasId(),
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

		It("When KCP NfsInstance gets Ready condition", func() {
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

		It("Then SKR AwsNfsVolume will get Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsNfsVolume,
					NewObjActions(),
					AssertHasConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		pv := &corev1.PersistentVolume{}
		It("And Then SKR PersistentVolume will be created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					pv,
					NewObjActions(
						WithName(awsNfsVolume.Name),
					),
				).
				Should(Succeed())
		})

	})

})
