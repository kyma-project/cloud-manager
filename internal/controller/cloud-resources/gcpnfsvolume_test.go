package cloudresources

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Feature: SKR GcpNfsVolume", func() {

	It("Scenario: GcpNfsVolume is created and deleted", func() {
		name := "6dab3e40-a951-4d84-8dba-ea4551b3e721"
		ipRangeId := "e8ec828d-c9c1-4242-b0e3-5d6525446b22"

		mock := infra.GcpMock2().NewSubscription("gcpnfsvolume1")

		scope := &cloudcontrolv1beta1.Scope{}
		ipRange := &cloudresourcesv1beta1.IpRange{}
		var gcpNfsVolume *cloudresourcesv1beta1.GcpNfsVolume

		pvSpec := &cloudresourcesv1beta1.GcpNfsVolumePvSpec{
			Name: "6dab3e40-a951-4d84-8dba-pv",
			Labels: map[string]string{
				"app": "gcp-nfs",
			},
			Annotations: map[string]string{
				"app/annot": "gcp-nfs-volume-1",
			},
		}
		pvcSpec := &cloudresourcesv1beta1.GcpNfsVolumePvcSpec{
			Name: "6dab3e40-a951-4d84-8dba-pvc",
			Labels: map[string]string{
				"foo": "bar",
			},
			Annotations: map[string]string{
				"baz": "qux",
			},
		}

		By("Given KCP Scope exists", func() {
			err := CreateScopeGcp2(infra.Ctx(), infra, scope, mock.ProjectId(), WithName(name))
			Expect(err).To(Succeed())
			infra.ScopeProvider().Add(scopeprovider.MatchingObj(name, scope))
		})

		By("And Given SKR IpRange exists", func() {
			skriprange.Ignore.AddName(name)
			err := CreateObj(infra.Ctx(), infra.SKR().Client(), ipRange, WithName(name))
			Expect(err).To(Succeed())
		})

		By("And Given SKR IpRange is Ready", func() {
			ipRange.Status.Id = ipRangeId
			ipRange.Status.State = cloudresourcesv1beta1.StateReady
			meta.SetStatusCondition(&ipRange.Status.Conditions, metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeReady,
				Message: cloudresourcesv1beta1.ConditionTypeReady,
			})
			err := composed.PatchObjStatus(infra.Ctx(), ipRange, infra.SKR().Client())
			Expect(err).To(Succeed())
		})

		By("When GcpNfsVolume is created", func() {
			gcpNfsVolume = cloudresourcesv1beta1.NewGcpNfsVolumeBuilder().
				WithName(name).
				WithIpRange(ipRange.Name).
				WithCapacityGb(1024).
				WithTier(cloudresourcesv1beta1.BASIC_HDD).
				WithFileShareName("myfileshare").
				WithPvSpec(pvSpec.Name, pvSpec.Labels, pvSpec.Annotations).
				WithPvcSpec(pvcSpec.Name, pvcSpec.Labels, pvcSpec.Annotations).
				Build()
			err := CreateObj(infra.Ctx(), infra.SKR().Client(), gcpNfsVolume)
			Expect(err).To(Succeed())
		})

		By("Then GcpNfsVolume has finalizer", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolume, NewObjActions(), HavingFinalizer(api.CommonFinalizerDeletionHook)).
				Should(Succeed())
		})

		By("And Then GcpNfsVolume has status.id", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolume, NewObjActions(), HavingFieldSet("status", "id")).
				Should(Succeed())
		})

		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}

		By("And Then KCP NfsInstance is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNfsInstance, NewObjActions(WithName(gcpNfsVolume.Status.Id))).
				Should(Succeed())
		})

		By("And Then KCP NfsInstance has scope set", func() {
			Expect(kcpNfsInstance.Spec.Scope.Name).To(Equal(scope.Name))
		})

		By("And Then KCP NfsInstance has ipRange set", func() {
			Expect(kcpNfsInstance.Spec.IpRange.Name).To(Equal(ipRange.Status.Id))
		})

		By("And Then KCP NfsInstance has remoteRef set", func() {
			Expect(kcpNfsInstance.Spec.RemoteRef.Name).To(Equal(gcpNfsVolume.Name))
			Expect(kcpNfsInstance.Spec.RemoteRef.Namespace).To(Equal(gcpNfsVolume.Namespace))
		})

		By("And Then KCP NfsIntance has GCP instance details", func() {
			Expect(kcpNfsInstance.Spec.Instance.Aws).To(BeNil())
			Expect(kcpNfsInstance.Spec.Instance.Azure).To(BeNil())
			Expect(kcpNfsInstance.Spec.Instance.OpenStack).To(BeNil())
			Expect(kcpNfsInstance.Spec.Instance.Gcp).NotTo(BeNil())

			Expect(kcpNfsInstance.Spec.Instance.Gcp.CapacityGb).To(Equal(gcpNfsVolume.Spec.CapacityGb))
			// if tier regional, then it's equal to region, otherwise it's a random zone from scope,
			// but starts with region name, and has -a, -b, -c
			Expect(strings.HasPrefix(kcpNfsInstance.Spec.Instance.Gcp.Location, scope.Spec.Region)).To(BeTrue())
			Expect(kcpNfsInstance.Spec.Instance.Gcp.Tier).To(Equal(cloudcontrolv1beta1.GcpFileTier(gcpNfsVolume.Spec.Tier)))
			Expect(kcpNfsInstance.Spec.Instance.Gcp.FileShareName).To(Equal(gcpNfsVolume.Spec.FileShareName))
		})

		By("When KCP NfsInstance has finalizer", func() {
			_, err := composed.PatchObjAddFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, kcpNfsInstance, infra.KCP().Client())
			Expect(err).To(Succeed())
		})

		By("And When KCP NfsInstance is ready", func() {
			kcpNfsInstance.Status.State = cloudcontrolv1beta1.StateReady
			meta.SetStatusCondition(&kcpNfsInstance.Status.Conditions, metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeReady,
				Message: cloudcontrolv1beta1.ConditionTypeReady,
			})
			qty, err := resource.ParseQuantity(fmt.Sprintf("%dGi", gcpNfsVolume.Spec.CapacityGb))
			Expect(err).To(Succeed())
			kcpNfsInstance.Status.Capacity = qty
			kcpNfsInstance.Status.Hosts = []string{"10.20.30.40"}
			kcpNfsInstance.Status.Host = "10.20.30.40"
			kcpNfsInstance.Status.Path = gcpNfsVolume.Spec.FileShareName
			kcpNfsInstance.Status.CapacityGb = gcpNfsVolume.Spec.CapacityGb

			err = composed.PatchObjStatus(infra.Ctx(), kcpNfsInstance, infra.KCP().Client())
			Expect(err).To(Succeed())
		})

		pv := &corev1.PersistentVolume{}

		By("Then SKR PersistentVolume is created", func() {
			pv.Name = gcpNfsVolume.Spec.PersistentVolume.Name
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), pv, NewObjActions()).
				Should(Succeed())
		})

		By("And Then SKR PersistentVolume has finalizer", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), pv, NewObjActions(), HavingFinalizer(api.CommonFinalizerDeletionHook)).
				Should(Succeed())
		})

		By("And Then SKR PersistentVolume has requested labels", func() {
			for k, v := range pvSpec.Labels {
				Expect(pv.Labels).To(HaveKeyWithValue(k, v))
			}
		})

		By("And Then SKR PersistentVolume has standard k8s labels", func() {
			Expect(pv.Labels).To(HaveKeyWithValue(util.WellKnownK8sLabelComponent, util.DefaultCloudManagerComponentLabelValue))
			Expect(pv.Labels).To(HaveKeyWithValue(util.WellKnownK8sLabelPartOf, util.DefaultCloudManagerPartOfLabelValue))
			Expect(pv.Labels).To(HaveKeyWithValue(util.WellKnownK8sLabelManagedBy, util.DefaultCloudManagerManagedByLabelValue))
		})

		By("And Then SKR PersistentVolume has CloudManager specific labels", func() {
			Expect(pv.Labels).To(HaveKeyWithValue(cloudresourcesv1beta1.LabelNfsVolName, gcpNfsVolume.Name))
			Expect(pv.Labels).To(HaveKeyWithValue(cloudresourcesv1beta1.LabelNfsVolNS, gcpNfsVolume.Namespace))
			Expect(pv.Labels).To(HaveKeyWithValue(cloudresourcesv1beta1.LabelCloudManaged, "true"))
		})

		By("And Then SKR PersistentVolume has requested annotations", func() {
			for k, v := range pvSpec.Annotations {
				Expect(pv.Annotations).To(HaveKeyWithValue(k, v))
			}
		})

		By("And Then SKR PersistentVolume has NFS", func() {
			Expect(pv.Spec.PersistentVolumeSource.NFS).ToNot(BeNil())
			Expect(pv.Spec.PersistentVolumeSource.NFS.Server).To(Equal(kcpNfsInstance.Status.Host))
			Expect(pv.Spec.PersistentVolumeSource.NFS.Path).To(Equal("/" + gcpNfsVolume.Spec.FileShareName))
		})

		pvc := &corev1.PersistentVolumeClaim{}

		By("And Then SKR PersistentVolumeClaim is created", func() {
			pvc.Name = gcpNfsVolume.Spec.PersistentVolumeClaim.Name
			pvc.Namespace = gcpNfsVolume.Namespace
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), pvc, NewObjActions()).
				Should(Succeed())
		})

		By("And Then SKR PersistentVolumeClaim has finalizer", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), pvc, NewObjActions(), HavingFinalizer(api.CommonFinalizerDeletionHook)).
				Should(Succeed())
		})

		By("And Then SKR PersistentVolumeClaim has requested labels", func() {
			for k, v := range pvcSpec.Labels {
				Expect(pvc.Labels).To(HaveKeyWithValue(k, v))
			}
		})

		By("And Then SKR PersistentVolumeClaim has standard k8s labels", func() {
			Expect(pvc.Labels).To(HaveKeyWithValue(util.WellKnownK8sLabelComponent, util.DefaultCloudManagerComponentLabelValue))
			Expect(pvc.Labels).To(HaveKeyWithValue(util.WellKnownK8sLabelPartOf, util.DefaultCloudManagerPartOfLabelValue))
			Expect(pvc.Labels).To(HaveKeyWithValue(util.WellKnownK8sLabelManagedBy, util.DefaultCloudManagerManagedByLabelValue))
		})

		By("And Then SKR PersistentVolumeClaim has CloudManager specific labels", func() {
			Expect(pvc.Labels).To(HaveKeyWithValue(cloudresourcesv1beta1.LabelNfsVolName, gcpNfsVolume.Name))
			Expect(pvc.Labels).To(HaveKeyWithValue(cloudresourcesv1beta1.LabelNfsVolNS, gcpNfsVolume.Namespace))
			Expect(pvc.Labels).To(HaveKeyWithValue(cloudresourcesv1beta1.LabelCloudManaged, "true"))
		})

		By("And Then SKR PersistentVolumeClaim references PV", func() {
			Expect(pvc.Spec.VolumeName).To(Equal(pv.Name))
		})

		By("And Then SKR PersistentVolumeClaim has ReadWriteMany access mode", func() {
			Expect(pvc.Spec.AccessModes).To(ContainElement(corev1.ReadWriteMany))
		})

		By("And Then SKR GcpNfsVolume is ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolume, NewObjActions(), HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady)).
				Should(Succeed())
		})

		// Delete

		By("When SKR GcpNfsVolume is deleted", func() {
			err := Delete(infra.Ctx(), infra.SKR().Client(), gcpNfsVolume)
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then SKR PersistentVolumeClaim does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), pvc).
				Should(Succeed())
		})

		By("And Then SKR PersistentVolume does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), pv).
				Should(Succeed())
		})

		By("And Then KCP NfsInstance is deleted", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNfsInstance, NewObjActions(), HavingDeletionTimestamp()).
				Should(Succeed())
		})

		By("When KCP NfsInstance finalizer is removed", func() {
			_, err := composed.PatchObjRemoveFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, kcpNfsInstance, infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then SKR GcpNfsVolume does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolume).
				Should(Succeed())
		})

		By("// cleanup: Delete SKR IpRange", func() {
			Expect(Delete(infra.Ctx(), infra.SKR().Client(), ipRange)).To(Succeed())
		})

		By("// cleanup: Delete KCP Scope", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), scope)).To(Succeed())
		})

	})

})
