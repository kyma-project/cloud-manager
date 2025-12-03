package cloudresources

import (
	"fmt"
	"time"

	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR GcpNfsVolume", func() {

	const (
		interval = time.Millisecond * 250
	)
	var (
		timeout = time.Second * 20
	)

	skrIpRangeName := "gcp-iprange-1-v"
	skrIpRange := &cloudresourcesv1beta1.IpRange{}
	kcpIpRangeName := "513f20b4-7b73-4246-9397-f8dd55344479"
	kcpIpRange := &cloudcontrolv1beta1.IpRange{}

	BeforeEach(func() {

		By("And Given SKR IPRange exists", func() {
			// tell skriprange reconciler to ignore this SKR IpRange
			skriprange.Ignore.AddName(skrIpRangeName)
			//Create SKR IPRange if it doesn't exist.
			Eventually(CreateSkrIpRange).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
					WithSkrIpRangeSpecCidr(addressSpace.MustAllocate(24)),
				).
				Should(Succeed())

			// tell kcpiprange reconciler to ignore this KCP IpRange
			kcpiprange.Ignore.AddName(kcpIpRangeName)

			//Create KCP IPRange if it doesn't exist.
			Eventually(CreateKcpIpRange).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithNamespace(DefaultKcpNamespace),
					WithScope(infra.SkrKymaRef().Name),
					WithLabels(map[string]string{
						cloudcontrolv1beta1.LabelKymaName:        infra.SkrKymaRef().Name,
						cloudcontrolv1beta1.LabelRemoteName:      skrIpRangeName,
						cloudcontrolv1beta1.LabelRemoteNamespace: DefaultSkrNamespace,
					}),
				).
				Should(Succeed())
		})
		By("And Given SKR IPRange in Ready state", func() {

			//Update SKR IpRange status to Ready
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())

			//Update KCP IpRange status to Ready
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusCidr(skrIpRange.Spec.Cidr),
					WithSkrIpRangeStatusId(kcpIpRangeName),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})
	})

	Describe("Scenario: SKR GcpNfsVolume Create", func() {
		//Define variables.
		gcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		pv := &corev1.PersistentVolume{}

		gcpNfsVolumeName := "gcp-nfs-volume-1"
		nfsIpAddress := "10.11.12.11"
		pvSpec := &cloudresourcesv1beta1.GcpNfsVolumePvSpec{
			Name: "gcp-nfs-pv-1",
			Labels: map[string]string{
				"app": "gcp-nfs",
			},
			Annotations: map[string]string{
				"volume": "gcp-nfs-volume-1",
			},
		}

		pvc := &corev1.PersistentVolumeClaim{}
		pvcSpec := &cloudresourcesv1beta1.GcpNfsVolumePvcSpec{
			Name: "gcp-nfs-pvc-1",
			Labels: map[string]string{
				"foo": "bar",
			},
			Annotations: map[string]string{
				"baz": "qux",
			},
		}

		It("When GcpNfsVolume Create is called", func() {
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
					WithName(gcpNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRange.Name),
					WithPvSpec(pvSpec),
					WithPvcSpec(pvcSpec),
				).
				Should(Succeed())

			By("Then GcpNfsVolume is created in SKR", func() {
				// load GcpNfsVolume to get ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
						AssertGcpNfsVolumeHasId(),
					).
					Should(Succeed())
			})

			By("Then NfsInstance is created in KCP", func() {

				// check KCP NfsInstance is created with name=gcpNfsVolume.ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())

				By("And has label cloud-manager.kyma-project.io/kymaName")
				Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))

				By("And has label cloud-manager.kyma-project.io/remoteName")
				Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(gcpNfsVolume.Name))

				By("And has label cloud-manager.kyma-project.io/remoteNamespace")
				Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(gcpNfsVolume.Namespace))

				By("And has spec.scope.name equal to SKR Cluster kyma name")
				Expect(kcpNfsInstance.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))

				By("And has spec.remoteRef matching to to SKR IpRange")
				Expect(kcpNfsInstance.Spec.RemoteRef.Namespace).To(Equal(gcpNfsVolume.Namespace))
				Expect(kcpNfsInstance.Spec.RemoteRef.Name).To(Equal(gcpNfsVolume.Name))
			})

			By("When KCP NfsInstance is switched to Ready condition", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())
				Eventually(UpdateStatus).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
						WithConditions(KcpReadyCondition()),
						WithKcpNfsStatusState(cloudcontrolv1beta1.StateReady),
						WithKcpNfsStatusHost(nfsIpAddress),
						WithKcpNfsStatusCapacity(gcpNfsVolume.Spec.CapacityGb),
					).
					Should(Succeed())
			})

			By("Then SKR GcpNfsVolume will get to Ready condition", func() {

				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
						HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					).
					Should(Succeed())

				By("And has the File store Host")
				Expect(gcpNfsVolume.Status.Hosts).To(HaveLen(1))

				By("And has the capacity equal to provisioned value.")
				Expect(gcpNfsVolume.Status.CapacityGb).To(Equal(gcpNfsVolume.Spec.CapacityGb))
				Expect(gcpNfsVolume.Status.Capacity).To(BeComparableTo(resource.MustParse(fmt.Sprintf("%dGi", gcpNfsVolume.Spec.CapacityGb))))

				By("And has empty Protocol")
				Expect(gcpNfsVolume.Status.Protocol).To(BeEmpty())
			})

			By("Then PersistentVolume is created in SKR", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						pv,
						NewObjActions(
							WithName(pvSpec.Name),
						),
					).
					Should(Succeed())

				By("And has the Server Name matching the provisioned File server host.")
				Expect(pv.Spec.PersistentVolumeSource.NFS.Server).To(Equal(gcpNfsVolume.Status.Hosts[0]))

				By("And has the Volume Name matching the requested FileShare name.")
				path := fmt.Sprintf("/%s", gcpNfsVolume.Spec.FileShareName)
				Expect(pv.Spec.PersistentVolumeSource.NFS.Path).To(Equal(path))

				By("And has the Capacity equal to requested value in GB.")
				expectedCapacity := gcpNfsVolume.Status.Capacity.Value()
				quantity := pv.Spec.Capacity["storage"]
				pvQuantity, _ := quantity.AsInt64()
				Expect(pvQuantity).To(Equal(expectedCapacity))

				By("And has defined cloud-manager default labels")
				Expect(pv.Labels[util.WellKnownK8sLabelComponent]).ToNot(BeNil())
				Expect(pv.Labels[util.WellKnownK8sLabelPartOf]).ToNot(BeNil())
				Expect(pv.Labels[util.WellKnownK8sLabelManagedBy]).ToNot(BeNil())

				By("And has the Labels matching the requested values in PvSpec.")
				Expect(pv.Labels).Should(HaveKeyWithValue("app", pvSpec.Labels["app"]))

				By("And has the Annotations matching the requested values in PvSpec.")
				Expect(pvSpec.Annotations).To(Equal(pv.Annotations))

				By("And it has defined cloud-manager finalizer")
				Expect(pv.Finalizers).To(ContainElement(api.CommonFinalizerDeletionHook))

				By("And has empty mount options")
				Expect(pv.Spec.MountOptions).To(BeEmpty())
			})

			By("Then PersistantVolumeClaim is created in SKR", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						pvc,
						NewObjActions(
							WithName(pvcSpec.Name),
							WithNamespace(gcpNfsVolume.Namespace),
						),
					).
					Should(Succeed())

				By("And its .spec.volumeName is PV name")
				Expect(pvc.Spec.VolumeName).To(Equal(pv.GetName()))

				By("And it has defined cloud-manager default labels")
				Expect(pv.Labels[util.WellKnownK8sLabelComponent]).ToNot(BeNil())
				Expect(pv.Labels[util.WellKnownK8sLabelPartOf]).ToNot(BeNil())
				Expect(pv.Labels[util.WellKnownK8sLabelManagedBy]).ToNot(BeNil())

				By("And it has defined custom label for capacity")
				storageCapacity := pv.Spec.Capacity["storage"]
				Expect(pvc.Labels[cloudresourcesv1beta1.LabelStorageCapacity]).To(Equal(storageCapacity.String()))

				By("And it has user defined custom labels")
				for k, v := range pvcSpec.Labels {
					Expect(pvc.Labels).To(HaveKeyWithValue(k, v), fmt.Sprintf("expected PVC to have label %s=%s", k, v))
				}
				By("And it has user defined custom annotations")
				for k, v := range pvcSpec.Annotations {
					Expect(pvc.Annotations).To(HaveKeyWithValue(k, v), fmt.Sprintf("expected PVC to have annotation %s=%s", k, v))
				}

				By("And it has defined cloud-manager finalizer")
				Expect(pv.Finalizers).To(ContainElement(api.CommonFinalizerDeletionHook))
			})
		})
	})

	Describe("Scenario: SKR GcpNfsVolume Update", func() {
		//Define variable
		gcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		pv := &corev1.PersistentVolume{}

		gcpNfsVolumeName := "gcp-nfs-volume-2"
		nfsIpAddress := "10.11.12.12"
		updatedCapacityGb := 1024

		prevPv := &corev1.PersistentVolume{}
		pvSpec := &cloudresourcesv1beta1.GcpNfsVolumePvSpec{
			Name: "gcp-nfs-pv-2",
			Labels: map[string]string{
				"app": "gcp-nfs-2",
			},
			Annotations: map[string]string{
				"volume": "gcp-nfs-volume-2",
			},
		}

		pvc := &corev1.PersistentVolumeClaim{}
		prevPvcSpec := &cloudresourcesv1beta1.GcpNfsVolumePvcSpec{
			Name: "gcp-nfs-pvc-2",
			Labels: map[string]string{
				"foo": "bar",
			},
			Annotations: map[string]string{
				"baz": "qux",
			},
		}
		pvcSpec := &cloudresourcesv1beta1.GcpNfsVolumePvcSpec{
			Name: "gcp-nfs-pvc-2",
			Labels: map[string]string{
				"foo":  "bar-changed",
				"foo2": "bar2",
			},
			Annotations: map[string]string{
				"baz":  "qux-changed",
				"baz2": "qux2",
			},
		}

		BeforeEach(func() {
			By("And Given SKR GcpNfsVolume exists", func() {
				//Create GcpNfsVolume
				Eventually(CreateGcpNfsVolume).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
						WithName(gcpNfsVolumeName),
						WithGcpNfsVolumeIpRange(skrIpRange.Name),
						WithPvcSpec(prevPvcSpec),
					).
					Should(Succeed())

				// Load GcpNfsVolume and check the Status.ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
						AssertGcpNfsVolumeHasId(),
					).
					Should(Succeed())

				// Load KCP NfsInstance is created with name=gcpNfsVolume.ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())

				//Update KCP NfsInstance to Ready state
				Eventually(UpdateStatus).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
						WithConditions(KcpReadyCondition()),
						WithKcpNfsStatusState(cloudcontrolv1beta1.StateReady),
						WithKcpNfsStatusHost(nfsIpAddress),
						WithKcpNfsStatusCapacity(gcpNfsVolume.Spec.CapacityGb),
					).
					Should(Succeed())

				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						prevPv,
						NewObjActions(
							WithName(gcpNfsVolume.Status.Id),
						),
					).
					Should(Succeed())

				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						pvc,
						NewObjActions(
							WithName(pvcSpec.Name),
							WithNamespace(gcpNfsVolume.Namespace),
						),
					).
					Should(Succeed())
			})
		})

		It("When SKR GcpNfsVolume Update is called ", func() {

			Eventually(Update).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
					WithGcpNfsVolumeCapacity(updatedCapacityGb),
					WithPvSpec(pvSpec),
					WithPvcSpec(pvcSpec),
				).
				Should(Succeed())

			By("Then GcpNfsVolume is updated with the new values.", func() {

				Eventually(LoadAndCheck, timeout, interval).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
						NewObjActions(),
					).
					Should(Succeed())

				Expect(gcpNfsVolume.Spec.CapacityGb).To(Equal(updatedCapacityGb))
			})

			By("And Then KCP NfsInstance too is updated", func() {
				Eventually(LoadAndCheck, timeout, interval).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())

				By("And has the CapacityGb matching that of the SKR GcpNfsVolume.")
				Expect(kcpNfsInstance.Spec.Instance.Gcp.CapacityGb).To(Equal(updatedCapacityGb))
			})

			By("When KCP NfsInstance Status is updated with new values.", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())

				Eventually(UpdateStatus).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
						WithConditions(KcpReadyCondition()),
						WithKcpNfsStatusState(cloudcontrolv1beta1.StateReady),
						WithKcpNfsStatusHost(nfsIpAddress),
						WithKcpNfsStatusCapacity(updatedCapacityGb),
					).
					Should(Succeed())
			})
			By("Then SKR GcpNfsVolume status is updated", func() {
				Eventually(LoadAndCheck, timeout, interval).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
					).
					Should(Succeed())

				By("And has status CapacityGb matching that of the SKR GcpNfsVolume Spec.")
				Expect(gcpNfsVolume.Status.CapacityGb).To(Equal(updatedCapacityGb))
				Expect(gcpNfsVolume.Status.Capacity).To(BeComparableTo(resource.MustParse(fmt.Sprintf("%dGi", updatedCapacityGb))))
			})

			By("And Then SKR PersistentVolume is updated", func() {
				Eventually(LoadAndCheck, timeout, interval).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						pv,
						NewObjActions(
							WithName(pvSpec.Name),
						),
					).
					Should(Succeed())

				By("And has the Capacity matching that of the SKR GcpNfsVolume.")
				expectedCapacity := int64(updatedCapacityGb) * 1024 * 1024 * 1024
				quantity := pv.Spec.Capacity["storage"]
				pvQuantity, _ := quantity.AsInt64()
				Expect(pvQuantity).To(Equal(expectedCapacity))

				By("And has the Labels matching the requested values in PvSpec.")
				Expect(pv.Labels).Should(HaveKeyWithValue("app", pvSpec.Labels["app"]))

				By("And has the Annotations matching the requested values in PvSpec.")
				Expect(pvSpec.Annotations).To(Equal(pv.Annotations))

				By("And Then previous PersistentVolume in SKR is deleted.", func() {
					Eventually(IsDeleted, timeout, interval).
						WithArguments(
							infra.Ctx(), infra.SKR().Client(), prevPv,
						).
						Should(Succeed())
				})
			})

			By("And Then SKR PersistentVolumeClaim is updated", func() {
				Eventually(LoadAndCheck, timeout, interval).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						pvc,
						NewObjActions(
							WithName(pvcSpec.Name),
							WithNamespace(gcpNfsVolume.Namespace),
						),
					).
					Should(Succeed())

				By("And it has defined cloud-manager default labels")
				Expect(pv.Labels[util.WellKnownK8sLabelComponent]).ToNot(BeNil())
				Expect(pv.Labels[util.WellKnownK8sLabelPartOf]).ToNot(BeNil())
				Expect(pv.Labels[util.WellKnownK8sLabelManagedBy]).ToNot(BeNil())

				By("And it has defined custom label for capacity")
				storageCapacity := pv.Spec.Capacity["storage"]
				Expect(pvc.Labels[cloudresourcesv1beta1.LabelStorageCapacity]).To(Equal(storageCapacity.String()))

				By("And it has user defined custom labels")
				for k, v := range pvcSpec.Labels {
					Expect(pvc.Labels).To(HaveKeyWithValue(k, v), fmt.Sprintf("expected PVC to have label %s=%s", k, v))
				}
				By("And it has user defined custom annotations")
				for k, v := range pvcSpec.Annotations {
					Expect(pvc.Annotations).To(HaveKeyWithValue(k, v), fmt.Sprintf("expected PVC to have annotation %s=%s", k, v))
				}
			})
		})
	})

	Describe("Scenario: SKR GcpNfsVolume Delete", func() {
		gcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		pv := &corev1.PersistentVolume{}
		pvc := &corev1.PersistentVolumeClaim{}

		gcpNfsVolumeName := "gcp-nfs-volume-3"
		nfsIpAddress := "10.11.12.13"

		BeforeEach(func() {
			By("And Given SKR GcpNfsVolume exists", func() {

				//Create GcpNfsVolume
				Eventually(CreateGcpNfsVolume).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
						WithName(gcpNfsVolumeName),
						WithGcpNfsVolumeIpRange(skrIpRange.Name),
					).
					Should(Succeed())

				// Load GcpNfsVolume and check the Status.ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
						AssertGcpNfsVolumeHasId(),
					).
					Should(Succeed())

				// Load KCP NfsInstance is created with name=gcpNfsVolume.ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())

				//Update KCP NfsInstance to Ready state
				Eventually(UpdateStatus).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
						WithConditions(KcpReadyCondition()),
						WithKcpNfsStatusState(cloudcontrolv1beta1.StateReady),
						WithKcpNfsStatusHost(nfsIpAddress),
						WithKcpNfsStatusCapacity(gcpNfsVolume.Spec.CapacityGb),
					).
					Should(Succeed())

				//Load PV
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), pv,
						NewObjActions(
							WithName(gcpNfsVolume.Status.Id),
						),
					).
					Should(Succeed())

				//Load PVC
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), pvc,
						NewObjActions(
							WithName(gcpNfsVolume.Name),
							WithNamespace(gcpNfsVolume.Namespace),
						),
					).
					Should(Succeed())
			})
		})
		It("When SKR GcpNfsVolume Delete is called ", func() {

			//Delete SKR GcpNfsVolume
			Eventually(Delete).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
				).Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
					NewObjActions(),
				).
				Should(Succeed())

			By("Then DeletionTimestamp is set in GcpNfsVolume", func() {
				Expect(gcpNfsVolume.DeletionTimestamp.IsZero()).NotTo(BeTrue())
			})

			By("And Then the PersistentVolumeClaim in SKR is deleted.", func() {
				Eventually(IsDeleted, timeout, interval).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), pvc,
					).
					Should(Succeed())
			})

			By("And Then the PersistentVolume in SKR is deleted.", func() {
				Eventually(IsDeleted, timeout, interval).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), pv,
					).
					Should(Succeed())
			})

			By("And Then the NfsInstance in KCP is deleted.", func() {
				Eventually(IsDeleted, timeout, interval).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
					).
					Should(Succeed())
			})

			By("And Then the GcpNfsVolume in SKR is deleted.", func() {
				Eventually(IsDeleted, timeout, interval).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
					).
					Should(Succeed())
			})
		})
	})

	Describe("Scenario: SKR GcpNfsVolume Create with empty location", func() {
		//Define variables.
		gcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		pv := &corev1.PersistentVolume{}
		scope := &cloudcontrolv1beta1.Scope{}
		gcpNfsVolumeName := "gcp-nfs-volume-4"
		nfsIpAddress := "10.11.12.14"
		pvSpec := &cloudresourcesv1beta1.GcpNfsVolumePvSpec{
			Name: "gcp-nfs-pv-4",
			Labels: map[string]string{
				"app": "gcp-nfs-4",
			},
			Annotations: map[string]string{
				"volume": "gcp-nfs-volume-4",
			},
		}

		pvc := &corev1.PersistentVolumeClaim{}
		pvcSpec := &cloudresourcesv1beta1.GcpNfsVolumePvcSpec{
			Name: "gcp-nfs-pvc-4",
			Labels: map[string]string{
				"foo": "bar",
			},
			Annotations: map[string]string{
				"baz": "qux",
			},
		}
		BeforeEach(func() {
			By("Given KCP Scope exists", func() {

				// Given Scope exists
				Expect(
					infra.GivenScopeGcpExists(infra.SkrKymaRef().Name),
				).NotTo(HaveOccurred())
				// Load created scope
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(infra.SkrKymaRef().Name), scope)
					exists = err == nil
					return exists, client.IgnoreNotFound(err)
				}, timeout, interval).
					Should(BeTrue(), "expected Scope to get created")
			})
		})

		It("When GcpNfsVolume Create is called", func() {
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
					WithName(gcpNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRange.Name),
					WithPvSpec(pvSpec),
					WithPvcSpec(pvcSpec),
				).
				Should(Succeed())

			By("Then GcpNfsVolume is created in SKR", func() {
				// load GcpNfsVolume to get ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
						AssertGcpNfsVolumeHasId(),
					).
					Should(Succeed())
			})

			By("Then NfsInstance is created in KCP", func() {

				// check KCP NfsInstance is created with name=gcpNfsVolume.ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())

				By("And has label cloud-manager.kyma-project.io/kymaName")
				Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))

				By("And has label cloud-manager.kyma-project.io/remoteName")
				Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(gcpNfsVolume.Name))

				By("And has label cloud-manager.kyma-project.io/remoteNamespace")
				Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(gcpNfsVolume.Namespace))

				By("And has spec.scope.name equal to SKR Cluster kyma name")
				Expect(kcpNfsInstance.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))

				By("And has spec.remoteRef matching to to SKR IpRange")
				Expect(kcpNfsInstance.Spec.RemoteRef.Namespace).To(Equal(gcpNfsVolume.Namespace))
				Expect(kcpNfsInstance.Spec.RemoteRef.Name).To(Equal(gcpNfsVolume.Name))
			})

			By("When KCP NfsInstance is switched to Ready condition", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())
				Eventually(UpdateStatus).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
						WithConditions(KcpReadyCondition()),
						WithKcpNfsStatusState(cloudcontrolv1beta1.StateReady),
						WithKcpNfsStatusHost(nfsIpAddress),
						WithKcpNfsStatusCapacity(gcpNfsVolume.Spec.CapacityGb),
					).
					Should(Succeed())
			})

			By("Then SKR GcpNfsVolume will get to Ready condition", func() {

				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
						HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					).
					Should(Succeed())

				By("And has the File store Host")
				Expect(gcpNfsVolume.Status.Hosts).To(HaveLen(1))

				By("And has the capacity equal to provisioned value.")
				Expect(gcpNfsVolume.Status.CapacityGb).To(Equal(gcpNfsVolume.Spec.CapacityGb))
				Expect(gcpNfsVolume.Status.Capacity).To(BeComparableTo(resource.MustParse(fmt.Sprintf("%dGi", gcpNfsVolume.Spec.CapacityGb))))
			})

			By("Then PersistentVolume is created in SKR", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						pv,
						NewObjActions(
							WithName(pvSpec.Name),
						),
					).
					Should(Succeed())

				By("And has the Server Name matching the provisioned File server host.")
				Expect(pv.Spec.PersistentVolumeSource.NFS.Server).To(Equal(gcpNfsVolume.Status.Hosts[0]))

				By("And has the Volume Name matching the requested FileShare name.")
				path := fmt.Sprintf("/%s", gcpNfsVolume.Spec.FileShareName)
				Expect(pv.Spec.PersistentVolumeSource.NFS.Path).To(Equal(path))

				By("And has the Capacity equal to requested value in GB.")
				expectedCapacity := gcpNfsVolume.Status.Capacity.Value()
				quantity := pv.Spec.Capacity["storage"]
				pvQuantity, _ := quantity.AsInt64()
				Expect(pvQuantity).To(Equal(expectedCapacity))

				By("And has defined cloud-manager default labels")
				Expect(pv.Labels[util.WellKnownK8sLabelComponent]).ToNot(BeNil())
				Expect(pv.Labels[util.WellKnownK8sLabelPartOf]).ToNot(BeNil())
				Expect(pv.Labels[util.WellKnownK8sLabelManagedBy]).ToNot(BeNil())

				By("And has the Labels matching the requested values in PvSpec.")
				Expect(pv.Labels).Should(HaveKeyWithValue("app", pvSpec.Labels["app"]))

				By("And has the Annotations matching the requested values in PvSpec.")
				Expect(pvSpec.Annotations).To(Equal(pv.Annotations))

				By("And it has defined cloud-manager finalizer")
				Expect(pv.Finalizers).To(ContainElement(api.CommonFinalizerDeletionHook))
			})

			By("Then PersistantVolumeClaim is created in SKR", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						pvc,
						NewObjActions(
							WithName(pvcSpec.Name),
							WithNamespace(gcpNfsVolume.Namespace),
						),
					).
					Should(Succeed())

				By("And its .spec.volumeName is PV name")
				Expect(pvc.Spec.VolumeName).To(Equal(pv.GetName()))

				By("And it has defined cloud-manager default labels")
				Expect(pv.Labels[util.WellKnownK8sLabelComponent]).ToNot(BeNil())
				Expect(pv.Labels[util.WellKnownK8sLabelPartOf]).ToNot(BeNil())
				Expect(pv.Labels[util.WellKnownK8sLabelManagedBy]).ToNot(BeNil())

				By("And it has defined custom label for capacity")
				storageCapacity := pv.Spec.Capacity["storage"]
				Expect(pvc.Labels[cloudresourcesv1beta1.LabelStorageCapacity]).To(Equal(storageCapacity.String()))

				By("And it has user defined custom labels")
				for k, v := range pvcSpec.Labels {
					Expect(pvc.Labels).To(HaveKeyWithValue(k, v), fmt.Sprintf("expected PVC to have label %s=%s", k, v))
				}
				By("And it has user defined custom annotations")
				for k, v := range pvcSpec.Annotations {
					Expect(pvc.Annotations).To(HaveKeyWithValue(k, v), fmt.Sprintf("expected PVC to have annotation %s=%s", k, v))
				}

				By("And it has defined cloud-manager finalizer")
				Expect(pv.Finalizers).To(ContainElement(api.CommonFinalizerDeletionHook))
			})
		})
	})

	Describe("Scenario: SKR GcpNfsVolume Update with empty location", func() {
		//Define variable
		gcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		pv := &corev1.PersistentVolume{}
		scope := &cloudcontrolv1beta1.Scope{}

		gcpNfsVolumeName := "gcp-nfs-volume-5"
		nfsIpAddress := "10.11.12.15"
		updatedCapacityGb := 1024

		prevPv := &corev1.PersistentVolume{}
		pvSpec := &cloudresourcesv1beta1.GcpNfsVolumePvSpec{
			Name: "gcp-nfs-pv-5",
			Labels: map[string]string{
				"app": "gcp-nfs-5",
			},
			Annotations: map[string]string{
				"volume": "gcp-nfs-volume-5",
			},
		}

		pvc := &corev1.PersistentVolumeClaim{}
		prevPvcSpec := &cloudresourcesv1beta1.GcpNfsVolumePvcSpec{
			Name: "gcp-nfs-pvc-5",
			Labels: map[string]string{
				"foo": "bar",
			},
			Annotations: map[string]string{
				"baz": "qux",
			},
		}
		pvcSpec := &cloudresourcesv1beta1.GcpNfsVolumePvcSpec{
			Name: "gcp-nfs-pvc-5",
			Labels: map[string]string{
				"foo":  "bar-changed",
				"foo2": "bar2",
			},
			Annotations: map[string]string{
				"baz":  "qux-changed",
				"baz2": "qux2",
			},
		}

		BeforeEach(func() {
			By("Given KCP Scope exists", func() {

				// Given Scope exists
				Expect(
					infra.GivenScopeGcpExists(infra.SkrKymaRef().Name),
				).NotTo(HaveOccurred())
				// Load created scope
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(infra.SkrKymaRef().Name), scope)
					exists = err == nil
					return exists, client.IgnoreNotFound(err)
				}, timeout, interval).
					Should(BeTrue(), "expected Scope to get created")
			})
			By("And Given SKR GcpNfsVolume exists", func() {
				//Create GcpNfsVolume
				Eventually(CreateGcpNfsVolume).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
						WithName(gcpNfsVolumeName),
						WithGcpNfsVolumeIpRange(skrIpRange.Name),
						WithPvcSpec(prevPvcSpec),
					).
					Should(Succeed())

				// Load GcpNfsVolume and check the Status.ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
						AssertGcpNfsVolumeHasId(),
					).
					Should(Succeed())

				// Load KCP NfsInstance is created with name=gcpNfsVolume.ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())

				//Update KCP NfsInstance to Ready state
				Eventually(UpdateStatus).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
						WithConditions(KcpReadyCondition()),
						WithKcpNfsStatusState(cloudcontrolv1beta1.StateReady),
						WithKcpNfsStatusHost(nfsIpAddress),
						WithKcpNfsStatusCapacity(gcpNfsVolume.Spec.CapacityGb),
					).
					Should(Succeed())

				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						prevPv,
						NewObjActions(
							WithName(gcpNfsVolume.Status.Id),
						),
					).
					Should(Succeed())

				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						pvc,
						NewObjActions(
							WithName(pvcSpec.Name),
							WithNamespace(gcpNfsVolume.Namespace),
						),
					).
					Should(Succeed())
			})
		})

		It("When SKR GcpNfsVolume Update is called ", func() {
			Eventually(Update).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
					WithGcpNfsVolumeCapacity(updatedCapacityGb),
					WithPvSpec(pvSpec),
					WithPvcSpec(pvcSpec),
				).
				Should(Succeed())

			By("Then GcpNfsVolume is updated with the new values.", func() {

				Eventually(LoadAndCheck, timeout, interval).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
						NewObjActions(),
					).
					Should(Succeed())

				Expect(gcpNfsVolume.Spec.CapacityGb).To(Equal(updatedCapacityGb))
			})

			By("And Then KCP NfsInstance too is updated", func() {
				Eventually(LoadAndCheck, timeout, interval).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())

				By("And has the CapacityGb matching that of the SKR GcpNfsVolume.")
				Expect(kcpNfsInstance.Spec.Instance.Gcp.CapacityGb).To(Equal(updatedCapacityGb))
			})

			By("When KCP NfsInstance Status is updated with new values.", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())

				Eventually(UpdateStatus).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
						WithConditions(KcpReadyCondition()),
						WithKcpNfsStatusState(cloudcontrolv1beta1.StateReady),
						WithKcpNfsStatusHost(nfsIpAddress),
						WithKcpNfsStatusCapacity(updatedCapacityGb),
					).
					Should(Succeed())
			})
			By("Then SKR GcpNfsVolume status is updated", func() {
				Eventually(LoadAndCheck, timeout, interval).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
					).
					Should(Succeed())

				By("And has status CapacityGb matching that of the SKR GcpNfsVolume Spec.")
				Expect(gcpNfsVolume.Status.CapacityGb).To(Equal(updatedCapacityGb))
				Expect(gcpNfsVolume.Status.Capacity).To(BeComparableTo(resource.MustParse(fmt.Sprintf("%dGi", updatedCapacityGb))))
			})

			By("And Then SKR PersistentVolume is updated", func() {
				Eventually(LoadAndCheck, timeout, interval).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						pv,
						NewObjActions(
							WithName(pvSpec.Name),
						),
					).
					Should(Succeed())

				By("And has the Capacity matching that of the SKR GcpNfsVolume.")
				expectedCapacity := int64(updatedCapacityGb) * 1024 * 1024 * 1024
				quantity := pv.Spec.Capacity["storage"]
				pvQuantity, _ := quantity.AsInt64()
				Expect(pvQuantity).To(Equal(expectedCapacity))

				By("And has the Labels matching the requested values in PvSpec.")
				Expect(pv.Labels).Should(HaveKeyWithValue("app", pvSpec.Labels["app"]))

				By("And has the Annotations matching the requested values in PvSpec.")
				Expect(pvSpec.Annotations).To(Equal(pv.Annotations))

				By("And Then previous PersistentVolume in SKR is deleted.", func() {
					Eventually(IsDeleted, timeout, interval).
						WithArguments(
							infra.Ctx(), infra.SKR().Client(), prevPv,
						).
						Should(Succeed())
				})
			})

			By("And Then SKR PersistentVolumeClaim is updated", func() {
				Eventually(LoadAndCheck, timeout, interval).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						pvc,
						NewObjActions(
							WithName(pvcSpec.Name),
							WithNamespace(gcpNfsVolume.Namespace),
						),
					).
					Should(Succeed())

				By("And it has defined cloud-manager default labels")
				Expect(pv.Labels[util.WellKnownK8sLabelComponent]).ToNot(BeNil())
				Expect(pv.Labels[util.WellKnownK8sLabelPartOf]).ToNot(BeNil())
				Expect(pv.Labels[util.WellKnownK8sLabelManagedBy]).ToNot(BeNil())

				By("And it has defined custom label for capacity")
				storageCapacity := pv.Spec.Capacity["storage"]
				Expect(pvc.Labels[cloudresourcesv1beta1.LabelStorageCapacity]).To(Equal(storageCapacity.String()))

				By("And it has user defined custom labels")
				for k, v := range pvcSpec.Labels {
					Expect(pvc.Labels).To(HaveKeyWithValue(k, v), fmt.Sprintf("expected PVC to have label %s=%s", k, v))
				}
				By("And it has user defined custom annotations")
				for k, v := range pvcSpec.Annotations {
					Expect(pvc.Annotations).To(HaveKeyWithValue(k, v), fmt.Sprintf("expected PVC to have annotation %s=%s", k, v))
				}
			})
		})
	})

	Describe("Scenario: SKR GcpNfsVolume Delete with empty location", func() {
		gcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		pv := &corev1.PersistentVolume{}
		pvc := &corev1.PersistentVolumeClaim{}
		scope := &cloudcontrolv1beta1.Scope{}

		gcpNfsVolumeName := "gcp-nfs-volume-6"
		nfsIpAddress := "10.11.12.16"

		BeforeEach(func() {
			By("Given KCP Scope exists", func() {

				// Given Scope exists
				Expect(
					infra.GivenScopeGcpExists(infra.SkrKymaRef().Name),
				).NotTo(HaveOccurred())
				// Load created scope
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(infra.SkrKymaRef().Name), scope)
					exists = err == nil
					return exists, client.IgnoreNotFound(err)
				}, timeout, interval).
					Should(BeTrue(), "expected Scope to get created")
			})

			By("And Given SKR GcpNfsVolume exists", func() {

				//Create GcpNfsVolume
				Eventually(CreateGcpNfsVolume).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
						WithName(gcpNfsVolumeName),
						WithGcpNfsVolumeIpRange(skrIpRange.Name),
					).
					Should(Succeed())

				// Load GcpNfsVolume and check the Status.ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
						AssertGcpNfsVolumeHasId(),
					).
					Should(Succeed())

				// Load KCP NfsInstance is created with name=gcpNfsVolume.ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())

				//Update KCP NfsInstance to Ready state
				Eventually(UpdateStatus).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
						WithConditions(KcpReadyCondition()),
						WithKcpNfsStatusState(cloudcontrolv1beta1.StateReady),
						WithKcpNfsStatusHost(nfsIpAddress),
						WithKcpNfsStatusCapacity(gcpNfsVolume.Spec.CapacityGb),
					).
					Should(Succeed())

				//Load PV
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), pv,
						NewObjActions(
							WithName(gcpNfsVolume.Status.Id),
						),
					).
					Should(Succeed())

				//Load PVC
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), pvc,
						NewObjActions(
							WithName(gcpNfsVolume.Name),
							WithNamespace(gcpNfsVolume.Namespace),
						),
					).
					Should(Succeed())
			})
		})
		It("When SKR GcpNfsVolume Delete is called ", func() {
			//Delete SKR GcpNfsVolume
			Eventually(Delete).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
				).Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
					NewObjActions(),
				).
				Should(Succeed())

			By("Then DeletionTimestamp is set in GcpNfsVolume", func() {
				Expect(gcpNfsVolume.DeletionTimestamp.IsZero()).NotTo(BeTrue())
			})

			By("And Then the PersistentVolumeClaim in SKR is deleted.", func() {
				Eventually(IsDeleted, timeout, interval).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), pvc,
					).
					Should(Succeed())
			})

			By("And Then the PersistentVolume in SKR is deleted.", func() {
				Eventually(IsDeleted, timeout, interval).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), pv,
					).
					Should(Succeed())
			})

			By("And Then the NfsInstance in KCP is deleted.", func() {
				Eventually(IsDeleted, timeout, interval).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
					).
					Should(Succeed())
			})

			By("And Then the GcpNfsVolume in SKR is deleted.", func() {
				Eventually(IsDeleted, timeout, interval).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
					).
					Should(Succeed())
			})
		})
	})

	Describe("Scenario: SKR GcpNfsVolume Create with Zonal tier", func() {
		//Define variables.
		gcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		pv := &corev1.PersistentVolume{}

		gcpNfsVolumeName := "gcp-nfs-volume-7"
		nfsIpAddress := "10.11.12.17"
		pvSpec := &cloudresourcesv1beta1.GcpNfsVolumePvSpec{
			Name: "gcp-nfs-pv-7",
			Labels: map[string]string{
				"app": "gcp-nfs",
			},
			Annotations: map[string]string{
				"volume": "gcp-nfs-volume-7",
			},
		}

		pvc := &corev1.PersistentVolumeClaim{}
		pvcSpec := &cloudresourcesv1beta1.GcpNfsVolumePvcSpec{
			Name: "gcp-nfs-pvc-7",
			Labels: map[string]string{
				"foo": "bar",
			},
			Annotations: map[string]string{
				"baz": "qux",
			},
		}

		It("When GcpNfsVolume Create is called", func() {
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
					WithName(gcpNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRange.Name),
					WithPvSpec(pvSpec),
					WithPvcSpec(pvcSpec),
					WithGcpNfsVolumeSpecTier(cloudresourcesv1beta1.ZONAL),
				).
				Should(Succeed())

			By("Then GcpNfsVolume is created in SKR", func() {
				// load GcpNfsVolume to get ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
						AssertGcpNfsVolumeHasId(),
					).
					Should(Succeed())
			})

			By("Then NfsInstance is created in KCP", func() {

				// check KCP NfsInstance is created with name=gcpNfsVolume.ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())

				By("And has label cloud-manager.kyma-project.io/kymaName")
				Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))

				By("And has label cloud-manager.kyma-project.io/remoteName")
				Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(gcpNfsVolume.Name))

				By("And has label cloud-manager.kyma-project.io/remoteNamespace")
				Expect(kcpNfsInstance.Labels[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(gcpNfsVolume.Namespace))

				By("And has spec.scope.name equal to SKR Cluster kyma name")
				Expect(kcpNfsInstance.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))

				By("And has spec.remoteRef matching to to SKR IpRange")
				Expect(kcpNfsInstance.Spec.RemoteRef.Namespace).To(Equal(gcpNfsVolume.Namespace))
				Expect(kcpNfsInstance.Spec.RemoteRef.Name).To(Equal(gcpNfsVolume.Name))
			})

			By("When KCP NfsInstance is switched to Ready condition", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())
				Eventually(UpdateStatus).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
						WithConditions(KcpReadyCondition()),
						WithKcpNfsStatusState(cloudcontrolv1beta1.StateReady),
						WithKcpNfsStatusHost(nfsIpAddress),
						WithKcpNfsStatusCapacity(gcpNfsVolume.Spec.CapacityGb),
						WithKcpNfsStatusStateData(gcpclient.GcpNfsStateDataProtocol, string(gcpclient.FilestoreProtocolNFSv41)),
					).
					Should(Succeed())
			})

			By("Then SKR GcpNfsVolume will get to Ready condition", func() {
				//amir
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
						HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					).
					Should(Succeed())

				By("And has the File store Host")
				Expect(gcpNfsVolume.Status.Hosts).To(HaveLen(1))

				By("And has the capacity equal to provisioned value.")
				Expect(gcpNfsVolume.Status.CapacityGb).To(Equal(gcpNfsVolume.Spec.CapacityGb))

				By("And has the protocol equal to NFSv4.1")
				Expect(gcpNfsVolume.Status.Protocol).To(Equal(string(gcpclient.FilestoreProtocolNFSv41)))
			})

			By("Then PersistentVolume is created in SKR", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						pv,
						NewObjActions(
							WithName(pvSpec.Name),
						),
					).
					Should(Succeed())

				By("And has the Server Name matching the provisioned File server host.")
				Expect(pv.Spec.PersistentVolumeSource.NFS.Server).To(Equal(gcpNfsVolume.Status.Hosts[0]))

				By("And has the Volume Name matching the requested FileShare name.")
				path := fmt.Sprintf("/%s", gcpNfsVolume.Spec.FileShareName)
				Expect(pv.Spec.PersistentVolumeSource.NFS.Path).To(Equal(path))

				By("And has the Capacity equal to requested value in GB.")
				expectedCapacity := int64(gcpNfsVolume.Status.CapacityGb) * 1024 * 1024 * 1024
				quantity := pv.Spec.Capacity["storage"]
				pvQuantity, _ := quantity.AsInt64()
				Expect(pvQuantity).To(Equal(expectedCapacity))

				By("And has defined cloud-manager default labels")
				Expect(pv.Labels[util.WellKnownK8sLabelComponent]).ToNot(BeNil())
				Expect(pv.Labels[util.WellKnownK8sLabelPartOf]).ToNot(BeNil())
				Expect(pv.Labels[util.WellKnownK8sLabelManagedBy]).ToNot(BeNil())

				By("And has the Labels matching the requested values in PvSpec.")
				Expect(pv.Labels).Should(HaveKeyWithValue("app", pvSpec.Labels["app"]))

				By("And has the Annotations matching the requested values in PvSpec.")
				Expect(pvSpec.Annotations).To(Equal(pv.Annotations))

				By("And it has defined cloud-manager finalizer")
				Expect(pv.Finalizers).To(ContainElement(api.CommonFinalizerDeletionHook))

				By("And has mount options for NFSv4.1")
				Expect(pv.Spec.MountOptions).To(ContainElement("nfsvers=4.1"))
			})

			By("Then PersistantVolumeClaim is created in SKR", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						pvc,
						NewObjActions(
							WithName(pvcSpec.Name),
							WithNamespace(gcpNfsVolume.Namespace),
						),
					).
					Should(Succeed())

				By("And its .spec.volumeName is PV name")
				Expect(pvc.Spec.VolumeName).To(Equal(pv.GetName()))

				By("And it has defined cloud-manager default labels")
				Expect(pv.Labels[util.WellKnownK8sLabelComponent]).ToNot(BeNil())
				Expect(pv.Labels[util.WellKnownK8sLabelPartOf]).ToNot(BeNil())
				Expect(pv.Labels[util.WellKnownK8sLabelManagedBy]).ToNot(BeNil())

				By("And it has defined custom label for capacity")
				storageCapacity := pv.Spec.Capacity["storage"]
				Expect(pvc.Labels[cloudresourcesv1beta1.LabelStorageCapacity]).To(Equal(storageCapacity.String()))

				By("And it has user defined custom labels")
				for k, v := range pvcSpec.Labels {
					Expect(pvc.Labels).To(HaveKeyWithValue(k, v), fmt.Sprintf("expected PVC to have label %s=%s", k, v))
				}
				By("And it has user defined custom annotations")
				for k, v := range pvcSpec.Annotations {
					Expect(pvc.Annotations).To(HaveKeyWithValue(k, v), fmt.Sprintf("expected PVC to have annotation %s=%s", k, v))
				}

				By("And it has defined cloud-manager finalizer")
				Expect(pv.Finalizers).To(ContainElement(api.CommonFinalizerDeletionHook))
			})
		})
	})

	Describe("Scenario: SKR GcpNfsVolume Create with REGIONAL tier and empty location", func() {
		//Define variables.
		gcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		pv := &corev1.PersistentVolume{}
		scope := &cloudcontrolv1beta1.Scope{}

		gcpNfsVolumeName := "gcp-nfs-volume-8"
		nfsIpAddress := "10.11.12.18"
		pvSpec := &cloudresourcesv1beta1.GcpNfsVolumePvSpec{
			Name: "gcp-nfs-pv-8",
			Labels: map[string]string{
				"app": "gcp-nfs-regional",
			},
			Annotations: map[string]string{
				"volume": "gcp-nfs-volume-8",
			},
		}

		pvc := &corev1.PersistentVolumeClaim{}
		pvcSpec := &cloudresourcesv1beta1.GcpNfsVolumePvcSpec{
			Name: "gcp-nfs-pvc-8",
			Labels: map[string]string{
				"foo": "bar",
			},
			Annotations: map[string]string{
				"baz": "qux",
			},
		}

		BeforeEach(func() {
			By("Given KCP Scope exists", func() {
				// Given Scope exists
				Expect(
					infra.GivenScopeGcpExists(infra.SkrKymaRef().Name),
				).NotTo(HaveOccurred())
				// Load created scope
				Eventually(func() (exists bool, err error) {
					err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(infra.SkrKymaRef().Name), scope)
					exists = err == nil
					return exists, client.IgnoreNotFound(err)
				}, timeout, interval).
					Should(BeTrue(), "expected Scope to get created")
			})
		})

		It("When GcpNfsVolume with REGIONAL tier is created without location", func() {
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolume,
					WithName(gcpNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRange.Name),
					WithGcpNfsVolumeSpecTier(cloudresourcesv1beta1.REGIONAL),
					WithPvSpec(pvSpec),
					WithPvcSpec(pvcSpec),
				).
				Should(Succeed())

			By("Then GcpNfsVolume is created in SKR", func() {
				// load GcpNfsVolume to get ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
						AssertGcpNfsVolumeHasId(),
					).
					Should(Succeed())
			})

			By("Then NfsInstance is created in KCP with region as location", func() {
				// check KCP NfsInstance is created with name=gcpNfsVolume.ID
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())

				By("And has location set to Scope region")
				Expect(kcpNfsInstance.Spec.Instance.Gcp.Location).To(Equal(scope.Spec.Region),
					"Expected REGIONAL tier to use Scope region as location")

				By("And has REGIONAL tier")
				Expect(kcpNfsInstance.Spec.Instance.Gcp.Tier).To(Equal(cloudcontrolv1beta1.REGIONAL))
			})

			By("And Status.Location is set in SKR GcpNfsVolume", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
					).
					Should(Succeed())

				By("And Status.Location matches the Scope region")
				Expect(gcpNfsVolume.Status.Location).To(Equal(scope.Spec.Region),
					"Expected Status.Location to be set to Scope region")
			})

			By("When KCP NfsInstance is switched to Ready condition", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.KCP().Client(),
						kcpNfsInstance,
						NewObjActions(WithName(gcpNfsVolume.Status.Id)),
					).
					Should(Succeed())
				Eventually(UpdateStatus).
					WithArguments(
						infra.Ctx(), infra.KCP().Client(), kcpNfsInstance,
						WithConditions(KcpReadyCondition()),
						WithKcpNfsStatusState(cloudcontrolv1beta1.StateReady),
						WithKcpNfsStatusHost(nfsIpAddress),
						WithKcpNfsStatusCapacity(gcpNfsVolume.Spec.CapacityGb),
					).
					Should(Succeed())
			})

			By("Then SKR GcpNfsVolume will get to Ready condition", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						gcpNfsVolume,
						NewObjActions(),
						HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					).
					Should(Succeed())
			})

			By("Then PersistentVolume is created", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						pv,
						NewObjActions(
							WithName(pvSpec.Name),
						),
					).
					Should(Succeed())
			})

			By("Then PersistentVolumeClaim is created", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						pvc,
						NewObjActions(
							WithName(pvcSpec.Name),
							WithNamespace(gcpNfsVolume.Namespace),
						),
					).
					Should(Succeed())
			})
		})
	})

})
