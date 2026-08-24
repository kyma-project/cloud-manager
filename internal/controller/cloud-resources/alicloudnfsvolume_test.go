package cloudresources

import (
	"github.com/kyma-project/cloud-manager/api"
	"k8s.io/apimachinery/pkg/types"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Feature: SKR AlicloudNfsVolume", func() {

	It("Scenario: SKR AlicloudNfsVolume is created and deleted", func() {

		skrIpRangeName := "alicloud-nfs-iprange-1"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		skrIpRangeId := "8b3a3f4a-58f5-4a3c-9d0f-3b1f9f8e0c11"

		By("And Given SKR IpRange exists", func() {
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

		aliNfsVolumeName := "alicloud-nfs-volume-1"
		aliNfsVolume := &cloudresourcesv1beta1.AlicloudNfsVolume{}
		aliNfsVolumeCapacity := "100G"
		nfsHost := "1a2b3c-abc.nas.aliyuncs.com"

		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: aliNfsVolumeName}))

		const (
			pvName  = "alicloud-nfs-pv-1"
			pvcName = "alicloud-nfs-pvc-1"
		)

		By("When AlicloudNfsVolume is created", func() {
			Eventually(CreateAlicloudNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), aliNfsVolume,
					WithName(aliNfsVolumeName),
					WithIpRange(skrIpRange.Name),
					WithAlicloudNfsVolumeCapacity(aliNfsVolumeCapacity),
					WithAlicloudNfsVolumeStorageType(cloudresourcesv1beta1.AlicloudNfsStorageTypeCapacity),
					WithAlicloudNfsVolumePvName(pvName),
					WithAlicloudNfsVolumePvcName(pvcName),
				).
				Should(Succeed())
		})

		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}

		By("Then KCP NfsInstance is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), aliNfsVolume,
					NewObjActions(),
					HavingFieldSet("status", "id"),
					HavingFieldValue(cloudresourcesv1beta1.StateCreating, "status", "state"),
				).
				Should(Succeed(), "expected SKR AlicloudNfsVolume to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
					NewObjActions(WithName(aliNfsVolume.Status.Id)),
				).
				Should(Succeed())

			By("And has KCP labels referencing the SKR resource")
			Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelKymaName]).To(Equal(skrKymaRef.Name))
			Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(aliNfsVolume.Name))
			Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(aliNfsVolume.Namespace))

			By("And has spec.scope.name equal to SKR Cluster kyma name")
			Expect(kcpNfsInstance.Spec.Scope.Name).To(Equal(skrKymaRef.Name))

			By("And has spec.remoteRef matching the SKR AlicloudNfsVolume")
			Expect(kcpNfsInstance.Spec.RemoteRef.Namespace).To(Equal(aliNfsVolume.Namespace))
			Expect(kcpNfsInstance.Spec.RemoteRef.Name).To(Equal(aliNfsVolume.Name))

			By("And has spec.instance.alicloud populated from SKR spec")
			Expect(kcpNfsInstance.Spec.Instance.Alicloud).NotTo(BeNil())
			Expect(string(kcpNfsInstance.Spec.Instance.Alicloud.StorageType)).To(Equal(string(cloudresourcesv1beta1.AlicloudNfsStorageTypeCapacity)))
			Expect(kcpNfsInstance.Spec.Instance.Alicloud.ProtocolType).To(Equal(cloudcontrolv1beta1.AlicloudProtocolTypeNFS))

			By("And has spec.ipRange.name equal to SKR IpRange.status.id")
			Expect(kcpNfsInstance.Spec.IpRange.Name).To(Equal(skrIpRange.Status.Id))
		})

		By("When KCP NfsInstance has Ready condition with a Host", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
					WithNfsInstanceStatusHost(nfsHost),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR AlicloudNfsVolume has Ready condition and Server from KCP Host", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), aliNfsVolume,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingFieldValue(cloudresourcesv1beta1.StateReady, "status", "state"),
					HavingFieldValue(nfsHost, "status", "server"),
				).
				Should(Succeed())
		})

		pv := &corev1.PersistentVolume{}
		By("And Then SKR PersistentVolume is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), pv,
					NewObjActions(WithName(pvName)),
				).
				Should(Succeed())

			By("And its NFS server is the KCP Host")
			Expect(pv.Spec.NFS).NotTo(BeNil())
			Expect(pv.Spec.NFS.Server).To(Equal(nfsHost))

			By("And it has the cloud-manager finalizer")
			Expect(pv.Finalizers).To(ContainElement(api.CommonFinalizerDeletionHook))
		})

		pvc := &corev1.PersistentVolumeClaim{}
		By("And Then SKR PersistentVolumeClaim is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), pvc,
					NewObjActions(WithName(pvcName), WithNamespace(aliNfsVolume.Namespace)),
				).
				Should(Succeed())

			By("And its .spec.volumeName is the PV name")
			Expect(pvc.Spec.VolumeName).To(Equal(pvName))
		})

		// DELETE

		By("When AlicloudNfsVolume is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), aliNfsVolume).
				Should(Succeed())
		})

		By("Then SKR PersistentVolumeClaim is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), pvc).
				Should(Succeed())
		})

		By("And Then SKR PersistentVolume is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), pv).
				Should(Succeed())
		})

		By("And Then KCP NfsInstance is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNfsInstance).
				Should(Succeed())
		})

		By("And Then SKR AlicloudNfsVolume is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), aliNfsVolume).
				Should(Succeed())
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
			Should(Succeed())
	})

})
