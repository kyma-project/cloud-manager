/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloudresources

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Feature: SKR CceeNfsVolume", func() {

	newRandomMapStringString := func() map[string]string {
		return map[string]string{
			uuid.NewString(): uuid.NewString()[0:8],
		}
	}

	It("Scenario: SKR CceeNfsVolume is created with empty IpRange when default IpRange does not exist", func() {
		cceeNfsVolumeName := "d2859451-39ed-4cc5-bf6d-d04aa8feeb5b"
		skrIpRangeId := "cddae623-d665-4dae-9899-515d1c2e1418"
		cceeNfsVolume := &cloudresourcesv1beta1.CceeNfsVolume{}
		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		capacityGb := 100

		pv := &corev1.PersistentVolume{}
		pvc := &corev1.PersistentVolumeClaim{}

		pvLabels := newRandomMapStringString()
		pvAnnotations := newRandomMapStringString()
		pvcLabels := newRandomMapStringString()
		pvcAnnotations := newRandomMapStringString()

		skriprange.Ignore.AddName("default")

		By("Given default SKR IpRange does not exist", func() {
			Consistently(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				ShouldNot(Succeed())
		})

		By("When CceeNfsVolume is created with empty IpRange", func() {
			Eventually(CreateCceeNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), cceeNfsVolume,
					WithName(cceeNfsVolumeName),
					WithCceeNfsVolumeCapacity(capacityGb),
					WithCceeNfsVolumePvLabels(pvLabels),
					WithCceeNfsVolumePvAnnotations(pvAnnotations),
					WithCceeNfsVolumePvcLabels(pvcLabels),
					WithCceeNfsVolumePvcAnnotations(pvcAnnotations),
				).
				Should(Succeed())
		})

		By("Then default SKR IpRange is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("And Then CceeNfsVolume is not ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), cceeNfsVolume, NewObjActions()).
				Should(Succeed())
			Expect(meta.IsStatusConditionTrue(cceeNfsVolume.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)).
				To(BeFalse(), "expected CceeNfsVolume not to have Ready condition, but it has")
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
			// load SKR CceeNfsVolume to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					cceeNfsVolume,
					NewObjActions(),
					HavingCceeNfsVolumeStatusId(),
					HavingCceeNfsVolumeStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR CceeNfsVolume to get status.id and status creating")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpNfsInstance,
					NewObjActions(
						WithName(cceeNfsVolume.Status.Id),
					),
				).
				Should(Succeed())

			Eventually(Update).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNfsInstance, AddFinalizer(api.CommonFinalizerDeletionHook)).
				Should(Succeed(), "failed adding finalizer on KCP NfsInstance")
		})

		By("When KCP NfsInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpNfsInstance,
					WithNfsInstanceStatusHost(""),
					WithNfsInstanceStatusPath(""),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR PersistentVolume is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					pv,
					NewObjActions(WithName(cceeNfsVolume.Status.Id)),
				).
				Should(Succeed())
		})

		By("And Then SKR PersistentVolumeClaim is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					pvc,
					NewObjActions(
						WithName(cceeNfsVolume.Name),
						WithNamespace(cceeNfsVolume.Namespace),
					),
				).
				Should(Succeed())
		})

		By("And Then SKR CceeNfsVolume has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					cceeNfsVolume,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingCceeNfsVolumeStatusState(cloudresourcesv1beta1.StateReady),
				).
				Should(Succeed())
		})

		// PV assertions ===============================================================

		By("And Then SKR PersistentVolume has storage capacity equal to CceeNfsVolume capacity", func() {
			expected := resource.MustParse(fmt.Sprintf("%dG", cceeNfsVolume.Spec.CapacityGb))
			Expect(expected.Equal(pv.Spec.Capacity["storage"])).To(BeTrue())
		})

		By("And Then SKR PersistentVolume ReadWriteMany access", func() {
			Expect(pv.Spec.AccessModes).To(HaveLen(1))
			Expect(pv.Spec.AccessModes[0]).To(Equal(corev1.ReadWriteMany))
		})

		By("And Then SKR PersistentVolume has well-known CloudManager labels", func() {
			Expect(pv.Labels[util.WellKnownK8sLabelComponent]).To(Equal(util.DefaultCloudManagerComponentLabelValue))
			Expect(pv.Labels[util.WellKnownK8sLabelPartOf]).To(Equal(util.DefaultCloudManagerPartOfLabelValue))
			Expect(pv.Labels[util.WellKnownK8sLabelManagedBy]).To(Equal(util.DefaultCloudManagerManagedByLabelValue))
		})

		By("And Then SKR PersistentVolume has parent NFS volume label", func() {
			Expect(pv.Labels[cloudresourcesv1beta1.LabelNfsVolName]).To(Equal(cceeNfsVolume.Name))
			Expect(pv.Labels[cloudresourcesv1beta1.LabelNfsVolNS]).To(Equal(cceeNfsVolume.Namespace))
		})

		By("And Then SKR PersistentVolume has user defined labels", func() {
			for k, v := range pvLabels {
				Expect(pv.Labels[k]).To(Equal(v))
			}
		})

		By("And Then SKR PersistentVolume has user defined annotations", func() {
			for k, v := range pvAnnotations {
				Expect(pv.Annotations[k]).To(Equal(v))
			}
		})

		By("And Then SKR PersistentVolume has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(pv, api.CommonFinalizerDeletionHook)).To(BeTrue())
		})

		By("And Then SKR PersistentVolume has NFS host and path equal to KCP NfsInstance values", func() {
			Expect(pv.Spec.PersistentVolumeSource.NFS).NotTo(BeNil())
			Expect(pv.Spec.PersistentVolumeSource.NFS.Server).To(Equal(kcpNfsInstance.Status.Host))
			Expect(pv.Spec.PersistentVolumeSource.NFS.Path).To(Equal(kcpNfsInstance.Status.Path))
		})

		// PVC assertions ===============================================================

		By("And Then SKR PersistentVolumeClaim has well-known CloudManager labels", func() {
			Expect(pvc.Labels[util.WellKnownK8sLabelComponent]).To(Equal(util.DefaultCloudManagerComponentLabelValue))
			Expect(pvc.Labels[util.WellKnownK8sLabelPartOf]).To(Equal(util.DefaultCloudManagerPartOfLabelValue))
			Expect(pvc.Labels[util.WellKnownK8sLabelManagedBy]).To(Equal(util.DefaultCloudManagerManagedByLabelValue))
		})

		By("And Then SKR PersistentVolumeClaim has user defined labels", func() {
			for k, v := range pvcLabels {
				Expect(pvc.Labels[k]).To(Equal(v))
			}
		})

		By("And Then SKR PersistentVolumeClaim has user defined annotations", func() {
			for k, v := range pvcAnnotations {
				Expect(pvc.Annotations[k]).To(Equal(v))
			}
		})

		By("And Then SKR PersistentVolumeClaim has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(pvc, api.CommonFinalizerDeletionHook)).To(BeTrue())
		})

		By("And Then SKR PersistentVolumeClaim references PersistentVolume", func() {
			Expect(pvc.Spec.VolumeName).To(Equal(pv.Name))
		})

		// DELETE ===============================================================

		By("When CceeNfsVolume is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), cceeNfsVolume).
				Should(Succeed(), "failed deleting CceeNfsVolume")
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
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpNfsInstance, RemoveFinalizer(api.CommonFinalizerDeletionHook)).
				Should(Succeed(), "failed removing finalizer on KCP NfsInstance")
		})

		By("Then SKR CceeNfsVolume is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), cceeNfsVolume).
				Should(Succeed(), "expected CceeNfsVolume not to exist")
		})

		By("// cleanup SKR IpRange", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
		})
	})
})
