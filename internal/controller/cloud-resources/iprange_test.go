package cloudresources

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Created SKR IpRange gets projects into KCP", func() {

	Context("Given SKR cluster", Ordered, func() {

		It("And Given SKR namespace exists", func() {
			Eventually(CreateNamespace).
				WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
				Should(Succeed())
		})

		skrIpRangeName := "aws-nfs-iprange-1"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		It("When SKR IpRange is created", func() {
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				).
				Should(Succeed())
		})

		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		It("Then KCP IpRange is created", func() {
			// load SKR IpRange to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					skrIpRange,
					NewObjActions(),
					AssertSkrIpRangeHasId(),
				).
				Should(Succeed())

			// check KCP IpRange is created with name=skrIpRange.ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpIpRange,
					NewObjActions(WithName(skrIpRange.Status.Id)),
				).
				Should(Succeed())
		})

		It("And KCP IpRange should be properly created", func() {
			By("KCP IpRange should have label cloud-manager.kyma-project.io/kymaName")
			Expect(kcpIpRange.Labels[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))

			By("KCP IpRange should have label cloud-manager.kyma-project.io/remoteName")
			Expect(kcpIpRange.Labels[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(skrIpRange.Name))

			By("KCP IpRange should have label cloud-manager.kyma-project.io/remoteNamespace")
			Expect(kcpIpRange.Labels[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(skrIpRange.Namespace))

			By("KCP IpRange should have spec.scope.name equal to SKR Cluster kyma name")
			Expect(kcpIpRange.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))

			By("KCP IpRange should have spec.cidr equal to SKR IpRange cidr")
			Expect(kcpIpRange.Spec.Cidr).To(Equal(skrIpRange.Spec.Cidr))

			By("KCP IpRange should have spec.remoteRef matching to to SKR IpRange")
			Expect(kcpIpRange.Spec.RemoteRef.Namespace).To(Equal(skrIpRange.Namespace))
			Expect(kcpIpRange.Spec.RemoteRef.Name).To(Equal(skrIpRange.Name))
		})

		It("When KCP IpRange gets Ready Condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpIpRange,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		It("Then SKR IpRange will get Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					skrIpRange,
					NewObjActions(),
					AssertHasConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		It("And Then SKR IpRange will have finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(skrIpRange, cloudresourcesv1beta1.Finalizer))
		})

		It("And Then SKR IpRange will have spec.cidr copy in status", func() {
			Expect(skrIpRange.Status.Cidr).To(Equal(skrIpRange.Spec.Cidr))
		})

		It("And Then SKR IpRange will not have Error condition", func() {
			Expect(meta.FindStatusCondition(skrIpRange.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)).To(BeNil())
		})

	})

	//It("IpRange lifecycle", func() {
	//
	//	Expect(infra.SKR().GivenNamespaceExists(ipRangeNamespace)).
	//		NotTo(HaveOccurred())
	//
	//	// When SKR IpRange is created
	//	skrIpRange := &cloudresourcesv1beta1.IpRange{
	//		ObjectMeta: metav1.ObjectMeta{
	//			Namespace: ipRangeNamespace,
	//			Name:      ipRangeName,
	//		},
	//		Spec: cloudresourcesv1beta1.IpRangeSpec{
	//			Cidr: "10.181.0.0/16",
	//		},
	//	}
	//	Expect(infra.SKR().Client().Create(infra.Ctx(), skrIpRange)).
	//		NotTo(HaveOccurred())
	//
	//	// Then Kcp IpRage will get created
	//	kcpIpRangeList := &cloudcontrolv1beta1.IpRangeList{}
	//	Eventually(func() (exists bool, err error) {
	//		err = infra.KCP().Client().List(
	//			infra.Ctx(),
	//			kcpIpRangeList,
	//			client.MatchingLabels{
	//				cloudcontrolv1beta1.LabelKymaName:        infra.SkrKymaRef().Name,
	//				cloudcontrolv1beta1.LabelRemoteName:      skrIpRange.Name,
	//				cloudcontrolv1beta1.LabelRemoteNamespace: skrIpRange.Namespace,
	//			},
	//		)
	//		exists = len(kcpIpRangeList.Items) > 0
	//		return
	//	}, timeout, interval).
	//		Should(BeTrue(), "expected KCP IpRange to be created")
	//	Expect(kcpIpRangeList.Items).
	//		To(HaveLen(1))
	//
	//	// Then Assert KCP IpRange =====================================================
	//	kcpIpRange := &kcpIpRangeList.Items[0]
	//	// has labels
	//	Expect(kcpIpRange.Labels).To(HaveKeyWithValue(cloudcontrolv1beta1.LabelKymaName, infra.SkrKymaRef().Name))
	//	Expect(kcpIpRange.Labels).To(HaveKeyWithValue(cloudcontrolv1beta1.LabelRemoteName, skrIpRange.Name))
	//	Expect(kcpIpRange.Labels).To(HaveKeyWithValue(cloudcontrolv1beta1.LabelRemoteNamespace, skrIpRange.Namespace))
	//
	//	// has scope specified with name equal to SKR's kymaRef
	//	Expect(kcpIpRange.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))
	//
	//	// has cidr same as in SKR
	//	Expect(kcpIpRange.Spec.Cidr).To(Equal(skrIpRange.Spec.Cidr))
	//
	//	// has remote ref set with SKR IpRange name
	//	Expect(kcpIpRange.Spec.RemoteRef.Name).To(Equal(skrIpRange.Name))
	//	Expect(kcpIpRange.Spec.RemoteRef.Namespace).To(Equal(skrIpRange.Namespace))
	//
	//	// Then Assert SKR IpRange =====================================================
	//	Expect(infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(skrIpRange), skrIpRange)).
	//		NotTo(HaveOccurred())
	//	// has cidr copy in status
	//	Expect(skrIpRange.Status.Cidr).To(Equal(skrIpRange.Spec.Cidr))
	//	// has id in status
	//	Expect(skrIpRange.Status.Id).To(Equal(kcpIpRange.Name))
	//	// has finalizer
	//	Expect(controllerutil.ContainsFinalizer(skrIpRange, cloudresourcesv1beta1.Finalizer))
	//	// does not have ready condition type, since KCP IpRange has none so far
	//	Expect(meta.FindStatusCondition(skrIpRange.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)).To(BeNil())
	//	// does not have error condition type, since KCP IpRange has none so far
	//	Expect(meta.FindStatusCondition(skrIpRange.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)).To(BeNil())
	//})
})
