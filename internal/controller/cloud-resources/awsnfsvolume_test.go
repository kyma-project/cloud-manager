package cloudresources

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Created SKR AwsNfsVolume is projected into KCP and it gets Ready condition and PV created", func() {

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

		It("When AwsNfsVolume is created", func() {
			Eventually(CreateAwsNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsNfsVolume,
					WithName("aws-nfs-volume-1"),
					WithNfsVolumeIpRange(skrIpRange.Name),
				).
				Should(Succeed())
		})

		//kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		//
		//It("Then KCP NfsInstance is created", func() {
		//	Eventually(LoadAndCheck).
		//		WithArguments(
		//			infra.Ctx(),
		//			infra.KCP().Client(),
		//			kcpNfsInstance,
		//			WithName(skrIpRange.Status.Id),
		//		).
		//		Should(Succeed())
		//})
		//
		//It("When KCP NfsInstance gets Ready condition", func() {
		//	Eventually(UpdateStatus).
		//		WithArguments(
		//			infra.Ctx(),
		//			infra.KCP().Client(),
		//			kcpNfsInstance,
		//			WithNfsInstanceStatusHost(DefaultNfsInstanceHost),
		//			WithConditions(KcpReadyCondition()),
		//		).
		//		Should(Succeed())
		//})
		//
		//It("Then SKR AwsNfsVolume will get Ready condition", func() {
		//	Eventually(LoadAndCheck).
		//		WithArguments(
		//			infra.Ctx(),
		//			infra.SKR().Client(),
		//			awsNfsVolume,
		//			NewObjActions(),
		//			AssertHasConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
		//		).
		//		Should(Succeed())
		//})

	})

})
